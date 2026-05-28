package api

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminEnrollHandler struct{}

func NewAdminEnrollHandler() *AdminEnrollHandler { return &AdminEnrollHandler{} }

// @Tags 打卡管理
// @Summary 获取打卡项目列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_list [get]
func (h *AdminEnrollHandler) GetAdminEnrollList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	keyword := c.Query("keyword")
	list, total, err := service.GetAdminEnrollList(keyword, page, size)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 打卡管理
// @Summary 新增打卡项目
// @Success 200 {object} response.Resp
// @Router /admin/enroll_insert [post]
func (h *AdminEnrollHandler) InsertEnroll(ctx context.Context, c *app.RequestContext) {
	response.Fail(c, "[打卡]该功能暂不开放，如有需要请加作者微信：cclinux0730")
}

// @Tags 打卡管理
// @Summary 获取打卡项目详情
// @Param id query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_detail [get]
func (h *AdminEnrollHandler) GetEnrollDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.GetEnrollDetail(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 打卡管理
// @Summary 编辑打卡项目
// @Success 200 {object} response.Resp
// @Router /admin/enroll_edit [post]
func (h *AdminEnrollHandler) EditEnroll(ctx context.Context, c *app.RequestContext) {
	response.Fail(c, "[打卡]该功能暂不开放，如有需要请加作者微信：cclinux0730")
}

// @Tags 打卡管理
// @Summary 更新打卡表单
// @Param id formData string true "项目ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_update_forms [post]
func (h *AdminEnrollHandler) UpdateEnrollForms(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	err := service.UpdateEnrollForms(id, forms)
	if err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 清除打卡全部数据
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_clear [post]
func (h *AdminEnrollHandler) ClearEnrollAll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.ClearEnrollAll(id)
	if err != nil {
		response.Fail(c, "清除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 删除打卡项目
// @Param id formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_del [post]
func (h *AdminEnrollHandler) DelEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelEnroll(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 打卡项目排序
// @Param id formData string true "项目ID"
// @Param sort formData string true "排序值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_sort [post]
func (h *AdminEnrollHandler) SortEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	sortStr := c.PostForm("sort")
	err := service.SortEnroll(id, sortStr)
	if err != nil {
		response.Fail(c, "排序失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 设置打卡推荐
// @Param id formData string true "项目ID"
// @Param vouch formData string true "推荐值"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_vouch [post]
func (h *AdminEnrollHandler) VouchEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	vouch, _ := strconv.Atoi(c.PostForm("vouch"))
	err := service.VouchEnroll(id, vouch)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 设置打卡状态
// @Param id formData string true "项目ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_status [post]
func (h *AdminEnrollHandler) StatusEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := service.StatusEnroll(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 获取打卡记录列表
// @Param enrollId query string true "项目ID"
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_list [get]
func (h *AdminEnrollHandler) GetEnrollJoinList(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	list, total, err := service.GetEnrollJoinList(enrollID, page, size)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 打卡管理
// @Summary 删除打卡记录
// @Param id formData string true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_del [post]
func (h *AdminEnrollHandler) DelEnrollJoin(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelEnrollJoin(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 打卡管理
// @Summary 获取打卡数据导出链接
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_get [get]
func (h *AdminEnrollHandler) EnrollJoinDataGet(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	data, err := service.GetEnrollJoinDataURL(enrollID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 打卡管理
// @Summary 导出打卡数据
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_export [get]
func (h *AdminEnrollHandler) EnrollJoinDataExport(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	data, err := service.ExportEnrollJoinDataExcel(enrollID)
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 打卡管理
// @Summary 删除打卡导出数据
// @Param enrollId formData string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_data_del [post]
func (h *AdminEnrollHandler) EnrollJoinDataDel(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	err := service.DeleteEnrollJoinDataExcel(enrollID)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
