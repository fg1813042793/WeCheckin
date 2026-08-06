package user

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	adminuserservice "wecheckin/backend/internal/service/admin/adminuser"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-用户管理
// @Summary 获取用户表单字段列表
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) GetUserFormFields(ctx context.Context, c *app.RequestContext) {
	list, err := adminuserservice.GetUserFormFieldsContext(ctx)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	if list == nil {
		list = []model.UserFormField{}
	}
	response.JSON(c, list)
}

// @Tags PC端-用户管理
// @Summary 保存用户表单字段配置(全量替换)
// @Param fields formData string true "字段JSON数组"
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) SaveUserFormFields(ctx context.Context, c *app.RequestContext) {
	fieldsJSON := c.PostForm("fields")
	var fields []model.UserFormField
	if fieldsJSON != "" {
		if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
			response.Fail(c, "JSON格式错误")
			return
		}
	}
	if err := adminuserservice.SaveUserFormFieldsContext(ctx, fields); err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}
