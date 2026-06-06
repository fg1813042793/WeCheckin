package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	surveySvc "wecheckin-backend/backend/internal/survey/service"
	"wecheckin-backend/backend/pkg/response"
)

// NewClientSurveyHandler 客户端 survey handler（用户填问卷）
func NewClientSurveyHandler() *ClientSurveyHandler { return &ClientSurveyHandler{} }

type ClientSurveyHandler struct {
	survey    *surveySvc.SurveyService
	responses *surveySvc.ResponseService
}

func (h *ClientSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = surveySvc.NewSurveyService()
	}
	if h.responses == nil {
		h.responses = surveySvc.NewResponseService()
	}
}

// List GET /survey/list
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
// 返回问卷的 schema（不含正确答案和后端配置）
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
	// 公开视图（不返回后端配置）
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
func (h *ClientSurveyHandler) Submit(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		SurveyID  uint                   `json:"surveyId"`
		Answers   map[string]interface{} `json:"answers"`
		Nickname  string                 `json:"nickname"`
		StartTime int64                  `json:"startTime"`
		Device    string                 `json:"device"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
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
