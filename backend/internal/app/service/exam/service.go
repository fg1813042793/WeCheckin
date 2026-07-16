package exam

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// Service manages online exam definitions, records, and statistics.
type Service struct{}

// NewService creates an exam service instance.
func NewService() *Service { return &Service{} }

type RecordDetailResult struct {
	Record  model.ExamRecord       `json:"record"`
	Answers map[string]interface{} `json:"answers"`
	Scoring map[string]bool        `json:"scoring"`
	Schema  interface{}            `json:"schema"`
}

func (s *Service) List(keyword, category, status string, page, pageSize int) ([]model.Exam, int64, error) {
	return s.ListContext(context.Background(), keyword, category, status, page, pageSize)
}

func (s *Service) ListContext(ctx context.Context, keyword, category, status string, page, pageSize int) ([]model.Exam, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Exam{})
	if keyword != "" {
		q = q.Where("`exam_title` LIKE ?", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`exam_category` = ?", category)
	}
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil && st >= 0 {
			q = q.Where("`exam_status` = ?", st)
		}
	}
	var total int64
	q.Count(&total)
	var list []model.Exam
	q.Order("`exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	return list, total, nil
}

func (s *Service) Get(id uint) (*model.Exam, error) {
	return s.GetContext(context.Background(), id)
}

func (s *Service) GetContext(ctx context.Context, id uint) (*model.Exam, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var exam model.Exam
	if err := db.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *Service) Create(req model.Exam) (*model.Exam, error) {
	return s.CreateContext(context.Background(), req)
}

func (s *Service) CreateContext(ctx context.Context, req model.Exam) (*model.Exam, error) {
	req = normalizeExamForCreate(req)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func normalizeExamForCreate(req model.Exam) model.Exam {
	req.AddTime = time.Now().UnixMilli()
	if req.Mode == "" {
		req.Mode = "exam"
	}
	if req.Schema == "" {
		req.Schema = `{"version":"2.0","questions":[],"setting":{}}`
	}
	if req.Settings == "" {
		req.Settings = "{}"
	}
	return req
}

func (s *Service) Update(id uint, updates map[string]interface{}) error {
	return s.UpdateContext(context.Background(), id, updates)
}

func (s *Service) UpdateContext(ctx context.Context, id uint, updates map[string]interface{}) error {
	updates["exam_edit_time"] = time.Now().UnixMilli()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Exam{}).Where("`exam_id` = ?", id).Updates(updates).Error
}

func (s *Service) SetStatus(id uint, status int) error {
	return s.SetStatusContext(context.Background(), id, status)
}

func (s *Service) SetStatusContext(ctx context.Context, id uint, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Exam{}).Where("`exam_id` = ?", id).Update("exam_status", status).Error
}

func (s *Service) Delete(id uint) error {
	return s.DeleteContext(context.Background(), id)
}

func (s *Service) DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Where("`exam_id` = ?", id).Delete(&model.Exam{}).Error; err != nil {
		return err
	}
	db.Where("`exam_r_exam_id` = ?", id).Delete(&model.ExamRecord{})
	return nil
}

// RecordList returns paged exam submission records.
func (s *Service) RecordList(examID int, keyword string, page, pageSize int) ([]model.ExamRecord, int64, error) {
	return s.RecordListContext(context.Background(), examID, keyword, page, pageSize)
}

func (s *Service) RecordListContext(ctx context.Context, examID int, keyword string, page, pageSize int) ([]model.ExamRecord, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examID)
	if keyword != "" {
		q = q.Where("`exam_r_user_id` LIKE ?", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ExamRecord
	q.Order("`exam_r_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	return list, total, nil
}

// RecordDetail returns one submission record with parsed answers and schema.
func (s *Service) RecordDetail(id uint) (*RecordDetailResult, error) {
	return s.RecordDetailContext(context.Background(), id)
}

func (s *Service) RecordDetailContext(ctx context.Context, id uint) (*RecordDetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	var exam model.Exam
	db.Where("`exam_id` = ?", record.ExamID).First(&exam)
	return decodeRecordDetailPayload(record, exam), nil
}

func decodeRecordDetailPayload(record model.ExamRecord, exam model.Exam) *RecordDetailResult {
	var answers map[string]interface{}
	if record.Answers != "" {
		_ = json.Unmarshal([]byte(record.Answers), &answers)
	}
	var scoring map[string]bool
	if record.Result != "" {
		_ = json.Unmarshal([]byte(record.Result), &scoring)
	}
	var schema interface{}
	if exam.Schema != "" {
		_ = json.Unmarshal([]byte(exam.Schema), &schema)
	}
	return &RecordDetailResult{
		Record:  record,
		Answers: answers,
		Scoring: scoring,
		Schema:  schema,
	}
}

func (s *Service) RecordDelete(id uint) {
	database.DB.Where("`exam_r_id` = ?", id).Delete(&model.ExamRecord{})
}

func (s *Service) RecordBatchDelete(ids string) {
	database.DB.Where("`exam_r_id` IN ?", strings.Split(ids, ",")).Delete(&model.ExamRecord{})
}

// Statistics returns summary, trend, score distribution, and field-level stats.
func (s *Service) Statistics(examID int) map[string]interface{} {
	var total, submitted, passed int64
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examID).Count(&total)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).Count(&submitted)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_pass` = 1", examID).Count(&passed)
	var passRate float64
	if submitted > 0 {
		passRate = float64(passed) / float64(submitted)
	}

	type dailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	var daily []dailyCount
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local).UnixMilli()
		end := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999, time.Local).UnixMilli()
		var cnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_submit_time` >= ? AND `exam_r_submit_time` <= ?", examID, start, end).Count(&cnt)
		daily = append(daily, dailyCount{Date: day.Format("01-02"), Count: cnt})
	}

	type scoreDist struct {
		Score int   `json:"score"`
		Count int64 `json:"count"`
	}
	var sds []scoreDist
	database.DB.Model(&model.ExamRecord{}).Select("FLOOR(`exam_r_score`/10)*10 as score, COUNT(*) as count").Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).Group("score").Order("score").Find(&sds)
	scoreDistMap := make(map[string]int64)
	for _, sd := range sds {
		key := strconv.Itoa(sd.Score) + "-" + strconv.Itoa(sd.Score+9)
		scoreDistMap[key] = sd.Count
	}

	var exam model.Exam
	database.DB.Where("`exam_id` = ?", examID).First(&exam)
	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	json.Unmarshal([]byte(exam.Schema), &schema)
	var fieldStats []map[string]interface{}
	for _, q := range schema.Questions {
		qid, _ := q["id"].(string)
		qtype, _ := q["type"].(string)
		title, _ := q["title"].(string)
		var nonEmpty, empty, totalCnt int64
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1 AND JSON_EXTRACT(`exam_r_answers`, '$.\""+qid+"\"') IS NOT NULL AND JSON_EXTRACT(`exam_r_answers`, '$.\""+qid+"\"') != ''", examID).Count(&nonEmpty)
		database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).Count(&totalCnt)
		empty = totalCnt - nonEmpty
		fieldStats = append(fieldStats, map[string]interface{}{
			"questionId": qid,
			"type":       qtype,
			"title":      title,
			"nonEmpty":   nonEmpty,
			"empty":      empty,
			"totalCount": totalCnt,
		})
	}
	return map[string]interface{}{
		"total":      total,
		"submitted":  submitted,
		"passed":     passed,
		"passRate":   passRate,
		"daily":      daily,
		"scoreDist":  scoreDistMap,
		"fieldStats": fieldStats,
	}
}
