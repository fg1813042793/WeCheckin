package exam

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"

	"gorm.io/gorm"
)

type ClientLimitInfo struct {
	DeviceFull bool `json:"deviceFull,omitempty"`
	IPFull     bool `json:"ipFull,omitempty"`
}

var publishedExamListColumns = []string{
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
	"exam_settings",
	"exam_start_time",
	"exam_end_time",
	"exam_duration",
	"exam_max_attempts",
	"exam_show_score",
	"exam_max_response",
	"exam_mode",
	"exam_order",
	"exam_status",
	"add_time",
	"edit_time",
}

type PaperQuestionOptions struct {
	IncludeAnswer       bool
	IncludeAnalysis     bool
	IncludeExamAnswer   bool
	IncludeExamAnalysis bool
	IncludeCategory     bool
	IncludeDifficulty   bool
}

type SafeExamQuestion struct {
	ID                uint    `json:"id"`
	Type              string  `json:"type"`
	Title             string  `json:"title"`
	Options           string  `json:"options"`
	Score             int     `json:"score"`
	Difficulty        *int    `json:"difficulty,omitempty"`
	Category          *string `json:"category,omitempty"`
	Answer            *string `json:"answer,omitempty"`
	Analysis          *string `json:"analysis,omitempty"`
	ExamCorrectAnswer *string `json:"examCorrectAnswer,omitempty"`
	ExamAnalysis      *string `json:"examAnalysis,omitempty"`
}

type PaperQuestionResult struct {
	Paper     model.ExamPaper    `json:"paper"`
	Questions []SafeExamQuestion `json:"questions"`
}

type StartResult struct {
	Record    model.ExamRecord       `json:"record"`
	Paper     model.ExamPaper        `json:"paper"`
	Exam      model.Exam             `json:"exam"`
	Questions []SafeExamQuestion     `json:"questions"`
	Answers   map[string]interface{} `json:"answers"`
}

type RecordViewResult struct {
	Record    model.ExamRecord       `json:"record"`
	Exam      model.Exam             `json:"exam"`
	Paper     model.ExamPaper        `json:"paper"`
	Questions []SafeExamQuestion     `json:"questions"`
	Answers   map[string]interface{} `json:"answers"`
	Results   []examPkg.Result       `json:"results"`
}

type SessionResult struct {
	Record model.ExamRecord `json:"record"`
	Exam   model.Exam       `json:"exam"`
}

