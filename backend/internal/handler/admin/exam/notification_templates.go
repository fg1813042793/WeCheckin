package exam

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	formkitadminservice "wecheckin/backend/internal/service/admin/formkitadmin"
	"wecheckin/backend/pkg/response"
)

func (h *AdminExamHandler) NotificationTemplatesGet(ctx context.Context, c *app.RequestContext) {
	adminValue, ok := c.Get("admin")
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	data, err := formkitadminservice.GetExamNotificationTemplatesContext(ctx, adminValue.(*model.Admin).ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *AdminExamHandler) NotificationTemplatesSave(ctx context.Context, c *app.RequestContext) {
	adminValue, ok := c.Get("admin")
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var req struct {
		Scope   string                               `json:"scope"`
		Presets []formkitadminservice.TemplatePreset `json:"presets"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	err := formkitadminservice.SaveExamNotificationTemplatesContext(ctx, adminValue.(*model.Admin).ID, req.Scope, req.Presets, c.ClientIP())
	if errors.Is(err, formkitadminservice.ErrInvalidNotificationTemplateScope) {
		response.Fail(c, "参数错误")
		return
	}
	if err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}
