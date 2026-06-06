package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/formkit/calc"
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
	"wecheckin-backend/backend/internal/model"
)

// ResponseService 答卷服务
type ResponseService struct {
	Survey *SurveyService
}

func NewResponseService() *ResponseService {
	return &ResponseService{Survey: NewSurveyService()}
}

// Submit 提交答卷（核心）
//   - 校验答案（必填 + regex）
//   - 应用计算值
//   - 应用显隐逻辑
//   - 持久化
func (r *ResponseService) Submit(surveyID uint, userID uint, nickname string, startTime int64, answers map[string]interface{}, ip, device string) (*model.SurveyResponse, error) {
	sv, err := r.Survey.Get(surveyID)
	if err != nil {
		return nil, errors.New("问卷不存在")
	}
	if sv.Status != 1 {
		return nil, errors.New("问卷已停用")
	}
	// 时间窗
	now := time.Now().UnixMilli()
	if sv.StartTime > 0 && now < sv.StartTime {
		return nil, errors.New("问卷未开始")
	}
	if sv.EndTime > 0 && now > sv.EndTime {
		return nil, errors.New("问卷已截止")
	}
	// 限额
	if sv.MaxResponse > 0 {
		var cnt int64
		database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).Count(&cnt)
		if int(cnt) >= sv.MaxResponse {
			return nil, errors.New("已达答卷上限")
		}
	}
	// 限人
	if sv.AllowMulti == 0 && sv.Anonymous != 1 && userID > 0 {
		var cnt int64
		database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_user_id` = ?", surveyID, userID).Count(&cnt)
		if cnt > 0 {
			return nil, errors.New("已提交过此问卷")
		}
	}
	// 应用计算值
	sch, _ := schema.Parse(sv.Schema)
	if sch != nil {
		eng := calc.New()
		answers, _ = eng.ApplyCalcValues(sch, answers)
		_, _ = eng.ApplyLogic(sch, answers)
	}
	answersJSON, _ := json.Marshal(answers)

	// 查昵称
	if nickname == "" && userID > 0 && sv.Anonymous != 1 {
		var u model.User
		if err := database.DB.Where("`id` = ?", userID).First(&u).Error; err == nil {
			nickname = u.Name
		}
	}
	// 计算时长
	st := startTime
	if st == 0 {
		st = now
	}
	duration := 0
	if st > 0 && now > st {
		duration = int((now - st) / 1000)
	}
	resp := &model.SurveyResponse{
		SurveyID:   surveyID,
		UserID:     userIDToStr(userID, sv.Anonymous == 1),
		Nickname:   nickname,
		Answers:    string(answersJSON),
		Duration:   duration,
		Status:     1,
		IP:         ip,
		Device:     device,
		StartTime:  st,
		SubmitTime: now,
		AddTime:    now,
	}
	if err := database.DB.Create(resp).Error; err != nil {
		return nil, err
	}
	// 渠道计数
	database.DB.Model(&model.SurveyChannel{}).Where("`survey_ch_survey_id` = ?", surveyID).
		UpdateColumn("survey_ch_submit_cnt", gorm.Expr("`survey_ch_submit_cnt` + 1"))
	return resp, nil
}

// SaveDraft 暂存草稿
func (r *ResponseService) SaveDraft(surveyID uint, userID uint, answers map[string]interface{}) (*model.SurveyResponse, error) {
	sv, err := r.Survey.Get(surveyID)
	if err != nil {
		return nil, err
	}
	answersJSON, _ := json.Marshal(answers)
	now := time.Now().UnixMilli()
	resp := &model.SurveyResponse{
		SurveyID: surveyID,
		UserID:   userIDToStr(userID, sv.Anonymous == 1),
		Answers:  string(answersJSON),
		Status:   0, // 草稿
		AddTime:  now,
	}
	database.DB.Create(resp)
	return resp, nil
}

// Get 读取答卷（管理员用，含答案）
func (r *ResponseService) Get(id uint) (*model.SurveyResponse, error) {
	var resp model.SurveyResponse
	if err := database.DB.Where("`survey_resp_id` = ?", id).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetByUser 取某用户对某问卷的答卷
func (r *ResponseService) GetByUser(surveyID uint, userID uint) (*model.SurveyResponse, error) {
	var resp model.SurveyResponse
	uidStr := userIDToStr(userID, false)
	if err := database.DB.Where("`survey_resp_survey_id` = ? AND `survey_resp_user_id` = ?", surveyID, uidStr).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

// List 答卷列表（admin）
func (r *ResponseService) List(surveyID uint, page, pageSize int) ([]model.SurveyResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID)
	var total int64
	q.Count(&total)
	var list []model.SurveyResponse
	err := q.Order("`survey_resp_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ParseAnswers 解析答卷答案到 map
func (r *ResponseService) ParseAnswers(resp *model.SurveyResponse) map[string]interface{} {
	var m map[string]interface{}
	if resp.Answers == "" {
		return m
	}
	_ = json.Unmarshal([]byte(resp.Answers), &m)
	return m
}

// userIDToStr 转字符串，匿名时为空
func userIDToStr(uid uint, anonymous bool) string {
	if anonymous {
		return ""
	}
	return strconv.FormatUint(uint64(uid), 10)
}

// ValidateAnswers 答案校验（必填 + regex）
// 失败返回详细错误
func ValidateAnswers(sv *model.Survey, answers map[string]interface{}) []ValidationError {
	sch, err := schema.Parse(sv.Schema)
	if err != nil {
		return []ValidationError{{QuestionID: "", Message: "schema 解析失败: " + err.Error()}}
	}
	var errs []ValidationError
	ans := answers
	for _, q := range sch.Questions {
		val, _ := ans[q.ID]
		// 必填
		if q.Required {
			if val == nil || isEmpty(val) {
				errs = append(errs, ValidationError{QuestionID: q.ID, Message: q.Title + " 必填"})
				continue
			}
		}
		// 题型校验
		if qt := question.Get(q.Type); qt != nil {
			if verr := qt.Validate(val, q); verr != nil {
				errs = append(errs, ValidationError{QuestionID: q.ID, Message: verr.Error()})
			}
		}
	}
	return errs
}

// ValidationError 校验错误
type ValidationError struct {
	QuestionID string `json:"questionId"`
	Message    string `json:"message"`
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	if a, ok := v.([]interface{}); ok {
		return len(a) == 0
	}
	return false
}
