package bootstrap

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Bootstrap(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.BootstrapContext(ctx, user)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) Workbench(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := dingtalkh5service.WorkbenchStatsContext(ctx, user)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}