func (s *Service) PublishedListWithLimitsContext(ctx context.Context, keyword string, page, pageSize int, deviceID, clientIP string) ([]model.Exam, int64, map[uint]ClientLimitInfo, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Exam{}).Where("`exam_status` = 1")
	if keyword != "" {
		q = q.Where("`exam_title` LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}
	var list []model.Exam
	if err := q.Select(publishedExamListColumns).Order("`exam_order` DESC, `exam_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, nil, err
	}
	limitsMap, err := s.loadExamListLimitInfoContext(ctx, db, list, deviceID, clientIP)
	if err != nil {
		return nil, 0, nil, err
	}
	return list, total, limitsMap, nil
}

func (s *Service) loadExamListLimitInfoContext(ctx context.Context, db *gorm.DB, list []model.Exam, deviceID, clientIP string) (map[uint]ClientLimitInfo, error) {
	limitsMap := make(map[uint]ClientLimitInfo)
	if len(list) == 0 {
		return limitsMap, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	deviceLimits := make(map[uint]int)
	ipLimits := make(map[uint]int)
	deviceExamIDs := make([]uint, 0, len(list))
	ipExamIDs := make([]uint, 0, len(list))
	for _, item := range list {
		deviceLimit, ipLimit := examSubmissionLimits(item.Settings)
		if deviceLimit > 0 && deviceID != "" {
			deviceLimits[item.ID] = deviceLimit
			deviceExamIDs = append(deviceExamIDs, item.ID)
		}
		if ipLimit > 0 && clientIP != "" {
			ipLimits[item.ID] = ipLimit
			ipExamIDs = append(ipExamIDs, item.ID)
		}
	}

	if len(deviceExamIDs) > 0 {
		countByExamID, err := loadExamRecordCountsByFilter(db, deviceExamIDs, "`exam_r_device_id` = ?", deviceID)
		if err != nil {
			return nil, err
		}
		for examID, limit := range deviceLimits {
			if countByExamID[examID] >= int64(limit) {
				info := limitsMap[examID]
				info.DeviceFull = true
				limitsMap[examID] = info
			}
		}
	}
	if len(ipExamIDs) > 0 {
		countByExamID, err := loadExamRecordCountsByFilter(db, ipExamIDs, "`exam_r_add_ip` = ?", clientIP)
		if err != nil {
			return nil, err
		}
		for examID, limit := range ipLimits {
			if countByExamID[examID] >= int64(limit) {
				info := limitsMap[examID]
				info.IPFull = true
				limitsMap[examID] = info
			}
		}
	}
	return limitsMap, nil
}

func loadExamRecordCountsByFilter(db *gorm.DB, examIDs []uint, filter string, arg interface{}) (map[uint]int64, error) {
	type countRow struct {
		ExamID uint  `gorm:"column:exam_r_exam_id"`
		Count  int64 `gorm:"column:cnt"`
	}
	var rows []countRow
	if err := db.Model(&model.ExamRecord{}).
		Select("`exam_r_exam_id`, COUNT(*) AS cnt").
		Where("`exam_r_exam_id` IN ? AND `exam_r_status` >= 1", examIDs).
		Where(filter, arg).
		Group("`exam_r_exam_id`").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	countByExamID := make(map[uint]int64, len(rows))
	for _, row := range rows {
		countByExamID[row.ExamID] = row.Count
	}
	return countByExamID, nil
}

func (s *Service) PublishedExamContext(ctx context.Context, id uint) (*model.Exam, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var exam model.Exam
	if err := db.Where("`exam_id` = ? AND `exam_status` = 1", id).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *Service) UserDeptIDContext(ctx context.Context, userID uint) (uint, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var userDept model.UserDept
	err := db.Where("`user_dept_user_id` = ?", userID).First(&userDept).Error
	if err != nil {
		return 0, err
	}
	return userDept.DeptID, nil
}

func (s *Service) PaperQuestionsContext(ctx context.Context, paperID uint, options PaperQuestionOptions) (PaperQuestionResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var paper model.ExamPaper
	if err := db.Where("`exam_p_id` = ?", paperID).First(&paper).Error; err != nil {
		return PaperQuestionResult{}, err
	}
	questions, err := examPaperQuestions(db, paper.QuestionIDs)
	if err != nil {
		return PaperQuestionResult{}, err
	}
	return PaperQuestionResult{
		Paper:     paper,
		Questions: safeExamQuestions(questions, options),
	}, nil
}

func examPaperQuestions(db *gorm.DB, questionIDsJSON string) ([]model.ExamQuestion, error) {
	var questionIDs []uint
	_ = json.Unmarshal([]byte(questionIDsJSON), &questionIDs)
	var questions []model.ExamQuestion
	if len(questionIDs) == 0 {
		return questions, nil
	}
	err := db.Where("`exam_q_id` IN ?", questionIDs).Find(&questions).Error
	return questions, err
}

func safeExamQuestions(questions []model.ExamQuestion, options PaperQuestionOptions) []SafeExamQuestion {
	safe := make([]SafeExamQuestion, 0, len(questions))
	for _, question := range questions {
		item := SafeExamQuestion{
			ID:      question.ID,
			Type:    question.Type,
			Title:   question.Title,
			Options: question.Options,
			Score:   question.Score,
		}
		if options.IncludeDifficulty {
			item.Difficulty = &question.Difficulty
		}
		if options.IncludeCategory {
			item.Category = &question.Category
		}
		if options.IncludeAnswer {
			item.Answer = &question.Answer
		}
		if options.IncludeAnalysis {
			item.Analysis = &question.Analysis
		}
		if options.IncludeExamAnswer {
			item.ExamCorrectAnswer = &question.Answer
		}
		if options.IncludeExamAnalysis {
			item.ExamAnalysis = &question.Analysis
		}
		safe = append(safe, item)
	}
	return safe
}

func (s *Service) StartContext(ctx context.Context, examID int, userID uint, deviceID string) (StartResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var exam model.Exam
	if err := db.Where("`exam_id` = ?", examID).First(&exam).Error; err != nil {
		return StartResult{}, err
	}
	if exam.Status != 1 {
		return StartResult{}, ErrExamNotPublished
	}
	uidStr := strconv.FormatUint(uint64(userID), 10)
	if exam.MaxAttempts > 0 {
		var count int64
		if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ?", examID, uidStr).Count(&count).Error; err != nil {
			return StartResult{}, err
		}
		if int(count) >= exam.MaxAttempts {
			return StartResult{}, ErrExamMaxAttempts
		}
	}
	nowMs := time.Now().UnixMilli()
	if exam.StartTime > 0 && nowMs < exam.StartTime {
		return StartResult{}, ErrExamNotStarted
	}
	if exam.EndTime > 0 && nowMs > exam.EndTime {
		return StartResult{}, ErrExamEnded
	}
	var record model.ExamRecord
	err := db.Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` IN (0,1)", examID, uidStr).
		Order("`exam_r_id` DESC").
		First(&record).Error
	if err != nil {
		if !errorsIsRecordNotFound(err) {
			return StartResult{}, err
		}
		var paper model.ExamPaper
		if err := db.Where("`exam_p_id` = ?", exam.PaperID).First(&paper).Error; err != nil {
			return StartResult{}, ErrExamPaperNotFound
		}
		record = model.ExamRecord{
			ExamID:     uint(examID),
			PaperID:    exam.PaperID,
			UserID:     uidStr,
			TotalScore: paper.TotalScore,
			Status:     0,
			StartTime:  nowMs,
			DeviceID:   deviceID,
		}
		if err := db.Create(&record).Error; err != nil {
			return StartResult{}, err
		}
	}
	paperResult, err := s.PaperQuestionsContext(ctx, exam.PaperID, PaperQuestionOptions{IncludeDifficulty: true})
	if err != nil {
		return StartResult{}, ErrExamPaperNotFound
	}
	var prevAnswers map[string]interface{}
	if record.Answers != "" {
		_ = json.Unmarshal([]byte(record.Answers), &prevAnswers)
	}
	return StartResult{
		Record:    record,
		Paper:     paperResult.Paper,
		Exam:      exam,
		Questions: paperResult.Questions,
		Answers:   prevAnswers,
	}, nil
}

func (s *Service) SaveAnswerContext(ctx context.Context, recordID int, userID uint, answersJSON string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var record model.ExamRecord
	uidStr := strconv.FormatUint(uint64(userID), 10)
	if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uidStr).First(&record).Error; err != nil {
		if errorsIsRecordNotFound(err) {
			return ErrExamRecordNotFound
		}
		return err
	}
	if record.Status == 2 {
		return ErrExamRecordSubmitted
	}
	return db.Model(&record).Update("exam_r_answers", answersJSON).Error
}

func (s *Service) PaperSubmissionContext(ctx context.Context, recordID int, userID uint) (model.ExamRecord, model.ExamPaper, []model.ExamQuestion, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	uidStr := strconv.FormatUint(uint64(userID), 10)
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uidStr).First(&record).Error; err != nil {
		if errorsIsRecordNotFound(err) {
			return model.ExamRecord{}, model.ExamPaper{}, nil, ErrExamRecordNotFound
		}
		return model.ExamRecord{}, model.ExamPaper{}, nil, err
	}
	if record.Status == 2 {
		return model.ExamRecord{}, model.ExamPaper{}, nil, ErrExamRecordSubmitted
	}
	var paper model.ExamPaper
	if err := db.Where("`exam_p_id` = ?", record.PaperID).First(&paper).Error; err != nil {
		return model.ExamRecord{}, model.ExamPaper{}, nil, err
	}
	questions, err := examPaperQuestions(db, paper.QuestionIDs)
	if err != nil {
		return model.ExamRecord{}, model.ExamPaper{}, nil, err
	}
	return record, paper, questions, nil
}

func (s *Service) UpdatePaperSubmissionContext(ctx context.Context, recordID uint, updates map[string]interface{}) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.ExamRecord{}).Where("`exam_r_id` = ?", recordID).Updates(updates).Error
}

func (s *Service) CreateSchemaSubmissionContext(ctx context.Context, record *model.ExamRecord) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Create(record).Error
}

func (s *Service) RecordForUserContext(ctx context.Context, recordID int, userID uint) (RecordViewResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	uidStr := strconv.FormatUint(uint64(userID), 10)
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", recordID, uidStr).First(&record).Error; err != nil {
		if errorsIsRecordNotFound(err) {
			return RecordViewResult{}, ErrExamRecordNotFound
		}
		return RecordViewResult{}, err
	}
	var paper model.ExamPaper
	if err := db.Where("`exam_p_id` = ?", record.PaperID).First(&paper).Error; err != nil {
		return RecordViewResult{}, err
	}
	var exam model.Exam
	if err := db.Where("`exam_id` = ?", record.ExamID).First(&exam).Error; err != nil {
		return RecordViewResult{}, err
	}
	questions, err := examPaperQuestions(db, paper.QuestionIDs)
	if err != nil {
		return RecordViewResult{}, err
	}
	options := PaperQuestionOptions{IncludeAnswer: paper.ShowAnswer == 1 || record.Status == 2, IncludeAnalysis: paper.ShowAnswer == 1 || record.Status == 2}
	var prevAnswers map[string]interface{}
	if record.Answers != "" {
		_ = json.Unmarshal([]byte(record.Answers), &prevAnswers)
	}
	var results []examPkg.Result
	if record.Result != "" {
		_ = json.Unmarshal([]byte(record.Result), &results)
	}
	return RecordViewResult{
		Record:    record,
		Exam:      exam,
		Paper:     paper,
		Questions: safeExamQuestions(questions, options),
		Answers:   prevAnswers,
		Results:   results,
	}, nil
}

func (s *Service) UserRecordsContext(ctx context.Context, userID uint, limit int) ([]model.ExamRecord, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	uidStr := strconv.FormatUint(uint64(userID), 10)
	var list []model.ExamRecord
	err := db.Where("`exam_r_user_id` = ?", uidStr).Order("`exam_r_id` DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (s *Service) SessionResultContext(ctx context.Context, session string) (SessionResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var record model.ExamRecord
	if err := db.Where("`exam_r_session` = ?", session).First(&record).Error; err != nil {
		if errorsIsRecordNotFound(err) {
			return SessionResult{}, ErrExamRecordNotFound
		}
		return SessionResult{}, err
	}
	var exam model.Exam
	if err := db.Where("`exam_id` = ?", record.ExamID).First(&exam).Error; err != nil {
		return SessionResult{}, err
	}
	return SessionResult{Record: record, Exam: exam}, nil
}

func (s *Service) CheckLimitContext(ctx context.Context, exam *model.Exam, uidStr string, device string, deviceID string, ip string) (string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if exam.MaxResponse > 0 {
		var count int64
		if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", exam.ID).Count(&count).Error; err != nil {
			return "", err
		}
		if int(count) >= exam.MaxResponse {
			logger.Logger.Printf("[ExamCheckLimit] 答卷上限 examId=%d max=%d current=%d", exam.ID, exam.MaxResponse, count)
			return "已达最大答卷数", nil
		}
	}
	if exam.MaxAttempts > 0 && uidStr != "" {
		var count int64
		if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_user_id` = ? AND `exam_r_status` >= 1", exam.ID, uidStr).Count(&count).Error; err != nil {
			return "", err
		}
		if int(count) >= exam.MaxAttempts {
			logger.Logger.Printf("[ExamCheckLimit] 个人答题次数上限 examId=%d uid=%s max=%d current=%d", exam.ID, uidStr, exam.MaxAttempts, count)
			return "已达最大答题次数", nil
		}
	}
	deviceLimit, ipLimit := examSubmissionLimits(exam.Settings)
	if deviceLimit > 0 && deviceID != "" {
		var count int64
		if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_device_id` = ? AND `exam_r_status` >= 1", exam.ID, deviceID).Count(&count).Error; err != nil {
			return "", err
		}
		if int(count) >= deviceLimit {
			logger.Logger.Printf("[ExamCheckLimit] 设备次数上限 examId=%d limit=%d current=%d deviceId=%s", exam.ID, deviceLimit, count, deviceID)
			return "已达每台设备最大答题次数", nil
		}
	}
	if ipLimit > 0 && ip != "" {
		var count int64
		if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_add_ip` = ? AND `exam_r_status` >= 1", exam.ID, ip).Count(&count).Error; err != nil {
			return "", err
		}
		if int(count) >= ipLimit {
			logger.Logger.Printf("[ExamCheckLimit] IP次数上限 examId=%d limit=%d current=%d ip=%s", exam.ID, ipLimit, count, ip)
			return "已达每个IP最大答题次数", nil
		}
	}
	return "", nil
}

func examSubmissionLimits(settingsJSON string) (int, int) {
	var deviceLimit, ipLimit int
	if settingsJSON == "" {
		return deviceLimit, ipLimit
	}
	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJSON), &settingsMap); err != nil {
		return deviceLimit, ipLimit
	}
	return positiveLimitValue(settingsMap["deviceLimit"]), positiveLimitValue(settingsMap["ipLimit"])
}

func positiveLimitValue(value interface{}) int {
	switch v := value.(type) {
	case float64:
		if v <= 0 {
			return 0
		}
		return int(v)
	case string:
		limit, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil || limit <= 0 {
			return 0
		}
		return int(limit)
	default:
		return 0
	}
}

func errorsIsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
