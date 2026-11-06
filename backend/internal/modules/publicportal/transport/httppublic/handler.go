package httppublic

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/modules/publicportal/application"
	"wecheckin/backend/pkg/response"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

var _ Service = (*application.Service)(nil)

func (h *Handler) List(ctx context.Context, c *app.RequestContext) {
	c.Response.Header.Set("Cache-Control", "no-store")
	if h == nil || h.service == nil {
		response.FailInternal(ctx, c, "publicportal.public", "公共页面暂不可用", errors.New("public portal public service unavailable"))
		return
	}
	data, err := h.service.List(ctx)
	if err == nil {
		response.JSON(c, data)
		return
	}
	if application.PublicErrorMessage(err) != "" {
		response.Fail(c, "公共页面暂不可用")
		return
	}
	response.FailInternal(ctx, c, "publicportal.public", "公共页面暂不可用", err)
}
