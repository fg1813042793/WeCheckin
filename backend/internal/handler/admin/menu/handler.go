package menu

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	menuservice "wecheckin/backend/internal/service/admin/menu"
	"wecheckin/backend/pkg/response"
)

type AdminMenuHandler struct{}

func NewAdminMenuHandler() *AdminMenuHandler { return &AdminMenuHandler{} }

// @Tags PC端-菜单管理
// @Summary 获取当前管理员的菜单树
// @Success 200 {object} response.Resp
func (h *AdminMenuHandler) GetAdminMenus(ctx context.Context, c *app.RequestContext) {
	adminVal, exists := c.Get("admin")
	if !exists {
		response.Fail(c, "未登录")
		return
	}
	admin := adminVal.(*model.Admin)
	data, err := menuservice.GetAdminMenuTree(admin)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-菜单管理
// @Summary 获取当前管理员的权限标识
// @Success 200 {object} response.Resp
func (h *AdminMenuHandler) GetAdminPerms(ctx context.Context, c *app.RequestContext) {
	adminVal, exists := c.Get("admin")
	if !exists {
		response.Fail(c, "未登录")
		return
	}
	admin := adminVal.(*model.Admin)
	perms := menuservice.GetAdminPerms(admin)
	response.JSON(c, perms)
}
