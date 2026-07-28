package permission

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	adminpermission "wecheckin-backend/backend/internal/app/service/adminpermission"
	"wecheckin-backend/backend/pkg/response"
)

type AdminPermissionHandler struct{}

func NewAdminPermissionHandler() *AdminPermissionHandler { return &AdminPermissionHandler{} }

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/permissions/tree
// @Security AdminToken
// @Param platform query string false "平台"
// @Param types query string false "权限类型，逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/tree [get]
func (h *AdminPermissionHandler) GetPermissionTree(ctx context.Context, c *app.RequestContext) {
	data, err := adminpermission.TreeContext(ctx, c.Query("platform"), c.Query("types"))
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags API v2-后台管理
// @Summary 查询 /api/v2/admin/permissions
// @Security AdminToken
// @Param platform query string false "平台"
// @Param types query string false "权限类型，逗号分隔"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [get]
func (h *AdminPermissionHandler) GetPermissionList(ctx context.Context, c *app.RequestContext) {
	data, err := adminpermission.ListContext(ctx, c.Query("platform"), c.Query("types"))
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags API v2-后台管理
// @Summary 提交 /api/v2/admin/permissions
// @Security AdminToken
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions [post]
func (h *AdminPermissionHandler) AddPermission(ctx context.Context, c *app.RequestContext) {
	req := permissionRequestFromForm(c)
	if err := adminpermission.AddContext(ctx, req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags API v2-后台管理
// @Summary 更新 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Param key path string true "权限编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [put]
func (h *AdminPermissionHandler) EditPermission(ctx context.Context, c *app.RequestContext) {
	key := c.PostForm("key")
	req := permissionRequestFromForm(c)
	if err := adminpermission.EditContext(ctx, key, req); err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags API v2-后台管理
// @Summary 删除 /api/v2/admin/permissions/{key}
// @Security AdminToken
// @Param key path string true "权限编码"
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/permissions/{key} [delete]
func (h *AdminPermissionHandler) DelPermission(ctx context.Context, c *app.RequestContext) {
	if err := adminpermission.DeleteContext(ctx, c.PostForm("key")); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func permissionRequestFromForm(c *app.RequestContext) adminpermission.SaveRequest {
	sortValue, _ := strconv.Atoi(c.PostForm("sort"))
	statusValue, _ := strconv.Atoi(c.PostForm("status"))
	return adminpermission.SaveRequest{
		Key:          c.PostForm("key"),
		Name:         c.PostForm("name"),
		Platform:     c.PostForm("platform"),
		Type:         c.PostForm("type"),
		ParentKey:    c.PostForm("parentKey"),
		ResourcePath: firstNonEmpty(c.PostForm("resourcePath"), c.PostForm("path")),
		Path:         c.PostForm("path"),
		Perms:        c.PostForm("perms"),
		Icon:         c.PostForm("icon"),
		Sort:         sortValue,
		Status:       statusValue,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
