package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	formkitadminservice "wecheckin/backend/internal/service/admin/formkitadmin"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// TemplatePresetsGet GET /admin/survey/template_presets
func (h *AdminSurveyHandler) TemplatePresetsGet(ctx context.Context, c *app.RequestContext) {
	admin, ok := c.Get("admin")
	if !ok {
		response.JSON(c, []formkitadminservice.TemplatePreset{})
		return
	}
	a := admin.(*model.Admin)
	out, err := formkitadminservice.GetTemplatePresetsContext(ctx, a.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, out)
}

// TemplatePresetsSave POST /admin/survey/template_presets
func (h *AdminSurveyHandler) TemplatePresetsSave(ctx context.Context, c *app.RequestContext) {
	admin, ok := c.Get("admin")
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	a := admin.(*model.Admin)
	var req struct {
		Presets []formkitadminservice.TemplatePreset `json:"presets"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if err := formkitadminservice.SaveTemplatePresetsContext(ctx, a.ID, req.Presets); err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}
