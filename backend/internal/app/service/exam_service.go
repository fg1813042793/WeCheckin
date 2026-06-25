package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/internal/model"
)

type ExamService struct{}

func NewExamService() *ExamService { return &ExamService{} }

func (s *ExamService) List(keyword, category, status string, page, pageSize int) ([]model.Exam, int64, error) {
	q := database.DB.Model(&model.Exam{})
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

func (s *ExamService) Get(id uint) (*model.Exam, error) {
	var exam model.Exam
	if err := database.DB.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *ExamService) Create(req model.Exam) (*model.Exam, error) {
	now := time.Now().UnixMilli()
	req.AddTime = now
	if req.Mode == "" {
		req.Mode = "exam"
	}
	if req.Schema == "" {
		req.Schema = `{"version":"2.0","questions":[],"setting":{}}`
	}
	if req.Settings == "" {
		req.Settings = "{}"
	}
	if err := database.DB.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *ExamService) Update(id uint, updates map[string]interface{}) error {
	updates["exam_edit_time"] = time.Now().UnixMilli()
	return database.DB.Model(&model.Exam{}).Where("`exam_id` = ?", id).Updates(updates).Error
}

func (s *ExamService) SetStatus(id uint, status int) error {
	return database.DB.Model(&model.Exam{}).Where("`exam_id` = ?", id).Update("exam_status", status).Error
}

func (s *ExamService) Delete(id uint) error {
	if err := database.DB.Where("`exam_id` = ?", id).Delete(&model.Exam{}).Error; err != nil {
		return err
	}
	database.DB.Where("`exam_r_exam_id` = ?", id).Delete(&model.ExamRecord{})
	return nil
}

// RecordList 考试记录列表
func (s *ExamService) RecordList(examID int, keyword string, page, pageSize int) ([]model.ExamRecord, int64, error) {
	q := database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examID)
	if keyword != "" {
		q = q.Where("`exam_r_user_id` LIKE ?", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ExamRecord
	q.Order("`exam_r_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	return list, total, nil
}

// RecordDetail 考试记录详情
func (s *ExamService) RecordDetail(id uint) (map[string]interface{}, error) {
	var record model.ExamRecord
	if err := database.DB.Where("`exam_r_id` = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	var exam model.Exam
	database.DB.Where("`exam_id` = ?", record.ExamID).First(&exam)
	var answers map[string]interface{}
	json.Unmarshal([]byte(record.Answers), &answers)
	var scoring map[string]bool
	json.Unmarshal([]byte(record.Result), &scoring)
	var schema interface{}
	if exam.Schema != "" {
		json.Unmarshal([]byte(exam.Schema), &schema)
	}
	return map[string]interface{}{
		"record":  record,
		"answers": answers,
		"scoring": scoring,
		"schema":  schema,
	}, nil
}

func (s *ExamService) RecordDelete(id uint) {
	database.DB.Where("`exam_r_id` = ?", id).Delete(&model.ExamRecord{})
}

func (s *ExamService) RecordBatchDelete(ids string) {
	database.DB.Where("`exam_r_id` IN ?", strings.Split(ids, ",")).Delete(&model.ExamRecord{})
}

// Statistics 考试统计
func (s *ExamService) Statistics(examID int) map[string]interface{} {
	var total, submitted, passed int64
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", examID).Count(&total)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).Count(&submitted)
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_pass` = 1", examID).Count(&passed)
	var passRate float64
	if submitted > 0 {
		passRate = float64(passed) / float64(submitted)
	}
	// daily trends (last 7 days)
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
	// score distribution
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
	// field stats from saved schema
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
		fs := map[string]interface{}{
			"questionId": qid,
			"type":       qtype,
			"title":      title,
			"nonEmpty":   nonEmpty,
			"empty":      empty,
			"totalCount": totalCnt,
		}
		fieldStats = append(fieldStats, fs)
	}
	return map[string]interface{}{
		"total":     total,
		"submitted": submitted,
		"passed":    passed,
		"passRate":  passRate,
		"daily":     daily,
		"scoreDist": scoreDistMap,
		"fieldStats": fieldStats,
	}
}
