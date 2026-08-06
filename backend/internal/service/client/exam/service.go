package exam

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"

	"gorm.io/gorm"
)

// Service manages online exam definitions, records, and statistics.
type Service struct{}

// NewService creates an exam service instance.
func NewService() *Service { return &Service{} }

var adminExamListColumns = []string{
	"exam_id",
	"exam_title",
	"exam_desc",
	"exam_category",
	"exam_tags",
	"exam_visibility",
	"exam_allow_multi",
	"exam_anonymous",
	"exam_show_result",
	"exam_paper_id",
	"exam_start_time",
	"exam_end_time",
	"exam_duration",
	"exam_max_attempts",
	"exam_show_score",
	"exam_max_response",
	"exam_dept_ids",
	"exam_mode",
	"exam_publish_dept_ids",
	"exam_qr",
	"exam_status",
	"exam_order",
	"create_dept_id",
	"create_by",
	"update_by",
	"update_dept_id",
	"add_time",
	"edit_time",
}

func scopedExamQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Exam{}, access.ExamAuditFields)
}

func ensureExamVisibleContext(ctx context.Context, db *gorm.DB, examID uint, adminID uint) error {
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return queryBuilder.Where("`exam_id` = ?", examID).First(&model.Exam{}).Error
}

type RecordDetailResult struct {
	Record  model.ExamRecord       `json:"record"`
	Answers map[string]interface{} `json:"answers"`
	Scoring map[string]bool        `json:"scoring"`
	Schema  interface{}            `json:"schema"`
}

type DetailResult struct {
	Exam          *model.Exam `json:"exam"`
	ResponseCount int64       `json:"responseCount"`
	Schema        string      `json:"schema"`
}

type StatisticsResult struct {
	Total      int64                `json:"total"`
	Submitted  int64                `json:"submitted"`
	Passed     int64                `json:"passed"`
	PassRate   float64              `json:"passRate"`
	Daily      []DailyCount         `json:"daily"`
	ScoreDist  map[string]int64     `json:"scoreDist"`
	FieldStats []QuestionFieldStats `json:"fieldStats"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type QuestionFieldStats struct {
	QuestionID string `json:"questionId"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	NonEmpty   int64  `json:"nonEmpty"`
	Empty      int64  `json:"empty"`
	TotalCount int64  `json:"totalCount"`
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

func (s *Service) ListForAdminContext(ctx context.Context, keyword, category, status string, page, pageSize int, adminID uint) ([]model.Exam, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, 0, err
	}
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
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Exam
	if err := q.Select(adminExamListColumns).Order("`exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
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

func (s *Service) GetForAdminContext(ctx context.Context, id uint, adminID uint) (*model.Exam, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var exam model.Exam
	if err := queryBuilder.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *Service) Detail(id uint) (*DetailResult, error) {
	return s.DetailContext(context.Background(), id)
}

func (s *Service) DetailContext(ctx context.Context, id uint) (*DetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var exam model.Exam
	if err := db.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		return nil, err
	}
	var responseCount int64
	if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", id).Count(&responseCount).Error; err != nil {
		return nil, err
	}
	return &DetailResult{
		Exam:          &exam,
		ResponseCount: responseCount,
		Schema:        exam.Schema,
	}, nil
}

func (s *Service) DetailForAdminContext(ctx context.Context, id uint, adminID uint) (*DetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var exam model.Exam
	if err := queryBuilder.Where("`exam_id` = ?", id).First(&exam).Error; err != nil {
		return nil, err
	}
	var responseCount int64
	if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", id).Count(&responseCount).Error; err != nil {
		return nil, err
	}
	return &DetailResult{
		Exam:          &exam,
		ResponseCount: responseCount,
		Schema:        exam.Schema,
	}, nil
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
	now := time.Now().UnixMilli()
	req.AddTime = now
	req.EditTime = now
	if req.UpdateBy == 0 {
		req.UpdateBy = req.CreateBy
	}
	if req.UpdateDeptID == 0 {
		req.UpdateDeptID = req.DeptID
	}
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
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["edit_time"] = time.Now().UnixMilli()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Exam{}).Where("`exam_id` = ?", id).Updates(updates).Error
}

