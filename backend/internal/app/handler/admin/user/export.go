package user

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin/backend/internal/app/service/admincontent"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-用户管理
// @Summary 获取用户数据导出链接
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) UserDataGet(ctx context.Context, c *app.RequestContext) {
	data, err := admincontentservice.GetUserDataURL()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-用户管理
// @Summary 导出用户数据
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) UserDataExport(ctx context.Context, c *app.RequestContext) {
	data, err := admincontentservice.ExportUserDataExcel()
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-用户管理
// @Summary 删除用户数据
// @Success 200 {object} response.Resp
func (h *AdminUserHandler) UserDataDel(ctx context.Context, c *app.RequestContext) {
	err := admincontentservice.DeleteUserDataExcel()
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
