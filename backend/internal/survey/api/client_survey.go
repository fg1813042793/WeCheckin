package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	surveySvc "wecheckin-backend/backend/internal/survey/service"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/response"
)

type ClientSurveyHandler struct {
	survey    *surveySvc.SurveyService
	responses *surveySvc.ResponseService
}

func NewClientSurveyHandler() *ClientSurveyHandler { return &ClientSurveyHandler{} }

func (h *ClientSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = surveySvc.NewSurveyService()
	}
	if h.responses == nil {
		h.responses = surveySvc.NewResponseService()
	}
}

// List GET /survey/list
// @Tags 问卷-客户端
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
// @Tags 问卷-客户端
// @Summary 查看问卷详情
// @Param id query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /survey/view [get]
func (h *ClientSurveyHandler) Detail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	sv, err := h.survey.Get(uint(id))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	if sv.Status != 1 {
		response.Fail(c, "问卷已停用")
		return
	}
	// 检查时间窗
	now := time.Now().UnixMilli()
	if sv.StartTime > 0 && now < sv.StartTime {
		response.Fail(c, "问卷未开始")
		return
	}
	if sv.EndTime > 0 && now > sv.EndTime {
		response.Fail(c, "问卷已截止")
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
	exists, _ := rd.RDB.Exists(rd.Ctx, redisKey).Result()
	if exists == 0 {
		rd.RDB.Set(rd.Ctx, redisKey, now, 24*time.Hour)
	}
	// 公开视图
	publicView := map[string]interface{}{
		"id":           sv.ID,
		"title":        sv.Title,
		"description":  sv.Desc,
		"category":     sv.Category,
		"cover":        sv.Cover,
		"anonymous":    sv.Anonymous,
		"allowMulti":   sv.AllowMulti,
		"startTime":    sv.StartTime,
		"endTime":      sv.EndTime,
		"maxResponse":  sv.MaxResponse,
		"showResult":   sv.ShowResult,
		"schema":       schMap,
		"settings":     settingsMap,
		"session":      session,
	}
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
	errs := surveySvc.ValidateAnswers(sv, req.Answers)
	response.JSON(c, map[string]interface{}{"errors": errs, "valid": len(errs) == 0})
}

// Submit POST /survey/submit
// @Tags 问卷-客户端
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
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{"id": resp.ID, "submitTime": resp.SubmitTime})
}

// MyResponses GET /survey/my_responses
// @Tags 问卷-客户端
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
// @Tags 问卷-客户端
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
