package exam

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-考试管理
// @Summary 考试记录列表
// @Param examId query int true "考试ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_list [get]
func (h *AdminExamHandler) RecordList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	examId, _ := strconv.Atoi(c.Query("examId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := h.svc.RecordListForAdminContext(ctx, examId, c.Query("keyword"), page, pageSize, admin.ID)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, examRecordListResponse{List: list, Total: total})
}

// @Tags PC端-考试管理
// @Summary 考试记录详情
// @Param id query int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_detail [get]
func (h *AdminExamHandler) RecordDetail(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	data, err := h.svc.RecordDetailForAdminContext(ctx, uint(id), admin.ID)
	if err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-考试管理
// @Summary 删除考试记录
// @Param id formData int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_del [post]
func (h *AdminExamHandler) RecordDel(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.svc.RecordDeleteForAdminContext(ctx, uint(id), admin.ID); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-考试管理
// @Summary 批量删除考试记录
// @Param ids formData string true "逗号分隔的记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/record_batch_del [post]
func (h *AdminExamHandler) RecordBatchDel(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	ids := c.PostForm("ids")
	if ids == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.svc.RecordBatchDeleteForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-考试管理
// @Summary 考试统计
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/statistics [get]
func (h *AdminExamHandler) Statistics(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	examId, _ := strconv.Atoi(c.Query("examId"))
	if examId <= 0 {
		response.Fail(c, "参数错误")
		return
	}
	data, err := h.svc.StatisticsForAdminContext(ctx, examId, admin.ID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, data)
}
