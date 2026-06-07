package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	_ "wecheckin-backend/backend/internal/formkit/question/builtin" // 注册 24 个内置题型
	"wecheckin-backend/backend/internal/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
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
// @Tags 问卷管理
// @Summary 问卷列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_list [get]
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
// @Tags 问卷管理
// @Summary 问卷详情
// @Param id query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_detail [get]
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
// @Tags 问卷管理
// @Summary 创建问卷
// @Param survey body model.Survey true "问卷数据"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_insert [post]
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
// @Tags 问卷管理
// @Summary 编辑问卷
// @Param survey body model.Survey true "问卷数据（需包含ID）"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_edit [post]
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
// @Tags 问卷管理
// @Summary 删除问卷
// @Param id formData int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_del [post]
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
// @Tags 问卷管理
// @Summary 更新问卷状态
// @Param id formData int true "问卷ID"
// @Param status formData int true "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_status [post]
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
// @Tags 问卷管理
// @Summary 复制问卷
// @Param id formData int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_copy [post]
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
// @Tags 问卷管理
// @Summary 答卷列表
// @Param surveyId query int true "问卷ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_list [get]
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
// @Tags 问卷管理
// @Summary 答卷详情
// @Param id query int true "答卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_detail [get]
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
// @Tags 问卷管理
// @Summary 删除答卷
// @Param id formData int true "答卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_del [post]
func (h *AdminSurveyHandler) ResponseDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_resp_id` = ?", id).Delete(&model.SurveyResponse{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ResponseExport GET /admin/survey/response_export?surveyId= → CSV
// @Tags 问卷管理
// @Summary 导出答卷CSV
// @Param surveyId query int true "问卷ID"
// @Success 200 {file} string
// @Router /admin/survey/response_export [get]
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
// @Tags 问卷管理
// @Summary 问卷统计
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/statistic [get]
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
// @Tags 问卷管理
// @Summary 渠道列表
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/channel_list [get]
func (h *AdminSurveyHandler) ChannelList(_ context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	var list []model.SurveyChannel
	database.DB.Where("`survey_ch_survey_id` = ?", surveyID).Order("`survey_ch_id` DESC").Find(&list)
	response.JSON(c, map[string]interface{}{"list": list})
}

// ChannelInsert POST /admin/survey/channel_insert
// @Tags 问卷管理
// @Summary 创建渠道
// @Param channel body model.SurveyChannel true "渠道数据"
// @Success 200 {object} response.Resp
// @Router /admin/survey/channel_insert [post]
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
// @Tags 问卷管理
// @Summary 删除渠道
// @Param id formData int true "渠道ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/channel_del [post]
func (h *AdminSurveyHandler) ChannelDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_ch_id` = ?", id).Delete(&model.SurveyChannel{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ResourceUpload POST /admin/survey/resource_upload
// @Tags 问卷管理
// @Summary 上传问卷资源（背景图/页眉图）
// @Param file formData file true "文件"
// @Param surveyId formData int true "问卷ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_upload [post]
func (h *AdminSurveyHandler) ResourceUpload(_ context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp")
		return
	}
	if file.Size > 20*1024*1024 {
		response.Fail(c, "上传文件过大，最大20MB")
		return
	}
	surveyID, _ := strconv.Atoi(string(c.FormValue("surveyId")))
	resType := string(c.FormValue("resType"))
	if resType != "bg" && resType != "header" {
		response.Fail(c, "无效的资源类型")
		return
	}
	if surveyID <= 0 {
		response.Fail(c, "无效的问卷ID")
		return
	}

	uploadDir := "./uploads"
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	saveDir := filepath.Join(uploadDir, dateDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		response.Fail(c, "创建目录失败")
		return
	}
	filename := fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(file.Filename))
	dst := filepath.Join(saveDir, filename)

	src, err := file.Open()
	if err != nil {
		response.Fail(c, "上传失败")
		return
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		response.Fail(c, "上传失败")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		response.Fail(c, "上传失败")
		return
	}

	relPath := dateDir + "/" + filename
	absUpload, _ := filepath.Abs(uploadDir)
	relFile := "/uploads/" + relPath

	domain := service.GetStaticDomain()
	res := model.SurveyResource{
		SurveyID: uint(surveyID),
		Type:     resType,
		URL:      relFile,
		Filename: filename,
		Path:     filepath.Join(absUpload, relPath),
		Domain:   domain,
		AddTime:  now.UnixMilli(),
	}
	if err := database.DB.Create(&res).Error; err != nil {
		response.Fail(c, "保存记录失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]any{
		"id":       res.ID,
		"url":      relFile,
		"filename": filename,
		"path":     filepath.Join(absUpload, relPath),
		"domain":   domain,
		"type":     resType,
	})
}

