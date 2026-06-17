package api

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	examSvc "wecheckin-backend/backend/internal/exam/service"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type AdminExamHandler struct {
	svc *examSvc.ExamService
}

func NewAdminExamHandler() *AdminExamHandler {
	return &AdminExamHandler{svc: examSvc.NewExamService()}
}

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
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

// @Tags 考试管理
// @Summary 删除考试记录
// @Param id formData int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_del [post]
func (h *AdminExamHandler) RecordDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	h.svc.RecordDelete(uint(id))
	response.JSON(c, nil)
}

// @Tags 考试管理
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

// @Tags 考试管理
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
