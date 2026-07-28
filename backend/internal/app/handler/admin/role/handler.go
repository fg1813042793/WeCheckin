package role

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	roleservice "wecheckin-backend/backend/internal/app/service/role"
	"wecheckin-backend/backend/internal/model"
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
	data, err := roleservice.GetListContext(ctx, admin.ID, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-角色管理
// @Summary 获取应用菜单权限树
// @Success 200 {object} response.Resp
// @Router /api/v2/admin/roles/application-permissions [get]
func (h *AdminRoleHandler) GetApplicationPermissionTree(ctx context.Context, c *app.RequestContext) {
	response.JSON(c, roleservice.ApplicationPermissionTree())
}

// @Tags PC端-角色管理
// @Summary 新增角色
// @Param name formData string true "角色名称"
// @Param remark formData string false "备注"
// @Param sort formData int false "排序"
// @Param dataScope formData int false "数据权限范围(1=全部 2=自定义)"
// @Param adminPermissionKeys formData string false "后台权限编码列表(逗号分隔)"
// @Param adminApiPermissionKeys formData string false "后台接口权限编码列表(逗号分隔)"
// @Param deptIds formData string false "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
// @Router /admin/role/add [post]
func (h *AdminRoleHandler) AddRole(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	remark := c.PostForm("remark")
	sort, _ := strconv.Atoi(c.PostForm("sort"))
	dataScope, _ := strconv.Atoi(c.PostForm("dataScope"))
	allowAdminLogin := parseAllowAdminLogin(c.PostForm("allowAdminLogin"), true)
	adminPermissionKeys := parsePermissionKeys(c.PostForm("adminPermissionKeys"))
	adminAPIPermissionKeys := parsePermissionKeys(c.PostForm("adminApiPermissionKeys"))
	clientMenuKeys := parsePermissionKeys(c.PostForm("clientMenuKeys"))
	dingtalkH5MenuKeys := parsePermissionKeys(c.PostForm("dingtalkH5MenuKeys"))
	if dataScope == 0 {
		dataScope = 1
	}
	if name == "" {
		response.Fail(c, "角色名称不能为空")
		return
	}
	var deptIDs []uint
	deptIDsStr := c.PostForm("deptIds")
	if deptIDsStr != "" {
		for _, s := range strings.Split(deptIDsStr, ",") {
			s = strings.TrimSpace(s)
			if did, err := strconv.Atoi(s); err == nil && did > 0 {
				deptIDs = append(deptIDs, uint(did))
			}
		}
	}
	if _, err := roleservice.AddWithAssignmentsContext(ctx, name, remark, c.ClientIP(), sort, dataScope, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, deptIDs, clientMenuKeys, dingtalkH5MenuKeys); err != nil {
		response.Fail(c, "添加失败")
		return
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
// @Param adminPermissionKeys formData string false "后台权限编码列表(逗号分隔)"
// @Param adminApiPermissionKeys formData string false "后台接口权限编码列表(逗号分隔)"
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
	allowAdminLogin := -1
	if value, ok := c.GetPostForm("allowAdminLogin"); ok {
		allowAdminLogin = parseAllowAdminLogin(value, true)
	}
	adminPermissionKeys := parsePermissionKeys(c.PostForm("adminPermissionKeys"))
	adminAPIPermissionKeys := parsePermissionKeys(c.PostForm("adminApiPermissionKeys"))
	clientMenuKeys := parsePermissionKeys(c.PostForm("clientMenuKeys"))
	dingtalkH5MenuKeys := parsePermissionKeys(c.PostForm("dingtalkH5MenuKeys"))
	if dataScope == 0 {
		dataScope = 1
	}
	if name == "" || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var deptIDs []uint
	deptIDsStr := c.PostForm("deptIds")
	if deptIDsStr != "" {
		for _, s := range strings.Split(deptIDsStr, ",") {
			s = strings.TrimSpace(s)
			if did, err := strconv.Atoi(s); err == nil && did > 0 {
				deptIDs = append(deptIDs, uint(did))
			}
		}
	}
	if err := roleservice.EditWithAssignmentsContext(ctx, uint(id), name, remark, c.ClientIP(), sort, status, dataScope, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys, deptIDs, clientMenuKeys, dingtalkH5MenuKeys); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

func parseAllowAdminLogin(value string, defaultValue bool) int {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "0", "false", "off", "no":
		return 0
	case "1", "true", "on", "yes":
		return 1
	default:
		if defaultValue {
			return 1
		}
		return 0
	}
}

func parsePermissionKeys(value string) []string {
	if value == "" {
		return nil
	}
	keys := make([]string, 0)
	seen := map[string]bool{}
	for _, raw := range strings.Split(value, ",") {
		key := strings.TrimSpace(raw)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	return keys
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
	if err := roleservice.DeleteContext(ctx, uint(id)); err != nil {
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
	if err := roleservice.BatchDeleteContext(ctx, ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
