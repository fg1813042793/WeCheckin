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
	examSvc "wecheckin-backend/backend/internal/exam/service"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminExamHandler struct {
	svc *examSvc.ExamService
}

func NewAdminExamHandler() *AdminExamHandler {
	return &AdminExamHandler{svc: examSvc.NewExamService()}
}

// @Tags 考试管理, 管理端 API
// @Summary 考试详情
// @Param id query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/detail [get]
func (h *AdminExamHandler) Detail(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "无效的考试ID")
		return
	}
	exam, err := h.svc.Get(uint(id))
	if err != nil {
		response.Fail(c, "考试不存在")
		return
	}
	var respCnt int64
	database.DB.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ?", id).Count(&respCnt)
	var rawSchema string
	database.DB.Model(&model.Exam{}).Select("exam_schema").Where("`exam_id` = ?", id).Scan(&rawSchema)
	survey := map[string]interface{}{
		"id":          exam.ID,
		"title":       exam.Title,
		"description": exam.Description,
		"category":    exam.Category,
		"tags":        exam.Tags,
		"visibility":  exam.Visibility,
		"allowMulti":  exam.AllowMulti,
		"anonymous":   exam.Anonymous,
		"showResult":  exam.ShowResult,
		"startTime":   exam.StartTime,
		"endTime":     exam.EndTime,
		"maxResponse": exam.MaxResponse,
		"duration":    exam.Duration,
		"maxAttempts": exam.MaxAttempts,
		"showScore":   exam.ShowScore,
		"status":      exam.Status,
		"deptIds":     exam.DeptIds,
		"mode":        exam.Mode,
		"createBy":    exam.CreateBy,
		"settings":    exam.Settings,
	}
	response.JSON(c, map[string]interface{}{"survey": survey, "responseCount": respCnt, "schema": rawSchema})
}

// @Tags 考试管理, 管理端 API
// @Summary 创建/更新考试
// @Param title formData string true "标题"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param tags formData string false "标签"
// @Param visibility formData int false "可见性"
// @Param allowMulti formData int false "允许多次"
// @Param anonymous formData int false "匿名"
// @Param showResult formData int false "显示结果"
// @Param startTime formData int false "开始时间"
// @Param endTime formData int false "结束时间"
// @Param maxResponse formData int false "最大答卷数"
// @Param duration formData int false "答题时长"
// @Param maxAttempts formData int false "最大次数"
// @Param showScore formData int false "显示分数"
// @Param status formData int false "状态"
// @Param schema formData string false "题目JSON"
// @Param deptIds formData string false "部门ID"
// @Param mode formData string false "模式"
// @Param settings formData string false "设置JSON"
// @Success 200 {object} response.Resp
// @Router /admin/exam/save [post]
func (h *AdminExamHandler) Save(_ context.Context, c *app.RequestContext) {
	type ExamSaveReq struct {
		ID          uint   `json:"id" form:"id"`
		Title       string `json:"title" form:"title"`
		Description string `json:"description" form:"description"`
		Category    string `json:"category" form:"category"`
		Tags        string `json:"tags" form:"tags"`
		Visibility  int    `json:"visibility" form:"visibility"`
		AllowMulti  int    `json:"allowMulti" form:"allowMulti"`
		Anonymous   int    `json:"anonymous" form:"anonymous"`
		ShowResult  int    `json:"showResult" form:"showResult"`
		StartTime   int64  `json:"startTime" form:"startTime"`
		EndTime     int64  `json:"endTime" form:"endTime"`
		MaxResponse int    `json:"maxResponse" form:"maxResponse"`
		Duration    int    `json:"duration" form:"duration"`
		MaxAttempts int    `json:"maxAttempts" form:"maxAttempts"`
		ShowScore   int    `json:"showScore" form:"showScore"`
		Status      int    `json:"status" form:"status"`
		Schema      string `json:"schema" form:"schema"`
		DeptIds     string `json:"deptIds" form:"deptIds"`
		Mode        string `json:"mode" form:"mode"`
		Settings    string `json:"settings" form:"settings"`
	}
	var req ExamSaveReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if req.Title == "" {
		response.Fail(c, "标题不能为空")
		return
	}
	if req.ID == 0 {
		var deptID uint
		var createBy uint
		if admin, ok := c.Get("admin"); ok {
			if a, ok := admin.(*model.Admin); ok {
				createBy = a.ID
				var adminDept model.AdminDept
				if err := database.DB.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
					deptID = adminDept.DeptID
				}
			}
		}
		exam := model.Exam{
			Title:       req.Title,
			Description: req.Description,
			Category:    req.Category,
			Tags:        req.Tags,
			Visibility:  req.Visibility,
			AllowMulti:  req.AllowMulti,
			Anonymous:   req.Anonymous,
			ShowResult:  req.ShowResult,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			MaxResponse: req.MaxResponse,
			Schema:      req.Schema,
			DeptIds:     req.DeptIds,
			Mode:        req.Mode,
			Settings:    req.Settings,
			Duration:    req.Duration,
			MaxAttempts: req.MaxAttempts,
			ShowScore:   req.ShowScore,
			Status:      req.Status,
			DeptID:      deptID,
			CreateBy:    createBy,
		}
		result, err := h.svc.Create(exam)
		if err != nil {
			response.Fail(c, "创建失败: "+err.Error())
			return
		}
		response.JSON(c, result)
	} else {
		updates := map[string]interface{}{
			"exam_title":        req.Title,
			"exam_desc":         req.Description,
			"exam_category":     req.Category,
			"exam_tags":         req.Tags,
			"exam_visibility":   req.Visibility,
			"exam_allow_multi":  req.AllowMulti,
			"exam_anonymous":    req.Anonymous,
			"exam_show_result":  req.ShowResult,
			"exam_start_time":   req.StartTime,
			"exam_end_time":     req.EndTime,
			"exam_max_response": req.MaxResponse,
			"exam_schema":       req.Schema,
			"exam_dept_ids":     req.DeptIds,
			"exam_mode":         req.Mode,
			"exam_settings":     req.Settings,
			"exam_duration":     req.Duration,
			"exam_max_attempts": req.MaxAttempts,
			"exam_show_score":   req.ShowScore,
			"exam_status":       req.Status,
		}
		if err := h.svc.Update(req.ID, updates); err != nil {
			response.Fail(c, "更新失败: "+err.Error())
			return
		}
		response.JSON(c, map[string]interface{}{"id": req.ID})
	}
}

