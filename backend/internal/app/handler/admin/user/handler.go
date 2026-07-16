package user

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/app/handler/internal/param"
	adminuserservice "wecheckin-backend/backend/internal/app/service/adminuser"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

type AdminUserHandler struct{}

func NewAdminUserHandler() *AdminUserHandler { return &AdminUserHandler{} }

// @Tags PC端-用户管理
// @Summary 获取用户列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/user_list [get]
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
// @Router /admin/user_detail [get]
func (h *AdminUserHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	openID := c.Query("openid")
	data, err := adminuserservice.GetUserByOpenIDContext(ctx, openID)
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
// @Router /admin/user_detail_by_id [get]
func (h *AdminUserHandler) GetUserByID(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := adminuserservice.GetUserByIDContext(ctx, id)
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
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
// @Router /admin/user_add [post]
func (h *AdminUserHandler) AddUser(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := param.ParseUintSlice(c.PostForm("deptIds"))
	err := adminuserservice.AddUserContext(ctx, name, mobile, pic, forms, addIP, deptIds)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-用户管理
// @Summary 编辑用户
// @Param id formData string true "用户ID"
// @Param name formData string false "用户名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
// @Router /admin/user_edit [post]
func (h *AdminUserHandler) EditUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := param.ParseUintSlice(c.PostForm("deptIds"))
	err := adminuserservice.EditUserContext(ctx, id, name, mobile, pic, forms, addIP, deptIds)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-用户管理
// @Summary 删除用户
// @Param id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/user_del [post]
func (h *AdminUserHandler) DelUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := adminuserservice.DelUserContext(ctx, id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) DelUsers(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := adminuserservice.DelUsersContext(ctx, ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-用户管理
// @Summary 设置用户状态
// @Param id formData string true "用户ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/user_status [post]
func (h *AdminUserHandler) StatusUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	reason := c.PostForm("reason")
	err := adminuserservice.StatusUserContext(ctx, id, status, reason)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := adminuserservice.ResetUserPasswordContext(ctx, id); err != nil {
		response.Fail(c, "重置失败")
		return
	}
	response.JSON(c, nil)
}
