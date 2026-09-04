package user

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/handler/internal/param"
	"wecheckin/backend/internal/model"
	adminuserservice "wecheckin/backend/internal/service/admin/adminuser"
	"wecheckin/backend/pkg/response"
)

type AdminUserHandler struct{}

func NewAdminUserHandler() *AdminUserHandler { return &AdminUserHandler{} }

// @Tags PC端-用户管理
// @Summary 获取用户列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) GetUserList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	page, _ := strconv.Atoi(c.Query("page"))
	sizeStr := c.Query("pageSize")
	if sizeStr == "" {
		sizeStr = c.Query("size")
	}
	size, _ := strconv.Atoi(sizeStr)
	keyword := c.Query("keyword")
	sortStr := c.Query("sort")
	list, total, err := adminuserservice.GetUserListContext(ctx, keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, userListResponse{List: list, Total: total})
}

// @Tags PC端-用户管理
// @Summary 获取用户详情
// @Param openid query string true "用户OpenID"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	openID := c.Query("openid")
	data, err := adminuserservice.GetUserByOpenIDForAdminContext(ctx, openID, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-用户管理
// @Summary 根据ID获取用户详情
// @Param id query string true "用户ID"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) GetUserByID(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.Query("id")
	data, err := adminuserservice.GetUserByIDForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-用户管理
// @Summary 新增用户
// @Param name formData string true "用户名"
// @Param mobile formData string false "手机号"
// @Param positionId formData string false "岗位ID"
// @Param managerUserId formData string false "直属上级用户ID"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) AddUser(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	positionID := parsePositionID(c.PostForm("positionId"))
	managerUserID := parsePositionID(c.PostForm("managerUserId"))
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := param.ParseUintSlice(c.PostForm("deptIds"))
	adminAccess, _ := parseAdminAccess(c)
	err := adminuserservice.AddUserWithManagerAndAdminAccessContext(ctx, name, mobile, pic, forms, addIP, positionID, managerUserID, deptIds, adminAccess)
	if err != nil {
		response.FailInternal(ctx, c, "admin.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-用户管理
// @Summary 编辑用户
// @Param id formData string true "用户ID"
// @Param name formData string false "用户名"
// @Param mobile formData string false "手机号"
// @Param positionId formData string false "岗位ID"
// @Param managerUserId formData string false "直属上级用户ID"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) EditUser(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	positionID := parsePositionID(c.PostForm("positionId"))
	managerUserID := parsePositionID(c.PostForm("managerUserId"))
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := param.ParseUintSlice(c.PostForm("deptIds"))
	adminAccess, hasAdminAccess := parseAdminAccess(c)
	err := adminuserservice.EditUserWithManagerAndAdminAccessForAdminContext(ctx, id, name, mobile, pic, forms, addIP, positionID, managerUserID, deptIds, adminAccess, hasAdminAccess, admin.ID)
	if err != nil {
		response.FailInternal(ctx, c, "admin.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-用户管理
// @Summary 删除用户
// @Param id formData string true "用户ID"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) DelUser(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	err := adminuserservice.DelUserForAdminContext(ctx, id, admin.ID)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) DelUsers(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := adminuserservice.DelUsersForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func parsePositionID(value string) uint {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0
	}
	return uint(id)
}

func parseAdminAccess(c *app.RequestContext) (adminuserservice.AdminAccessInput, bool) {
	hasPermissionKeys := hasPostForm(c, "allowPermissionKeys") ||
		hasPostForm(c, "denyPermissionKeys")
	hasDataScopeExtras := hasPostForm(c, "extraDataDeptIds") ||
		hasPostForm(c, "extraDataUserIds")
	hasAdminAccess := hasPostForm(c, "password") ||
		hasPostForm(c, "roleId") ||
		hasPostForm(c, "roleIds") ||
		hasPermissionKeys ||
		hasDataScopeExtras

	roleID, _ := strconv.Atoi(c.PostForm("roleId"))
	roleIDs := param.ParseUintSlice(c.PostForm("roleIds"))

	return adminuserservice.AdminAccessInput{
		Password:               c.PostForm("password"),
		RoleID:                 uint(roleID),
		RoleIDs:                roleIDs,
		AllowPermissionKeys:    parsePermissionKeys(c.PostForm("allowPermissionKeys")),
		DenyPermissionKeys:     parsePermissionKeys(c.PostForm("denyPermissionKeys")),
		PermissionKeysTouched:  hasPermissionKeys,
		ExtraDataDeptIDs:       param.ParseUintSlice(c.PostForm("extraDataDeptIds")),
		ExtraDataUserIDs:       param.ParseUintSlice(c.PostForm("extraDataUserIds")),
		DataScopeExtrasTouched: hasDataScopeExtras,
	}, hasAdminAccess
}

func parsePermissionKeys(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			keys = append(keys, part)
		}
	}
	return keys
}

func hasPostForm(c *app.RequestContext, key string) bool {
	_, ok := c.GetPostForm(key)
	return ok
}

// @Tags PC端-用户管理
// @Summary 设置用户状态
// @Param id formData string true "用户ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) StatusUser(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	reason := c.PostForm("reason")
	err := adminuserservice.StatusUserForAdminContext(ctx, id, status, reason, admin.ID)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := adminuserservice.ResetUserPasswordForAdminContext(ctx, id, admin.ID); err != nil {
		response.Fail(c, "重置失败")
		return
	}
	response.JSON(c, nil)
}
