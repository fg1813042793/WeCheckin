package survey

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// TemplatePresetsGet GET /admin/survey/template_presets
func (h *AdminSurveyHandler) TemplatePresetsGet(_ context.Context, c *app.RequestContext) {
	admin, ok := c.Get("admin")
	if !ok {
		response.JSON(c, []map[string]string{})
		return
	}
	a := admin.(*model.Admin)
	key := fmt.Sprintf("template_presets_%d", a.ID)
	var entry model.Setup
	if err := database.DB.Where("`setup_key` = ?", key).First(&entry).Error; err != nil {
		response.JSON(c, []map[string]string{})
		return
	}
	var out []map[string]string
	json.Unmarshal([]byte(entry.Value), &out)
	response.JSON(c, out)
}

// TemplatePresetsSave POST /admin/survey/template_presets
func (h *AdminSurveyHandler) TemplatePresetsSave(_ context.Context, c *app.RequestContext) {
	admin, ok := c.Get("admin")
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	a := admin.(*model.Admin)
	var req struct {
		Presets []map[string]string `json:"presets"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	key := fmt.Sprintf("template_presets_%d", a.ID)
	now := time.Now().Unix()
	bytes, _ := json.Marshal(req.Presets)
	var entry model.Setup
	if err := database.DB.Where("`setup_key` = ?", key).First(&entry).Error; err == nil {
		database.DB.Model(&model.Setup{}).Where("`setup_key` = ?", key).Updates(map[string]interface{}{
			"setup_value":     string(bytes),
			"setup_edit_time": now,
		})
	} else {
		database.DB.Create(&model.Setup{
			Key:      key,
			Value:    string(bytes),
			Type:     "template_presets",
			AddTime:  now,
			EditTime: now,
		})
	}
	response.JSON(c, nil)
}
