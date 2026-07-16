package survey

import (
	"context"
	"encoding/json"

	calcPkg "wecheckin-backend/backend/internal/app/formkit/calc"
	questionPkg "wecheckin-backend/backend/internal/app/formkit/question"
	schemaPkg "wecheckin-backend/backend/internal/app/formkit/schema"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	"wecheckin-backend/backend/pkg/response"
)

// PublicValidate POST /survey/validate
// @Tags 客户端-表单工具
// @Summary 校验答案格式（通用）
// @Router /survey/validate [post]
func (h *ClientSurveyHandler) PublicValidate(ctx context.Context, c *app.RequestContext) {
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Schema   string                 `json:"schema"`
		Answers  map[string]interface{} `json:"answers"`
		Device   string                 `json:"device"`
		DeviceID string                 `json:"deviceId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	schema := req.Schema
	if req.SurveyID > 0 {
		db, cancel := database.WithContext(ctx)
		defer cancel()
		var sv model.Survey
		if err := db.Where("`survey_id` = ? AND `survey_status` = 1", req.SurveyID).First(&sv).Error; err != nil {
			logger.Logger.Printf("[SurveyValidate] 问卷不存在或未发布 surveyId=%d", req.SurveyID)
			response.Fail(c, "问卷不存在或未发布")
			return
		}
		if sv.MaxResponse > 0 {
			var count int64
			db.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", req.SurveyID).
				Count(&count)
			if count >= int64(sv.MaxResponse) {
				logger.Logger.Printf("[SurveyValidate] 回收上限已满 surveyId=%d max=%d current=%d", req.SurveyID, sv.MaxResponse, count)
				response.JSON(c, publicValidationResponse{Valid: false, Errors: []map[string]string{{"questionId": "", "message": "回收上限已满"}}})
				return
			}
		}
		var deviceLimit, ipLimit int
		if sv.Settings != "" {
			var settingsMap map[string]interface{}
			if err := json.Unmarshal([]byte(sv.Settings), &settingsMap); err == nil {
				if v, ok := settingsMap["deviceLimit"].(float64); ok {
					deviceLimit = int(v)
				}
				if v, ok := settingsMap["ipLimit"].(float64); ok {
					ipLimit = int(v)
				}
			}
		}
		clientIP := c.ClientIP()
		if deviceLimit > 0 && req.DeviceID != "" {
			var devCount int64
			db.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_device_id` = ? AND `survey_resp_status` = 1", req.SurveyID, req.DeviceID).
				Count(&devCount)
			if devCount >= int64(deviceLimit) {
				logger.Logger.Printf("[SurveyValidate] 设备次数上限 surveyId=%d limit=%d current=%d deviceId=%s", req.SurveyID, deviceLimit, devCount, req.DeviceID)
				response.JSON(c, publicValidationResponse{Valid: false, Errors: []map[string]string{{"questionId": "", "message": "该设备答题次数已达上限"}}})
				return
			}
		}
		if ipLimit > 0 && clientIP != "" {
			var ipCount int64
			db.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_ip` = ? AND `survey_resp_status` = 1", req.SurveyID, clientIP).
				Count(&ipCount)
			if ipCount >= int64(ipLimit) {
				logger.Logger.Printf("[SurveyValidate] IP次数上限 surveyId=%d limit=%d current=%d ip=%s", req.SurveyID, ipLimit, ipCount, clientIP)
				response.JSON(c, publicValidationResponse{Valid: false, Errors: []map[string]string{{"questionId": "", "message": "该IP答题次数已达上限"}}})
				return
			}
		}
		schema = sv.Schema
	}
	s, err := schemaPkg.Parse(schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	type fieldErr struct {
		QuestionID string `json:"questionId"`
		Message    string `json:"message"`
	}
	var errs []fieldErr
	for _, q := range s.Questions {
		v, ok := req.Answers[q.ID]
		if !ok && q.Required {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: "此项为必填"})
			continue
		}
		inst := questionPkg.Get(q.Type)
		if inst == nil {
			continue
		}
		if err := inst.Validate(v, q); err != nil {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: err.Error()})
		}
	}
	response.JSON(c, publicValidationResponse{Valid: len(errs) == 0, Errors: errs})
}

// PublicApply POST /survey/apply
// @Tags 客户端-表单工具
// @Summary 应用表单逻辑（通用）
// @Router /survey/apply [post]
func (h *ClientSurveyHandler) PublicApply(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema  string                 `json:"schema"`
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	eng := calcPkg.New()
	newAns, _ := eng.ApplyCalcValues(s, req.Answers)
	states, _ := eng.ApplyLogic(s, newAns)
	response.JSON(c, publicApplyResponse{Answers: newAns, States: states})
}
