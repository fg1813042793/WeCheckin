package survey

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/pkg/logger"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

// Submit POST /survey/submit
// @Tags 客户端-问卷
// @Summary 提交答卷
// @Param surveyId formData int true "问卷ID"
// @Param answers formData string true "答案JSON"
// @Param nickname formData string false "昵称"
// @Param session formData string false "答题会话"
// @Param startTime formData int false "开始时间"
// @Param device formData string false "设备信息"
// @Success 200 {object} response.Resp
// @Router /survey/submit [post]
func (h *ClientSurveyHandler) Submit(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID   uint                   `json:"surveyId"`
		Answers    map[string]interface{} `json:"answers"`
		Nickname   string                 `json:"nickname"`
		Session    string                 `json:"session"`
		StartTime  int64                  `json:"startTime"`
		Device     string                 `json:"device"`
		AutoSubmit bool                   `json:"autoSubmit"`
		DeviceID   string                 `json:"deviceId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	sv2, err2 := h.survey.Get(req.SurveyID)
	if err2 != nil {
		logger.Logger.Printf("[SurveySubmit] 问卷不存在 surveyId=%d", req.SurveyID)
		response.Fail(c, "问卷不存在")
		return
	}
	var settingsMap2 map[string]interface{}
	_ = json.Unmarshal([]byte(sv2.Settings), &settingsMap2)
	loginRequired2 := false
	if v, ok := settingsMap2["loginRequired"].(bool); ok {
		loginRequired2 = v
	} else if v, ok := settingsMap2["loginRequired"].(float64); ok {
		loginRequired2 = v != 0
	}
	token := string(c.Request.Header.Peek("Authorization"))
	if token != "" {
		_, prefix := tokenutil.GetTokenConfig("user")
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()
		jsonStr, err := rd.RDB.Get(redisCtx, prefix+"a:"+token).Result()
		if err == nil && jsonStr != "" {
			var tokenInfo struct {
				ID uint `json:"id"`
			}
			if err := json.Unmarshal([]byte(jsonStr), &tokenInfo); err == nil && tokenInfo.ID > 0 {
				c.Set("user_id", tokenInfo.ID)
			}
		}
	}
	if loginRequired2 {
		token := string(c.Request.Header.Peek("Authorization"))
		if token == "" {
			logger.Logger.Printf("[SurveySubmit] 需要登录 surveyId=%d", req.SurveyID)
			response.Fail(c, "请登录")
			return
		}
		_, prefix := tokenutil.GetTokenConfig("user")
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()
		jsonStr, err := rd.RDB.Get(redisCtx, prefix+"a:"+token).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[SurveySubmit] 登录校验失败 surveyId=%d", req.SurveyID)
			response.Fail(c, "请登录")
			return
		}
	}
	if req.StartTime <= 0 && req.Session != "" {
		redisKey := fmt.Sprintf("survey_session:%d:%s", req.SurveyID, req.Session)
		redisCtx, cancel := rd.OperationContext(ctx)
		defer cancel()
		v, err := rd.RDB.Get(redisCtx, redisKey).Int64()
		if err == nil {
			req.StartTime = v
			rd.RDB.Del(redisCtx, redisKey)
		}
	}
	uid := getUID(c)
	ip := c.ClientIP()
	resp, err := h.responses.Submit(req.SurveyID, uid, req.Nickname, req.StartTime, req.Answers, ip, req.Device, req.AutoSubmit, req.DeviceID)
	if err != nil {
		logger.Logger.Printf("[SurveySubmit] 失败 surveyId=%d uid=%d err=%s ip=%s device=%s", req.SurveyID, uid, err.Error(), ip, req.Device)
		response.Fail(c, err.Error())
		return
	}
	logger.Logger.Printf("[SurveySubmit] 成功 surveyId=%d uid=%d respId=%d ip=%s device=%s", req.SurveyID, uid, resp.ID, ip, req.Device)
	response.JSON(c, submitResponse{ID: resp.ID, SubmitTime: resp.SubmitTime})
}
