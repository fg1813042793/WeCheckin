package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminHomeHandler struct{}

func NewAdminHomeHandler() *AdminHomeHandler { return &AdminHomeHandler{} }

// @Tags 管理后台首页
// @Summary 管理后台首页数据
// @Success 200 {object} response.Resp
// @Router /admin/home [get]
func (h *AdminHomeHandler) AdminHome(ctx context.Context, c *app.RequestContext) {
	data, err := service.AdminHome()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 管理后台首页
// @Summary 清除推荐数据
// @Success 200 {object} response.Resp
// @Router /admin/clear_vouch [get]
func (h *AdminHomeHandler) ClearVouchData(ctx context.Context, c *app.RequestContext) {
	err := service.ClearVouchData()
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
