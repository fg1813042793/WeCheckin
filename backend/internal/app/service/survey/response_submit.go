package survey

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/app/formkit/calc"
	"wecheckin-backend/backend/internal/app/formkit/schema"
	poststatservice "wecheckin-backend/backend/internal/app/service/poststat"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
)

// Submit 提交答卷（核心）
//   - 校验答案（必填 + regex）
//   - 应用计算值
//   - 应用显隐逻辑
//   - 持久化
func (r *ResponseService) Submit(surveyID uint, userID uint, nickname string, startTime int64, answers map[string]interface{}, ip, device string, autoSubmit bool, deviceId string) (*model.SurveyResponse, error) {
	sv, err := r.Survey.Get(surveyID)
	if err != nil {
		logger.Logger.Printf("[Submit] 问卷不存在 surveyId=%d", surveyID)
		return nil, errors.New("问卷不存在")
	}
	if sv.Status != 1 {
		logger.Logger.Printf("[Submit] 问卷已停用 surveyId=%d title=%s", surveyID, sv.Title)
		return nil, errors.New("问卷已停用")
	}
	now := time.Now().UnixMilli()
	if sv.StartTime > 0 && now < sv.StartTime {
		logger.Logger.Printf("[Submit] 问卷未开始 surveyId=%d startTime=%d", surveyID, sv.StartTime)
		return nil, errors.New("问卷未开始")
	}
	if sv.EndTime > 0 && now > sv.EndTime {
		logger.Logger.Printf("[Submit] 问卷已结束 surveyId=%d endTime=%d", surveyID, sv.EndTime)
		return nil, errors.New("问卷已结束")
	}
	if sv.MaxResponse > 0 {
		var cnt int64
		database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).Count(&cnt)
		if int(cnt) >= sv.MaxResponse {
			logger.Logger.Printf("[Submit] 已达答卷上限 surveyId=%d max=%d current=%d", surveyID, sv.MaxResponse, cnt)
			return nil, errors.New("已达答卷上限")
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
	if deviceLimit > 0 && deviceId != "" {
		var devCnt int64
		database.DB.Model(&model.SurveyResponse{}).
			Where("`survey_resp_survey_id` = ? AND `survey_resp_device_id` = ? AND `survey_resp_status` = 1", surveyID, deviceId).
			Count(&devCnt)
		if devCnt >= int64(deviceLimit) {
			logger.Logger.Printf("[Submit] 设备次数上限 surveyId=%d limit=%d current=%d deviceId=%s", surveyID, deviceLimit, devCnt, deviceId)
			return nil, errors.New("该设备答题次数已达上限")
		}
	}
	if ipLimit > 0 && ip != "" {
		var ipCnt int64
		database.DB.Model(&model.SurveyResponse{}).
			Where("`survey_resp_survey_id` = ? AND `survey_resp_ip` = ? AND `survey_resp_status` = 1", surveyID, ip).
			Count(&ipCnt)
		if ipCnt >= int64(ipLimit) {
			logger.Logger.Printf("[Submit] IP次数上限 surveyId=%d limit=%d current=%d ip=%s", surveyID, ipLimit, ipCnt, ip)
			return nil, errors.New("该IP答题次数已达上限")
		}
	}
	if sv.AllowMulti == 0 && sv.Anonymous != 1 && userID > 0 {
		var cnt int64
		database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_user_id` = ?", surveyID, userID).Count(&cnt)
		if cnt > 0 {
			logger.Logger.Printf("[Submit] 已提交过 surveyId=%d userId=%d", surveyID, userID)
			return nil, errors.New("已提交过此问卷")
		}
	}
	sch, _ := schema.Parse(sv.Schema)
	if sch != nil {
		eng := calc.New()
		answers, _ = eng.ApplyCalcValues(sch, answers)
		_, _ = eng.ApplyLogic(sch, answers)
	}
	answersJSON, _ := json.Marshal(answers)

	if nickname == "" && userID > 0 && sv.Anonymous != 1 {
		var u model.User
		if err := database.DB.Where("`id` = ?", userID).First(&u).Error; err == nil {
			nickname = u.Name
		}
	}
	if nickname == "" && sch != nil {
		personalTypes := []string{"name", "phone", "email", "studentId", "employeeId"}
		for _, pt := range personalTypes {
			for _, q := range sch.Questions {
				if q.Type == pt {
					if val, ok := answers[q.ID]; ok && val != nil {
						if s, ok := val.(string); ok && s != "" {
							nickname = s
							break
						}
					}
				}
			}
			if nickname != "" {
				break
			}
		}
	}
	st := startTime
	if st == 0 {
		st = now
	}
	duration := 0
	if st > 0 && now > st {
		duration = int((now - st) / 1000)
	}
	browser, deviceType, platformType := parseUA(device)
	autoSubmitVal := 0
	if autoSubmit {
		autoSubmitVal = 1
	}
	resp := &model.SurveyResponse{
		SurveyID:     surveyID,
		UserID:       userIDToStr(userID, sv.Anonymous == 1),
		Nickname:     nickname,
		Answers:      string(answersJSON),
		Duration:     duration,
		Status:       1,
		IP:           ip,
		Device:       device,
		DeviceID:     deviceId,
		Browser:      browser,
		DeviceType:   deviceType,
		PlatformType: platformType,
		StartTime:    st,
		SubmitTime:   now,
		IsAutoSubmit: autoSubmitVal,
		AddTime:      now,
	}
	if err := database.DB.Create(resp).Error; err != nil {
		logger.Logger.Printf("[Submit] 持久化失败 surveyId=%d err=%s", surveyID, err.Error())
		return nil, err
	}
	database.DB.Model(&model.SurveyChannel{}).Where("`survey_ch_survey_id` = ?", surveyID).
		UpdateColumn("survey_ch_submit_cnt", gorm.Expr("`survey_ch_submit_cnt` + 1"))
	logger.Logger.Printf("[Submit] 成功 surveyId=%d userId=%d respId=%d ip=%s device=%s", surveyID, userID, resp.ID, ip, device)
	go poststatservice.Process(surveyID, userID, nickname, resp.Answers)
	return resp, nil
}
