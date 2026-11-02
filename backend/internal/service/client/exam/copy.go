package exam

import (
	"context"
	"encoding/json"
	"time"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

// CopyForAdminContext creates an unpublished independent exam from a visible source exam.
func (s *Service) CopyForAdminContext(ctx context.Context, id, createBy, deptID, adminID uint) (*model.Exam, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var source model.Exam
	if err := queryBuilder.Where("`exam_id` = ?", id).First(&source).Error; err != nil {
		return nil, err
	}
	copyExam := buildExamCopy(source, createBy, deptID, time.Now().UnixMilli())
	if err := db.Create(&copyExam).Error; err != nil {
		return nil, err
	}
	return &copyExam, nil
}

func buildExamCopy(source model.Exam, createBy, deptID uint, now int64) model.Exam {
	copyExam := source
	copyExam.ID = 0
	copyExam.Title = source.Title + " (副本)"
	copyExam.Status = 0
	copyExam.QR = ""
	copyExam.Settings = resetExamCopyShareSettings(source.Settings)
	copyExam.CreateBy = createBy
	copyExam.UpdateBy = createBy
	copyExam.DeptID = deptID
	copyExam.UpdateDeptID = deptID
	copyExam.AddTime = now
	copyExam.EditTime = now
	return copyExam
}

func resetExamCopyShareSettings(raw string) string {
	settings := make(map[string]interface{})
	if json.Unmarshal([]byte(raw), &settings) != nil {
		settings = make(map[string]interface{})
	}
	settings["publicUrl"] = ""
	data, _ := json.Marshal(settings)
	return string(data)
}
