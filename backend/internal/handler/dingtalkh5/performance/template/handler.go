package template

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Template(ctx context.Context, c *app.RequestContext) {
	data, err := dingtalkh5service.TemplateContext(ctx)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.template.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *Handler) SaveTemplate(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var req dingtalkh5service.TemplateDTO
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	data, err := dingtalkh5service.SaveTemplateContext(ctx, user, req)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.performance.template.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}
