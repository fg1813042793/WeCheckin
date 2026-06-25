package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	calcPkg "wecheckin-backend/backend/internal/app/formkit/calc"
	questionPkg "wecheckin-backend/backend/internal/app/formkit/question"
	schemaPkg "wecheckin-backend/backend/internal/app/formkit/schema"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
	"wecheckin-backend/backend/pkg/tokenutil"
	"wecheckin-backend/backend/internal/app/service"
	"wecheckin-backend/backend/pkg/logger"
)

type ClientSurveyHandler struct {
	survey    *service.SurveyService
	responses *service.ResponseService
}

func NewClientSurveyHandler() *ClientSurveyHandler { return &ClientSurveyHandler{} }

func (h *ClientSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = service.NewSurveyService()
	}
	if h.responses == nil {
		h.responses =  service.NewResponseService()
	}
}

// List GET /survey/list
// @Tags 客户端-问卷
// @Summary 获取问卷列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Success 200 {object} response.Resp
// @Router /survey/list [get]
func (h *ClientSurveyHandler) List(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	// 客户端只看到 status=1
	list, total, err := h.survey.List(keyword, category, 1, page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// Detail GET /survey/view?id=
// @Tags 客户端-问卷
// @Summary 查看问卷详情
// @Param id query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /survey/view [get]
func (h *ClientSurveyHandler) Detail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	sv, err := h.survey.Get(uint(id))
	if err != nil {
		logger.Logger.Printf("[SurveyDetail] 问卷不存在 id=%d", id)
		response.Fail(c, "问卷不存在")
		return
	}
	if sv.Status != 1 {
		logger.Logger.Printf("[SurveyDetail] 问卷已停用 id=%d title=%s", id, sv.Title)
		response.Fail(c, "问卷已停用")
		return
	}
	// 登录可见 / 部门限定：检查用户登录
	if sv.Visibility == 1 || sv.Visibility == 2 {
		auth := c.GetHeader("Authorization")
		token := ""
		if len(auth) > 0 {
			token = string(auth)
		}
		if token == "" {
			logger.Logger.Printf("[SurveyDetail] 未登录 id=%d", id)
			response.Fail(c, "请先登录")
			return
		}
		rdKey := "user_token:a:" + token
		jsonStr, err := rd.RDB.Get(rd.Ctx, rdKey).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[SurveyDetail] token无效 id=%d", id)
			response.Fail(c, "请先登录")
			return
		}
		// 部门限定：校验用户部门
		if sv.Visibility == 2 && sv.DeptIDs != "" {
			var userInfo map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &userInfo)
			uid := uint(0)
			if id, ok := userInfo["id"].(float64); ok {
				uid = uint(id)
			}
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", uid).First(&ud)
			deptIds := strings.Split(sv.DeptIDs, ",")
			allowed := false
			for _, did := range deptIds {
				d, _ := strconv.Atoi(strings.TrimSpace(did))
				if uint(d) == ud.DeptID {
					allowed = true
					break
				}
			}
			if !allowed {
				logger.Logger.Printf("[SurveyDetail] 部门无权限 id=%d uid=%d deptId=%d", id, uid, ud.DeptID)
				response.Fail(c, "您不在该问卷的可见部门中")
				return
			}
		}
	}
	// 检查时间窗
	now := time.Now().UnixMilli()
	if sv.StartTime > 0 && now < sv.StartTime {
		logger.Logger.Printf("[SurveyDetail] 问卷未开始 id=%d startTime=%d", id, sv.StartTime)
		response.Fail(c, "问卷未开始")
		return
	}
	if sv.EndTime > 0 && now > sv.EndTime {
		logger.Logger.Printf("[SurveyDetail] 问卷已结束 id=%d endTime=%d", id, sv.EndTime)
		response.Fail(c, "问卷已结束")
		return
	}
	// 解析 schema 返回
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(sv.Schema), &schMap)
	// 解析 settings
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(sv.Settings), &settingsMap)
	// session 复用：前端传现有 session 则不再新生成
	session := c.Query("session")
	if session == "" {
		session = fmt.Sprintf("%x", time.Now().UnixNano()+rand.Int63())
	}
	redisKey := fmt.Sprintf("survey_session:%d:%s", sv.ID, session)
	var startAt int64
	exists, _ := rd.RDB.Exists(rd.Ctx, redisKey).Result()
	if exists == 0 {
		rd.RDB.Set(rd.Ctx, redisKey, now, 24*time.Hour)
		startAt = now
	} else {
		v, err := rd.RDB.Get(rd.Ctx, redisKey).Int64()
		if err == nil {
			startAt = v
		} else {
			startAt = now
		}
	}
	// 公开视图
	publicView := map[string]interface{}{
		"id":           sv.ID,
		"title":        sv.Title,
		"description":  sv.Desc,
		"category":     sv.Category,
		"cover":        sv.Cover,
		"visibility":   sv.Visibility,
		"anonymous":    sv.Anonymous,
		"allowMulti":   sv.AllowMulti,
		"startTime":    sv.StartTime,
		"endTime":      sv.EndTime,
		"maxResponse":  sv.MaxResponse,
		"showResult":   sv.ShowResult,
		"schema":       schMap,
		"settings":     settingsMap,
		"session":      session,
		"startAt":      startAt,
		"deptIds":      sv.DeptIDs,
	}
	logger.Logger.Printf("[SurveyDetail] 成功 id=%d title=%s", id, sv.Title)
	response.JSON(c, publicView)
}

