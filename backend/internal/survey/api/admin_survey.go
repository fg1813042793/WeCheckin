package api

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	_ "wecheckin-backend/backend/internal/formkit/question/builtin" // 注册 24 个内置题型
	"wecheckin-backend/backend/internal/formkit/report"
	"wecheckin-backend/backend/internal/model"
	surveySvc "wecheckin-backend/backend/internal/survey/service"
	"wecheckin-backend/backend/pkg/response"
)

// NewAdminSurveyHandler admin 端 survey handler
func NewAdminSurveyHandler() *AdminSurveyHandler { return &AdminSurveyHandler{} }

type AdminSurveyHandler struct {
	survey    *surveySvc.SurveyService
	responses *surveySvc.ResponseService
}

func (h *AdminSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = surveySvc.NewSurveyService()
	}
	if h.responses == nil {
		h.responses = surveySvc.NewResponseService()
	}
}

// ==================== Survey CRUD ====================

// List GET /admin/survey/survey_list
func (h *AdminSurveyHandler) List(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	status, _ := strconv.Atoi(c.Query("status"))
	list, total, err := h.survey.List(keyword, category, status, page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	// 批量查询答卷数（避免 N+1）
	var ids []uint
	for _, sv := range list {
		ids = append(ids, sv.ID)
	}
	type RespCount struct {
		SurveyID uint `gorm:"column:survey_resp_survey_id"`
		Count    int  `gorm:"column:cnt"`
	}
	var counts []RespCount
	if len(ids) > 0 {
		database.DB.Model(&model.SurveyResponse{}).
			Select("`survey_resp_survey_id`, COUNT(*) AS cnt").
			Where("`survey_resp_survey_id` IN ?", ids).
			Group("`survey_resp_survey_id`").
			Scan(&counts)
	}
	countMap := make(map[uint]int)
	for _, c := range counts {
		countMap[c.SurveyID] = c.Count
	}
	type SurveyWithCount struct {
		model.Survey
		ResponseCount int `json:"responseCount"`
	}
	var out []SurveyWithCount
	for _, sv := range list {
		out = append(out, SurveyWithCount{Survey: sv, ResponseCount: countMap[sv.ID]})
	}
	response.JSON(c, map[string]interface{}{"list": out, "total": total, "page": page, "size": pageSize})
}

// Detail GET /admin/survey/survey_detail?id=
func (h *AdminSurveyHandler) Detail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	sv, err := h.survey.Get(uint(id))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	// 答题统计
	var respCnt int64
	database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", sv.ID).Count(&respCnt)
	// 单独读取 schema（因 model json:"-" 不自动序列化）
	var rawSchema string
	database.DB.Model(&model.Survey{}).Select("survey_schema").Where("`survey_id` = ?", id).Scan(&rawSchema)
	response.JSON(c, map[string]interface{}{"survey": sv, "responseCount": respCnt, "schema": rawSchema})
}

// Insert POST /admin/survey/survey_insert
func (h *AdminSurveyHandler) Insert(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var sv model.Survey
	if err := c.BindAndValidate(&sv); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if err := h.survey.Create(&sv); err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, sv)
}

