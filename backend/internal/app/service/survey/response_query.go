package survey

import (
	"encoding/json"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

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
func (r *ResponseService) List(surveyID uint, page, pageSize int, keyword string) ([]model.SurveyResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("`survey_resp_nickname` LIKE ? OR `survey_resp_user_id` LIKE ? OR `survey_resp_device` LIKE ? OR CAST(`survey_resp_id` AS CHAR) LIKE ?", like, like, like, like)
	}
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
