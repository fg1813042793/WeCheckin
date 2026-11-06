package httpadmin

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/publicportal/application"
	"wecheckin/backend/pkg/response"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

var _ Service = (*application.Service)(nil)

func (h *Handler) ready(ctx context.Context, c *app.RequestContext) (uint, bool) {
	c.Response.Header.Set("Cache-Control", "no-store")
	value, ok := c.Get("admin")
	if !ok {
		respond(ctx, c, nil, application.ErrForbidden)
		return 0, false
	}
	admin, ok := value.(*model.Admin)
	if !ok || admin == nil || admin.ID == 0 {
		respond(ctx, c, nil, application.ErrForbidden)
		return 0, false
	}
	if h == nil || h.service == nil {
		respond(ctx, c, nil, errors.New("public portal admin service unavailable"))
		return 0, false
	}
	return admin.ID, true
}
func (h *Handler) Get(ctx context.Context, c *app.RequestContext) {
	id, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	data, err := h.service.Get(ctx, id)
	respond(ctx, c, data, err)
}
func (h *Handler) Preview(ctx context.Context, c *app.RequestContext) {
	id, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	data, err := h.service.Preview(ctx, id)
	respond(ctx, c, data, err)
}
func (h *Handler) SaveSettings(ctx context.Context, c *app.RequestContext) {
	id, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	var body application.Settings
	if err := decodeBody(c, &body); err != nil {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := h.service.SaveSettings(ctx, id, body)
	respond(ctx, c, data, err)
}
func (h *Handler) Options(ctx context.Context, c *app.RequestContext) {
	id, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	q, err := optionsQuery(c)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	data, err := h.service.Options(ctx, id, q)
	respond(ctx, c, data, err)
}
func (h *Handler) CreateItem(ctx context.Context, c *app.RequestContext) {
	id, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	var body application.ItemInput
	if err := decodeBody(c, &body); err != nil {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := h.service.CreateItem(ctx, id, body)
	respond(ctx, c, data, err)
}
func (h *Handler) UpdateItem(ctx context.Context, c *app.RequestContext) {
	actor, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	id, err := positive(c.Param("id"))
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	var body application.ItemUpdate
	if err = decodeBody(c, &body); err != nil || body.Version == 0 {
		respond(ctx, c, nil, application.ErrInvalidArgument)
		return
	}
	data, err := h.service.UpdateItem(ctx, actor, id, body)
	respond(ctx, c, data, err)
}
func (h *Handler) DeleteItem(ctx context.Context, c *app.RequestContext) {
	actor, ok := h.ready(ctx, c)
	if !ok {
		return
	}
	id, err := positive(c.Param("id"))
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	raw, err := queryValue(c, "version")
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	version, err := positive(raw)
	if err != nil {
		respond(ctx, c, nil, err)
		return
	}
	data, err := h.service.DeleteItem(ctx, actor, id, version)
	respond(ctx, c, data, err)
}
func respond(ctx context.Context, c *app.RequestContext, data any, err error) {
	if err == nil {
		response.JSON(c, data)
		return
	}
	if message := application.PublicErrorMessage(err); message != "" {
		response.Fail(c, message)
		return
	}
	response.FailInternal(ctx, c, "publicportal.admin", "公共页面操作失败，请稍后重试", err)
}
