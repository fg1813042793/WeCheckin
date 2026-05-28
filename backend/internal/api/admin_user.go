package api

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminUserHandler struct{}

func NewAdminUserHandler() *AdminUserHandler { return &AdminUserHandler{} }

// @Tags 用户管理
// @Summary 获取用户列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/user_list [get]
func (h *AdminUserHandler) GetUserList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	keyword := c.Query("keyword")
	list, total, err := service.GetUserList(keyword, page, size)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 用户管理
// @Summary 获取用户详情
// @Param openid query string true "用户OpenID"
// @Success 200 {object} response.Resp
// @Router /admin/user_detail [get]
func (h *AdminUserHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	openID := c.Query("openid")
	data, err := service.GetUserByOpenID(openID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 删除用户
// @Param id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/user_del [post]
func (h *AdminUserHandler) DelUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelUser(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 设置用户状态
// @Param id formData string true "用户ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/user_status [post]
func (h *AdminUserHandler) StatusUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := service.StatusUser(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 获取用户数据导出链接
// @Success 200 {object} response.Resp
// @Router /admin/user_data_get [get]
func (h *AdminUserHandler) UserDataGet(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetUserDataURL()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 导出用户数据
// @Success 200 {object} response.Resp
// @Router /admin/user_data_export [get]
func (h *AdminUserHandler) UserDataExport(ctx context.Context, c *app.RequestContext) {
	data, err := service.ExportUserDataExcel()
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 删除用户数据
// @Success 200 {object} response.Resp
// @Router /admin/user_data_del [post]
func (h *AdminUserHandler) UserDataDel(ctx context.Context, c *app.RequestContext) {
	err := service.DeleteUserDataExcel()
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
