package user

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) ListUsers(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.ListUsersContext(ctx, user)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *Handler) CreateUser(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var payload dingtalkh5service.UserPayload
	_ = c.BindAndValidate(&payload)
	created, users, err := dingtalkh5service.CreateUserContext(ctx, user, payload)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, utils.H{"user": created, "users": users})
}

func (h *Handler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var payload dingtalkh5service.UserPayload
	_ = c.BindAndValidate(&payload)
	updated, users, err := dingtalkh5service.UpdateUserContext(ctx, user, c.Param("id"), payload)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, utils.H{"user": updated, "users": users})
}

func (h *Handler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	users, err := dingtalkh5service.DeleteUserContext(ctx, user, c.Param("id"))
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.user.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, utils.H{"users": users})
}
