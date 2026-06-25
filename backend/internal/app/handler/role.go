package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/app/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminRoleHandler struct{}

func NewAdminRoleHandler() *AdminRoleHandler { return &AdminRoleHandler{} }

// @Tags PC端-角色管理
// @Summary 获取角色列表
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /admin/role/list [get]
func (h *AdminRoleHandler) GetRoleList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	data, err := service.GetRoleList(admin.ID, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-角色管理
// @Summary 新增角色
// @Param name formData string true "角色名称"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Param dataScope formData int false "数据权限范围(1=全部 2=自定义)"
// @Param menuIds formData string false "菜单ID列表(逗号分隔)"
// @Param deptIds formData string false "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/role/add [post]
func (h *AdminRoleHandler) AddRole(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	dataScope, _ := strconv.Atoi(c.PostForm("dataScope"))
	if dataScope == 0 {
		dataScope = 1
	}
	if name == "" {
		response.Fail(c, "角色名称不能为空")
		return
	}
	roleID, err := service.AddRole(name, remark, c.ClientIP(), sort, dataScope)
	if err != nil {
		response.Fail(c, "添加失败")
		return
	}
	// Save menu assignments
	menuIDsStr := c.PostForm("menuIds")
	if menuIDsStr != "" {
		var menuIDs []uint
		for _, s := range strings.Split(menuIDsStr, ",") {
			s = strings.TrimSpace(s)
			if mid, err := strconv.Atoi(s); err == nil && mid > 0 {
				menuIDs = append(menuIDs, uint(mid))
			}
		}
		service.SetRoleMenus(roleID, menuIDs)
	}
	// Save dept assignments (for custom data scope)
	deptIDsStr := c.PostForm("deptIds")
	if deptIDsStr != "" {
		var deptIDs []uint
		for _, s := range strings.Split(deptIDsStr, ",") {
			s = strings.TrimSpace(s)
			if did, err := strconv.Atoi(s); err == nil && did > 0 {
				deptIDs = append(deptIDs, uint(did))
			}
		}
		service.SetRoleDepts(roleID, deptIDs)
	}
	response.JSON(c, nil)
}

// @Tags PC端-角色管理
// @Summary 编辑角色
// @Param id formData string true "角色ID"
// @Param name formData string true "角色名称"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Param status formData int false "状态(1=启用 0=禁用)"
// @Param dataScope formData int false "数据权限范围"
// @Param menuIds formData string false "菜单ID列表(逗号分隔)"
// @Param deptIds formData string false "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/role/edit [post]
func (h *AdminRoleHandler) EditRole(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	name := c.PostForm("name")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	dataScope, _ := strconv.Atoi(c.PostForm("dataScope"))
	if dataScope == 0 {
		dataScope = 1
	}
	if name == "" || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.EditRole(uint(id), name, remark, c.ClientIP(), sort, status, dataScope); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	// Save menu assignments
	menuIDsStr := c.PostForm("menuIds")
	if menuIDsStr != "" {
		var menuIDs []uint
		for _, s := range strings.Split(menuIDsStr, ",") {
			s = strings.TrimSpace(s)
			if mid, err := strconv.Atoi(s); err == nil && mid > 0 {
				menuIDs = append(menuIDs, uint(mid))
			}
		}
		service.SetRoleMenus(uint(id), menuIDs)
	}
	// Save dept assignments (for custom data scope)
	deptIDsStr := c.PostForm("deptIds")
	if deptIDsStr != "" {
		var deptIDs []uint
		for _, s := range strings.Split(deptIDsStr, ",") {
			s = strings.TrimSpace(s)
			if did, err := strconv.Atoi(s); err == nil && did > 0 {
				deptIDs = append(deptIDs, uint(did))
			}
		}
		service.SetRoleDepts(uint(id), deptIDs)
	}
	response.JSON(c, nil)
}

// @Tags PC端-角色管理
// @Summary 删除角色
// @Param id formData string true "角色ID"
// @Success 200 {object} response.Resp
// @Router /admin/role/del [post]
func (h *AdminRoleHandler) DelRole(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelRole(uint(id)); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-角色管理
// @Summary 批量删除角色
// @Param ids formData string true "角色ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/role/dels [post]
func (h *AdminRoleHandler) DelRoles(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	var ids []uint
	for _, s := range strings.Split(idsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(s))
		if err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.DelRoles(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
