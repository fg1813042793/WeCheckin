package admin

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminMenuHandler struct{}

func NewAdminMenuHandler() *AdminMenuHandler { return &AdminMenuHandler{} }

func (h *AdminMenuHandler) GetMenuTree(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetMenuTree()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminMenuHandler) GetMenuList(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetMenuList()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminMenuHandler) AddMenu(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	parentID, _ := strconv.Atoi(c.PostForm("parentId"))
	path := c.PostForm("path")
	perms := c.PostForm("perms")
	icon := c.PostForm("icon")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	mtype, _ := strconv.Atoi(c.PostForm("type"))
	if name == "" {
		response.Fail(c, "菜单名称不能为空")
		return
	}
	if err := service.AddMenu(name, uint(parentID), path, perms, icon, sort, mtype); err != nil {
		response.Fail(c, "添加失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminMenuHandler) EditMenu(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	name := c.PostForm("name")
	parentID, _ := strconv.Atoi(c.PostForm("parentId"))
	path := c.PostForm("path")
	perms := c.PostForm("perms")
	icon := c.PostForm("icon")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	mtype, _ := strconv.Atoi(c.PostForm("type"))
	if name == "" || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.EditMenu(uint(id), name, uint(parentID), path, perms, icon, sort, status, mtype); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminMenuHandler) DelMenu(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelMenu(uint(id)); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// GetAdminMenus returns the menu tree for the current admin (filtered by role)
func (h *AdminMenuHandler) GetAdminMenus(ctx context.Context, c *app.RequestContext) {
	adminVal, exists := c.Get("admin")
	if !exists {
		response.Fail(c, "未登录")
		return
	}
	admin := adminVal.(*model.Admin)
	data, err := service.GetAdminMenuTree(admin)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// GetAdminPerms returns the permission keys for the current admin
func (h *AdminMenuHandler) GetAdminPerms(ctx context.Context, c *app.RequestContext) {
	adminVal, exists := c.Get("admin")
	if !exists {
		response.Fail(c, "未登录")
		return
	}
	admin := adminVal.(*model.Admin)
	perms := service.GetAdminPerms(admin)
	response.JSON(c, perms)
}