// ApplyLogic POST /survey/apply
// 应用 schema 逻辑（计算值 + 显隐）
func (h *ClientSurveyHandler) ApplyLogic(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Answers  map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	sv, err := h.survey.Get(req.SurveyID)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	out, _ := h.survey.ApplyLogic(sv, req.Answers)
	response.JSON(c, map[string]interface{}{"answers": out})
}

// Validate POST /survey/validate
func (h *ClientSurveyHandler) Validate(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Answers  map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	sv, err := h.survey.Get(req.SurveyID)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	errs :=  service.ValidateAnswers(sv, req.Answers)
	response.JSON(c, map[string]interface{}{"errors": errs, "valid": len(errs) == 0})
}

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
func (h *ClientSurveyHandler) Submit(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID  uint                   `json:"surveyId"`
		Answers   map[string]interface{} `json:"answers"`
		Nickname  string                 `json:"nickname"`
		Session   string                 `json:"session"`
		StartTime int64                  `json:"startTime"`
		Device    string                 `json:"device"`
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
	if loginRequired2 {
		token := string(c.Request.Header.Peek("Authorization"))
		if token == "" {
			logger.Logger.Printf("[SurveySubmit] 需要登录 surveyId=%d", req.SurveyID)
			response.Fail(c, "请登录")
			return
		}
		_, prefix := tokenutil.GetTokenConfig("user")
		jsonStr, err := rd.RDB.Get(rd.Ctx, prefix+"a:"+token).Result()
		if err != nil || jsonStr == "" {
			logger.Logger.Printf("[SurveySubmit] 登录校验失败 surveyId=%d", req.SurveyID)
			response.Fail(c, "请登录")
			return
		}
	}
	// 如果没传 startTime，尝试从 Redis session 恢复
	if req.StartTime <= 0 && req.Session != "" {
		redisKey := fmt.Sprintf("survey_session:%d:%s", req.SurveyID, req.Session)
		v, err := rd.RDB.Get(rd.Ctx, redisKey).Int64()
		if err == nil {
			req.StartTime = v
			rd.RDB.Del(rd.Ctx, redisKey)
		}
	}
	// 用户 ID（公开链接可以匿名）
	uidVal, _ := c.Get("user_id")
	var uid uint
	if uidVal != nil {
		if v, ok := uidVal.(int64); ok {
			uid = uint(v)
		}
	}
	// IP
	ip := c.ClientIP()
	resp, err := h.responses.Submit(req.SurveyID, uid, req.Nickname, req.StartTime, req.Answers, ip, req.Device)
	if err != nil {
		logger.Logger.Printf("[SurveySubmit] 失败 surveyId=%d uid=%d err=%s ip=%s device=%s", req.SurveyID, uid, err.Error(), ip, req.Device)
		response.Fail(c, err.Error())
		return
	}
	logger.Logger.Printf("[SurveySubmit] 成功 surveyId=%d uid=%d respId=%d ip=%s device=%s", req.SurveyID, uid, resp.ID, ip, req.Device)
	response.JSON(c, map[string]interface{}{"id": resp.ID, "submitTime": resp.SubmitTime})
}

// MyResponses GET /survey/my_responses
// @Tags 客户端-问卷
// @Summary 我的答卷列表
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /survey/my_responses [get]
func (h *ClientSurveyHandler) MyResponses(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	uidVal, _ := c.Get("user_id")
	if uidVal == nil {
		response.Fail(c, "未登录")
		return
	}
	uid := uint(uidVal.(int64))
	uidStr := strconv.FormatUint(uint64(uid), 10)
	var list []model.SurveyResponse
	database.DB.Where("`survey_resp_user_id` = ? AND `survey_resp_status` = 1", uidStr).
		Order("`survey_resp_id` DESC").Limit(50).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list})
}

