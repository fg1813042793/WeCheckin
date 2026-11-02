package exam

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func (s *Service) RecordGradingContext(ctx context.Context, examID, recordID, adminID uint) (*RecordGradingDetail, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	record, exam, snapshot, notice, err := loadRecordGradingContext(ctx, db, examID, recordID, adminID, false)
	if err != nil {
		return nil, err
	}
	return buildRecordGradingDetail(record, exam, snapshot, notice)
}

func (s *Service) GradeRecordContext(ctx context.Context, examID, recordID, adminID uint, request GradeRecordRequest) (*RecordGradingDetail, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var result *RecordGradingDetail
	err := db.Transaction(func(tx *gorm.DB) error {
		record, exam, snapshot, notice, err := loadRecordGradingContext(ctx, tx, examID, recordID, adminID, true)
		if err != nil {
			return err
		}
		detail, err := buildRecordGradingDetail(record, exam, snapshot, notice)
		if err != nil {
			return err
		}
		updated, err := applyManualGrading(record, detail, request, adminID, database.Now())
		if err != nil {
			return err
		}
		change := tx.Model(&model.ExamRecord{}).Where("exam_r_id = ? AND exam_r_exam_id = ? AND exam_r_status = 1", recordID, examID).Updates(map[string]interface{}{
			"exam_r_result":          updated.Result,
			"exam_r_score":           updated.Score,
			"exam_r_total_score":     updated.TotalScore,
			"exam_r_pass":            updated.Pass,
			"exam_r_status":          updated.Status,
			"exam_r_graded_by":       updated.GradedBy,
			"exam_r_graded_at":       updated.GradedAt,
			"exam_r_grading_comment": updated.GradingComment,
		})
		if change.Error != nil {
			return change.Error
		}
		if change.RowsAffected != 1 {
			return ErrRecordGradingChanged
		}
		result, err = buildRecordGradingDetail(updated, exam, snapshot, notice)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func loadRecordGradingContext(ctx context.Context, db *gorm.DB, examID, recordID, adminID uint, lock bool) (model.ExamRecord, model.Exam, recordReviewSnapshot, string, error) {
	var record model.ExamRecord
	var exam model.Exam
	var snapshot recordReviewSnapshot
	if examID == 0 || recordID == 0 || adminID == 0 {
		return record, exam, snapshot, "", gorm.ErrRecordNotFound
	}
	examQuery, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return record, exam, snapshot, "", err
	}
	if err := examQuery.Where("exam_id = ?", examID).First(&exam).Error; err != nil {
		return record, exam, snapshot, "", err
	}
	recordQuery := db.Where("exam_r_id = ? AND exam_r_exam_id = ?", recordID, examID)
	if lock {
		recordQuery = recordQuery.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := recordQuery.First(&record).Error; err != nil {
		return record, exam, snapshot, "", err
	}
	if strings.TrimSpace(record.ReviewSnapshot) != "" {
		snapshot, err = decodeRecordReviewSnapshot(record.ReviewSnapshot)
		return record, exam, snapshot, "", err
	}
	// Old records have no original question snapshot. Never rewrite them as if current definitions were historical originals.
	notice := "历史答卷未保存题目快照，题目和参考答案按当前配置展示；请核对题目是否改动，评分上限仍以交卷时结果为准。"
	var raw string
	if record.PaperID > 0 {
		var paper model.ExamPaper
		if err := db.Where("exam_p_id = ?", record.PaperID).First(&paper).Error; err != nil {
			return record, exam, snapshot, notice, err
		}
		questions, err := examPaperQuestions(db, paper.QuestionIDs)
		if err != nil {
			return record, exam, snapshot, notice, err
		}
		raw = BuildPaperReviewSnapshot(questions, paper.PassScore)
	} else {
		raw = BuildSchemaReviewSnapshot(BuildSchemaGradingPlan(exam.Schema), exam.PassScore)
	}
	snapshot, err = decodeRecordReviewSnapshot(raw)
	return record, exam, snapshot, notice, err
}