// @Tags 考试管理, 管理端 API
// @Summary 考试列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态"
// @Success 200 {object} response.Resp
// @Router /admin/exam/list [get]
func (h *AdminExamHandler) List(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := h.svc.List(c.Query("keyword"), c.Query("category"), c.Query("status"), page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total, "page": page, "size": pageSize})
}

// @Tags 考试管理, 管理端 API
// @Summary 更新考试状态
// @Param id formData int true "考试ID"
// @Param status formData int true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/exam/status [post]
func (h *AdminExamHandler) Status(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if id <= 0 {
		response.Fail(c, "无效的考试ID")
		return
	}
	if err := h.svc.SetStatus(uint(id), status); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 考试管理, 管理端 API
// @Summary 删除考试
// @Param id formData int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/delete [post]
func (h *AdminExamHandler) Delete(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags 考试管理, 管理端 API
// @Summary 考试记录列表
// @Param examId query int true "考试ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_list [get]
func (h *AdminExamHandler) RecordList(_ context.Context, c *app.RequestContext) {
	examId, _ := strconv.Atoi(c.Query("examId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := h.svc.RecordList(examId, c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 考试管理, 管理端 API
// @Summary 考试记录详情
// @Param id query int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_detail [get]
func (h *AdminExamHandler) RecordDetail(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	data, err := h.svc.RecordDetail(uint(id))
	if err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	response.JSON(c, data)
}

// @Tags 考试管理, 管理端 API
// @Summary 删除考试记录
// @Param id formData int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_del [post]
func (h *AdminExamHandler) RecordDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	h.svc.RecordDelete(uint(id))
	response.JSON(c, nil)
}

// @Tags 考试管理, 管理端 API
// @Summary 批量删除考试记录
// @Param ids formData string true "逗号分隔的记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_batch_del [post]
func (h *AdminExamHandler) RecordBatchDel(_ context.Context, c *app.RequestContext) {
	ids := c.PostForm("ids")
	if ids == "" {
		response.Fail(c, "参数错误")
		return
	}
	h.svc.RecordBatchDelete(ids)
	response.JSON(c, nil)
}

// @Tags 考试管理, 管理端 API
// @Summary 考试统计
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/statistics [get]
func (h *AdminExamHandler) Statistics(_ context.Context, c *app.RequestContext) {
	examId, _ := strconv.Atoi(c.Query("examId"))
	if examId <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	data := h.svc.Statistics(examId)
	response.JSON(c, data)
}

// @Tags 考试管理, 管理端 API
// @Summary 上传考试资源
// @Param file formData file true "文件"
// @Param examId formData int true "考试ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_upload [post]
func (h *AdminExamHandler) ResourceUpload(_ context.Context, c *app.RequestContext) {
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
	examID, _ := strconv.Atoi(string(c.FormValue("examId")))
	resType := string(c.FormValue("resType"))
	if resType != "bg" && resType != "header" {
		response.Fail(c, "无效的资源类型")
		return
	}
	if examID <= 0 {
		response.Fail(c, "无效的考试ID")
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
	res := model.ExamResource{
		ExamID:   uint(examID),
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

// @Tags 考试管理, 管理端 API
// @Summary 查询考试资源列表
// @Param examId query int true "考试ID"
// @Param resType query string false "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_list [get]
func (h *AdminExamHandler) ResourceList(_ context.Context, c *app.RequestContext) {
	examID, _ := strconv.Atoi(c.Query("examId"))
	resType := c.Query("resType")
	if examID <= 0 {
		response.JSON(c, []model.ExamResource{})
		return
	}
	query := database.DB.Where("`exam_res_exam_id` = ?", examID)
	if resType != "" {
		query = query.Where("`exam_res_type` = ?", resType)
	}
	var list []model.ExamResource
	if err := query.Order("`exam_res_add_time` DESC").Find(&list).Error; err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, list)
}

// @Tags 考试管理, 管理端 API
// @Summary 删除考试资源
// @Param id formData int true "资源ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_delete [post]
func (h *AdminExamHandler) ResourceDelete(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(string(c.FormValue("id")))
	if id <= 0 {
		response.Fail(c, "无效的资源ID")
		return
	}
	var res model.ExamResource
	if err := database.DB.First(&res, id).Error; err != nil {
		response.Fail(c, "资源不存在")
		return
	}
	if res.Path != "" {
		os.Remove(res.Path)
	}
	if err := database.DB.Delete(&res).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ==================== Exam Question Bank ====================

// QuestionBankList GET /admin/exam/question_bank_list
// @Tags 考试管理, 管理端 API
// @Summary 获取考试题库列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_list [get]
func (h *AdminExamHandler) QuestionBankList(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	q := database.DB.Model(&model.ExamQuestion{})
	if keyword != "" {
		q = q.Where("`exam_q_title` LIKE ? OR `exam_q_type` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.ExamQuestion
	q.Order("`exam_q_add_time` DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, response.PageData{List: list, Total: total, Size: pageSize, Page: page})
}

// QuestionBankInsert POST /admin/exam/question_bank_insert
// @Tags 考试管理, 管理端 API
// @Summary 添加题目到考试题库
// @Param title body string true "题干"
// @Param type body string true "题型"
// @Param schema body string true "完整 formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_insert [post]
func (h *AdminExamHandler) QuestionBankInsert(_ context.Context, c *app.RequestContext) {
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
	createBy := uint(0)
	deptID := uint(0)
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			createBy = a.ID
			var adminDept model.AdminDept
			if err := database.DB.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
				deptID = adminDept.DeptID
			}
		}
	}
	q := model.ExamQuestion{
		Title:    r.Title,
		Type:     r.Type,
		Schema:   r.Schema,
		Category: r.Category,
		Tags:     r.Tags,
		Status:   1,
		DeptID:   deptID,
		CreateBy: createBy,
		AddTime:  database.Now(),
	}
	if err := database.DB.Create(&q).Error; err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, q)
}

// QuestionBankEdit POST /admin/exam/question_bank_edit
// @Tags 考试管理, 管理端 API
// @Summary 编辑题库题目
// @Param id body uint true "题目ID"
// @Param title body string false "题干"
// @Param type body string false "题型"
// @Param schema body string false "formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_edit [post]
func (h *AdminExamHandler) QuestionBankEdit(_ context.Context, c *app.RequestContext) {
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
	if err := database.DB.Model(&model.ExamQuestion{}).Where("`exam_q_id` = ?", r.ID).Updates(map[string]interface{}{
		"exam_q_title":    r.Title,
		"exam_q_type":     r.Type,
		"exam_q_schema":   r.Schema,
		"exam_q_category": r.Category,
		"exam_q_tags":     r.Tags,
	}).Error; err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankDel POST /admin/exam/question_bank_del
// @Tags 考试管理, 管理端 API
// @Summary 删除题库题目
// @Param id formData int true "题目ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_del [post]
func (h *AdminExamHandler) QuestionBankDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_q_id` = ?", id).Delete(&model.ExamQuestion{}).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankCategories GET /admin/exam/question_bank_categories
// @Tags 考试管理, 管理端 API
// @Summary 获取考试题库所有分类
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_categories [get]
func (h *AdminExamHandler) QuestionBankCategories(_ context.Context, c *app.RequestContext) {
	var categories []string
	database.DB.Model(&model.ExamQuestion{}).
		Where("`exam_q_category` != '' AND `exam_q_category` IS NOT NULL").
		Select("DISTINCT `exam_q_category`").
		Order("`exam_q_category` ASC").
		Pluck("`exam_q_category`", &categories)
	response.JSON(c, categories)
}
