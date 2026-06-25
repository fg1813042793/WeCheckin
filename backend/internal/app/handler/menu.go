package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/app/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminMenuHandler struct{}

func NewAdminMenuHandler() *AdminMenuHandler { return &AdminMenuHandler{} }

// @Tags PC端-菜单管理
// @Summary 获取菜单树
// @Success 200 {object} response.Resp
// @Router /admin/menu/tree [get]
func (h *AdminMenuHandler) GetMenuTree(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetMenuTree()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-菜单管理
// @Summary 获取菜单列表
// @Success 200 {object} response.Resp
// @Router /admin/menu/list [get]
func (h *AdminMenuHandler) GetMenuList(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetMenuList()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-菜单管理
// @Summary 新增菜单
// @Param name formData string true "菜单名称"
// @Param parentId formData int false "父菜单ID"
// @Param path formData string false "路由路径"
// @Param perms formData string false "权限标识"
// @Param icon formData string false "图标"
// @Param sort formData int false "排序"
// @Param type formData int false "菜单类型"
// @Success 200 {object} response.Resp
// @Router /admin/menu/add [post]
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

// @Tags PC端-菜单管理
// @Summary 编辑菜单
// @Param id formData string true "菜单ID"
// @Param name formData string true "菜单名称"
// @Param parentId formData int false "父菜单ID"
// @Param path formData string false "路由路径"
// @Param perms formData string false "权限标识"
// @Param icon formData string false "图标"
// @Param sort formData int false "排序"
// @Param status formData int false "状态(1=启用 0=禁用)"
// @Param type formData int false "菜单类型"
// @Success 200 {object} response.Resp
// @Router /admin/menu/edit [post]
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

// @Tags PC端-菜单管理
// @Summary 删除菜单
// @Param id formData string true "菜单ID"
// @Success 200 {object} response.Resp
// @Router /admin/menu/del [post]
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

// @Tags PC端-菜单管理
// @Summary 获取当前管理员的菜单树
// @Success 200 {object} response.Resp
// @Router /admin/user/menus [get]
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

// @Tags PC端-菜单管理
// @Summary 获取当前管理员的权限标识
// @Success 200 {object} response.Resp
// @Router /admin/user/perms [get]
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