func (s *Service) UpdateForAdminContext(ctx context.Context, id uint, updates map[string]interface{}, adminID uint) error {
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["update_by"] = adminID
	updates["edit_time"] = time.Now().UnixMilli()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`exam_id` = ?", id).Updates(updates))
}

func (s *Service) SetStatus(id uint, status int) error {
	return s.SetStatusContext(context.Background(), id, status)
}

func (s *Service) SetStatusContext(ctx context.Context, id uint, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Exam{}).Where("`exam_id` = ?", id).Update("exam_status", status).Error
}

func (s *Service) SetStatusForAdminContext(ctx context.Context, id uint, status int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`exam_id` = ?", id).Update("exam_status", status))
}

func (s *Service) Delete(id uint) error {
	return s.DeleteContext(context.Background(), id)
}

func (s *Service) DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`exam_id` = ?", id).Delete(&model.Exam{}).Error; err != nil {
			return err
		}
		return tx.Where("`exam_r_exam_id` = ?", id).Delete(&model.ExamRecord{}).Error
	})
}

func (s *Service) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedExamQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`exam_id` = ?", id).First(&model.Exam{}).Error; err != nil {
			return err
		}
		if err := access.RequireRowsAffected(tx.Where("`exam_id` = ?", id).Delete(&model.Exam{})); err != nil {
			return err
		}
		return tx.Where("`exam_r_exam_id` = ?", id).Delete(&model.ExamRecord{}).Error
	})
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

func (s *Service) RecordListForAdminContext(ctx context.Context, examID int, keyword string, page, pageSize int, adminID uint) ([]model.ExamRecord, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureExamVisibleContext(ctx, db, uint(examID), adminID); err != nil {
		return nil, 0, err
	}
	return s.RecordListContext(ctx, examID, keyword, page, pageSize)
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

func (s *Service) RecordDetailForAdminContext(ctx context.Context, id uint, adminID uint) (*RecordDetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	if err := ensureExamVisibleContext(ctx, db, record.ExamID, adminID); err != nil {
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
	_ = s.RecordDeleteContext(context.Background(), id)
}

func (s *Service) RecordDeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`exam_r_id` = ?", id).Delete(&model.ExamRecord{}).Error
}

func (s *Service) RecordDeleteForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ?", id).First(&record).Error; err != nil {
		return err
	}
	if err := ensureExamVisibleContext(ctx, db, record.ExamID, adminID); err != nil {
		return err
	}
	return access.RequireRowsAffected(db.Where("`exam_r_id` = ?", id).Delete(&model.ExamRecord{}))
}

func (s *Service) RecordBatchDelete(ids string) {
	_ = s.RecordBatchDeleteContext(context.Background(), ids)
}

func (s *Service) RecordBatchDeleteContext(ctx context.Context, ids string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`exam_r_id` IN ?", strings.Split(ids, ",")).Delete(&model.ExamRecord{}).Error
}

func (s *Service) RecordBatchDeleteForAdminContext(ctx context.Context, ids string, adminID uint) error {
	for _, item := range strings.Split(ids, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id, err := strconv.Atoi(item)
		if err != nil {
			return err
		}
		if err := s.RecordDeleteForAdminContext(ctx, uint(id), adminID); err != nil {
			return err
		}
	}
	return nil
}

// Statistics returns summary, trend, score distribution, and field-level stats.
func (s *Service) Statistics(examID int) StatisticsResult {
	result, _ := s.StatisticsContext(context.Background(), examID)
	return result
}

