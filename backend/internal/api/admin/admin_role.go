package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminRoleHandler struct{}

func NewAdminRoleHandler() *AdminRoleHandler { return &AdminRoleHandler{} }

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