// MyResponseDetail GET /survey/my_response?id=
// @Tags 客户端-问卷
// @Summary 查看答卷详情
// @Param id query int true "答卷ID"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /survey/my_response [get]
func (h *ClientSurveyHandler) MyResponseDetail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	uidVal, _ := c.Get("user_id")
	if uidVal == nil {
		response.Fail(c, "未登录")
		return
	}
	uid := uint(uidVal.(int64))
	id, _ := strconv.Atoi(c.Query("id"))
	resp, err := h.responses.Get(uint(id))
	if err != nil {
		response.Fail(c, "答卷不存在")
		return
	}
	uidStr := strconv.FormatUint(uint64(uid), 10)
	if resp.UserID != uidStr {
		response.Fail(c, "无权查看")
		return
	}
	sv, _ := h.survey.Get(resp.SurveyID)
	answers := h.responses.ParseAnswers(resp)
	out := map[string]interface{}{"response": resp, "answers": answers}
	if sv != nil && sv.ShowResult == 1 {
		out["survey"] = sv
	}
	response.JSON(c, out)
}

// ==================== 通用表单工具（公开，无认证） ====================

// PublicValidate POST /survey/validate
// @Tags 客户端-表单工具
// @Summary 校验答案格式（通用）
// @Router /survey/validate [post]
func (h *ClientSurveyHandler) PublicValidate(_ context.Context, c *app.RequestContext) {
	var req struct {
		SurveyID uint                   `json:"surveyId"`
		Schema   string                 `json:"schema"`
		Answers  map[string]interface{} `json:"answers"`
		Device   string                 `json:"device"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	schema := req.Schema
	if req.SurveyID > 0 {
		var sv model.Survey
		if err := database.DB.Where("`survey_id` = ? AND `survey_status` = 1", req.SurveyID).First(&sv).Error; err != nil {
			logger.Logger.Printf("[SurveyValidate] 问卷不存在或未发布 surveyId=%d", req.SurveyID)
			response.Fail(c, "问卷不存在或未发布")
			return
		}
		// 回收上限
		if sv.MaxResponse > 0 {
			var count int64
			database.DB.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", req.SurveyID).
				Count(&count)
			if count >= int64(sv.MaxResponse) {
				logger.Logger.Printf("[SurveyValidate] 回收上限已满 surveyId=%d max=%d current=%d", req.SurveyID, sv.MaxResponse, count)
				response.JSON(c, map[string]interface{}{
					"ok": false, "errors": []map[string]string{{"questionId": "", "message": "回收上限已满"}},
				})
				return
			}
		}
		// 从 settings 解析 deviceLimit / ipLimit
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
		if deviceLimit > 0 && req.Device != "" {
			var devCount int64
			database.DB.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_device` = ? AND `survey_resp_status` = 1", req.SurveyID, req.Device).
				Count(&devCount)
			if devCount >= int64(deviceLimit) {
				logger.Logger.Printf("[SurveyValidate] 设备次数上限 surveyId=%d limit=%d current=%d device=%s", req.SurveyID, deviceLimit, devCount, req.Device)
				response.JSON(c, map[string]interface{}{
					"ok": false, "errors": []map[string]string{{"questionId": "", "message": "该设备答题次数已达上限"}},
				})
				return
			}
		}
		if ipLimit > 0 && clientIP != "" {
			var ipCount int64
			database.DB.Model(&model.SurveyResponse{}).
				Where("`survey_resp_survey_id` = ? AND `survey_resp_ip` = ? AND `survey_resp_status` = 1", req.SurveyID, clientIP).
				Count(&ipCount)
			if ipCount >= int64(ipLimit) {
				logger.Logger.Printf("[SurveyValidate] IP次数上限 surveyId=%d limit=%d current=%d ip=%s", req.SurveyID, ipLimit, ipCount, clientIP)
				response.JSON(c, map[string]interface{}{
					"ok": false, "errors": []map[string]string{{"questionId": "", "message": "该IP答题次数已达上限"}},
				})
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
	response.JSON(c, map[string]interface{}{"ok": len(errs) == 0, "errors": errs})
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
	response.JSON(c, map[string]interface{}{"answers": newAns, "states": states})
}