// ResourceList GET /admin/survey/resource_list
// @Tags 问卷管理
// @Summary 查询问卷资源列表
// @Param surveyId query int true "问卷ID"
// @Param resType query string false "资源类型: bg/header，为空则返回全部"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_list [get]
func (h *AdminSurveyHandler) ResourceList(_ context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	resType := c.Query("resType")
	if surveyID <= 0 {
		response.Fail(c, "无效的问卷ID")
		return
	}
	query := database.DB.Where("`survey_res_survey_id` = ?", surveyID)
	if resType != "" {
		query = query.Where("`survey_res_type` = ?", resType)
	}
	var list []model.SurveyResource
	if err := query.Order("`survey_res_add_time` DESC").Find(&list).Error; err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, list)
}

// ResourceDelete POST /admin/survey/resource_delete
// @Tags 问卷管理
// @Summary 删除问卷资源
// @Param id formData int true "资源ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_delete [post]
func (h *AdminSurveyHandler) ResourceDelete(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(string(c.FormValue("id")))
	if id <= 0 {
		response.Fail(c, "无效的资源ID")
		return
	}
	var res model.SurveyResource
	if err := database.DB.First(&res, id).Error; err != nil {
		response.Fail(c, "资源不存在")
		return
	}
	// 删除物理文件
	if res.Path != "" {
		os.Remove(res.Path)
	}
	// 删除数据库记录
	if err := database.DB.Delete(&res).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// QuestionBankList GET /admin/survey/question_bank_list
func (h *AdminSurveyHandler) QuestionBankList(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	q := database.DB.Model(&model.SurveyQuestion{})
	if keyword != "" {
		q = q.Where("`survey_q_title` LIKE ? OR `survey_q_type` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.SurveyQuestion
	q.Order("`survey_q_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// QuestionBankInsert POST /admin/survey/question_bank_insert
func (h *AdminSurveyHandler) QuestionBankInsert(_ context.Context, c *app.RequestContext) {
	type req struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		Schema   string `json:"schema"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	var r req
	if err := c.BindAndValidate(&r); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if r.Title == "" {
		response.Fail(c, "标题不能为空")
		return
	}
	q := model.SurveyQuestion{
		Title:    r.Title,
		Type:     r.Type,
		Schema:   r.Schema,
		Category: r.Category,
		Tags:     r.Tags,
		Status:   1,
		AddTime:  database.Now(),
	}
	if err := database.DB.Create(&q).Error; err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, q)
}

// QuestionBankEdit POST /admin/survey/question_bank_edit
func (h *AdminSurveyHandler) QuestionBankEdit(_ context.Context, c *app.RequestContext) {
	type req struct {
		ID       uint   `json:"id"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Schema   string `json:"schema"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	var r req
	if err := c.BindAndValidate(&r); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if err := database.DB.Model(&model.SurveyQuestion{}).Where("`survey_q_id` = ?", r.ID).Updates(map[string]interface{}{
		"survey_q_title":    r.Title,
		"survey_q_type":     r.Type,
		"survey_q_schema":   r.Schema,
		"survey_q_category": r.Category,
		"survey_q_tags":     r.Tags,
	}).Error; err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankDel POST /admin/survey/question_bank_del
func (h *AdminSurveyHandler) QuestionBankDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_q_id` = ?", id).Delete(&model.SurveyQuestion{}).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