func (s *Service) StatisticsContext(ctx context.Context, examID int) (StatisticsResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	type summaryRow struct {
		Total     int64 `gorm:"column:total"`
		Submitted int64 `gorm:"column:submitted"`
		Passed    int64 `gorm:"column:passed"`
	}
	var summary summaryRow
	if err := db.Model(&model.ExamRecord{}).
		Select("COUNT(*) AS total, COALESCE(SUM(CASE WHEN `exam_r_status` >= 1 THEN 1 ELSE 0 END), 0) AS submitted, COALESCE(SUM(CASE WHEN `exam_r_pass` = 1 THEN 1 ELSE 0 END), 0) AS passed").
		Where("`exam_r_exam_id` = ?", examID).
		Scan(&summary).Error; err != nil {
		return StatisticsResult{}, err
	}
	var passRate float64
	if summary.Submitted > 0 {
		passRate = float64(summary.Passed) / float64(summary.Submitted)
	}

	now := time.Now()
	daily := make([]DailyCount, 7)
	dayCountMap := map[string]int64{}
	sevenDayStart := time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, time.Local)
	tomorrowStart := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	type dailyRow struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}
	var dailyRows []dailyRow
	if err := db.Model(&model.ExamRecord{}).
		Select("DATE_FORMAT(FROM_UNIXTIME(`exam_r_submit_time` / 1000), '%m-%d') AS date, COUNT(*) AS count").
		Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1 AND `exam_r_submit_time` >= ? AND `exam_r_submit_time` < ?", examID, sevenDayStart.UnixMilli(), tomorrowStart.UnixMilli()).
		Group("date").
		Scan(&dailyRows).Error; err != nil {
		return StatisticsResult{}, err
	}
	for _, row := range dailyRows {
		dayCountMap[row.Date] = row.Count
	}
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		date := day.Format("01-02")
		daily[6-i] = DailyCount{Date: date, Count: dayCountMap[date]}
	}

	type scoreDist struct {
		Score int   `json:"score"`
		Count int64 `json:"count"`
	}
	var sds []scoreDist
	if err := db.Model(&model.ExamRecord{}).
		Select("FLOOR(`exam_r_score`/10)*10 as score, COUNT(*) as count").
		Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).
		Group("score").
		Order("score").
		Find(&sds).Error; err != nil {
		return StatisticsResult{}, err
	}
	scoreDistMap := make(map[string]int64)
	for _, sd := range sds {
		key := strconv.Itoa(sd.Score) + "-" + strconv.Itoa(sd.Score+9)
		scoreDistMap[key] = sd.Count
	}

	var exam model.Exam
	if err := db.Where("`exam_id` = ?", examID).First(&exam).Error; err != nil {
		return StatisticsResult{}, err
	}
	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	_ = json.Unmarshal([]byte(exam.Schema), &schema)
	var answerRows []model.ExamRecord
	if err := db.Select("exam_r_answers").
		Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).
		Find(&answerRows).Error; err != nil {
		return StatisticsResult{}, err
	}
	nonEmptyByQuestionID := countExamQuestionNonEmptyAnswers(schema.Questions, answerRows)
	var fieldStats []QuestionFieldStats
	for _, q := range schema.Questions {
		qid := questionIDString(q["id"])
		qtype, _ := q["type"].(string)
		title, _ := q["title"].(string)
		nonEmpty := nonEmptyByQuestionID[qid]
		fieldStats = append(fieldStats, QuestionFieldStats{
			QuestionID: qid,
			Type:       qtype,
			Title:      title,
			NonEmpty:   nonEmpty,
			Empty:      summary.Submitted - nonEmpty,
			TotalCount: summary.Submitted,
		})
	}
	return StatisticsResult{
		Total:      summary.Total,
		Submitted:  summary.Submitted,
		Passed:     summary.Passed,
		PassRate:   passRate,
		Daily:      daily,
		ScoreDist:  scoreDistMap,
		FieldStats: fieldStats,
	}, nil
}

func (s *Service) StatisticsForAdminContext(ctx context.Context, examID int, adminID uint) (StatisticsResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureExamVisibleContext(ctx, db, uint(examID), adminID); err != nil {
		return StatisticsResult{}, err
	}
	return s.StatisticsContext(ctx, examID)
}

func countExamQuestionNonEmptyAnswers(questions []map[string]interface{}, answerRows []model.ExamRecord) map[string]int64 {
	result := make(map[string]int64, len(questions))
	if len(questions) == 0 || len(answerRows) == 0 {
		return result
	}
	questionIDs := make([]string, 0, len(questions))
	for _, q := range questions {
		qid := questionIDString(q["id"])
		if qid == "" {
			continue
		}
		questionIDs = append(questionIDs, qid)
		result[qid] = 0
	}
	for _, row := range answerRows {
		var answers map[string]interface{}
		if row.Answers == "" || json.Unmarshal([]byte(row.Answers), &answers) != nil {
			continue
		}
		for _, qid := range questionIDs {
			if answerValueNonEmpty(answers[qid]) {
				result[qid]++
			}
		}
	}
	return result
}

func questionIDString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return ""
	}
}

func answerValueNonEmpty(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}