// Edit POST /admin/survey/survey_edit
func (h *AdminSurveyHandler) Edit(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	var sv model.Survey
	if err := c.BindAndValidate(&sv); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if err := h.survey.Update(&sv); err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Del POST /admin/survey/survey_del
func (h *AdminSurveyHandler) Del(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.survey.Delete(uint(id)); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Status POST /admin/survey/survey_status
func (h *AdminSurveyHandler) Status(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if err := database.DB.Model(&model.Survey{}).Where("`survey_id` = ?", id).
		Update("survey_status", status).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Copy POST /admin/survey/survey_copy
func (h *AdminSurveyHandler) Copy(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.PostForm("id"))
	sv, err := h.survey.Get(uint(id))
	if err != nil {
		response.Fail(c, "原问卷不存在")
		return
	}
	now := time.Now().UnixMilli()
	newSv := *sv
	newSv.ID = 0
	newSv.Title = sv.Title + " (副本)"
	newSv.Status = 0
	newSv.AddTime = now
	newSv.EditTime = now
	if err := database.DB.Create(&newSv).Error; err != nil {
		response.Fail(c, "复制失败: "+err.Error())
		return
	}
	response.JSON(c, newSv)
}

// ==================== Response ====================

// ResponseList GET /admin/survey/response_list?surveyId=
func (h *AdminSurveyHandler) ResponseList(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	list, total, err := h.responses.List(uint(surveyID), page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// ResponseDetail GET /admin/survey/response_detail?id=
func (h *AdminSurveyHandler) ResponseDetail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	resp, err := h.responses.Get(uint(id))
	if err != nil {
		response.Fail(c, "答卷不存在")
		return
	}
	sv, _ := h.survey.Get(resp.SurveyID)
	answers := h.responses.ParseAnswers(resp)
	response.JSON(c, map[string]interface{}{
		"response": resp,
		"survey":   sv,
		"answers":  answers,
	})
}

// ResponseDel POST /admin/survey/response_del
func (h *AdminSurveyHandler) ResponseDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_resp_id` = ?", id).Delete(&model.SurveyResponse{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ResponseExport GET /admin/survey/response_export?surveyId= → CSV
func (h *AdminSurveyHandler) ResponseExport(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	var list []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ?", surveyID).Order("`survey_resp_id` ASC").Find(&list)
	items := make([]report.AnswerItem, len(list))
	for i, r := range list {
		items[i] = report.AnswerItem{UserID: r.UserID, AddTime: time.UnixMilli(r.AddTime).Format("2006-01-02 15:04:05"), Forms: r.Answers}
	}
	tbl, err := report.RenderAnswers(sv.Schema, items)
	if err != nil {
		response.Fail(c, "导出失败: "+err.Error())
		return
	}
	csvData := report.ToCSV(tbl)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=survey_"+strconv.Itoa(surveyID)+".csv")
	c.Write(csvData)
}

// ==================== Statistic ====================

// Statistic GET /admin/survey/statistic?surveyId=
// 聚合：总数/今日/近 7 天 + 各题选项分布
func (h *AdminSurveyHandler) Statistic(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	// 总数
	var total int64
	database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID).Count(&total)
	// 今日
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	var todayCnt int64
	database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ?", surveyID, dayStart).Count(&todayCnt)
	// 近 7 天按日
	daily := make([]map[string]interface{}, 7)
	for i := 0; i < 7; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		next := day.Add(24 * time.Hour)
		var c int64
		database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ? AND `survey_resp_add_time` < ?", surveyID, day.UnixMilli(), next.UnixMilli()).Count(&c)
		daily[6-i] = map[string]interface{}{
			"date":  day.Format("01-02"),
			"count": c,
		}
	}
	// 设备分布
	var mobileCnt, pcCnt int64
	database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_device` LIKE ?", surveyID, "%Mobile%").Count(&mobileCnt)
	// 有 device 但不含 Mobile 的算 PC
	database.DB.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_device` != '' AND `survey_resp_device` NOT LIKE ?", surveyID, "%Mobile%").Count(&pcCnt)

	// 每题统计
	var allResp []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ?", surveyID).Find(&allResp)
	items := make([]report.AnswerItem, len(allResp))
	for i, r := range allResp {
		items[i] = report.AnswerItem{Forms: r.Answers}
	}
	fieldStats := report.FieldStats(sv.Schema, items)

	response.JSON(c, map[string]interface{}{
		"survey":      sv,
		"total":       total,
		"todayCount":  todayCnt,
		"daily":       daily,
		"deviceStat":  map[string]int64{"mobile": mobileCnt, "pc": pcCnt},
		"fieldStats":  fieldStats,
		"viewUrl":     "/admin/survey/response_list?surveyId=" + c.Query("surveyId"),
	})
}

// ==================== Channel ====================

// ChannelList GET /admin/survey/channel_list?surveyId=
func (h *AdminSurveyHandler) ChannelList(_ context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	var list []model.SurveyChannel
	database.DB.Where("`survey_ch_survey_id` = ?", surveyID).Order("`survey_ch_id` DESC").Find(&list)
	response.JSON(c, map[string]interface{}{"list": list})
}

// ChannelInsert POST /admin/survey/channel_insert
func (h *AdminSurveyHandler) ChannelInsert(_ context.Context, c *app.RequestContext) {
	var ch model.SurveyChannel
	if err := c.BindAndValidate(&ch); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	ch.AddTime = time.Now().UnixMilli()
	if err := database.DB.Create(&ch).Error; err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, ch)
}

// ChannelDel POST /admin/survey/channel_del
func (h *AdminSurveyHandler) ChannelDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_ch_id` = ?", id).Delete(&model.SurveyChannel{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
