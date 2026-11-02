package home

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/model"
	dashboardservice "wecheckin/backend/internal/service/admin/dashboard"
	"wecheckin/backend/pkg/response"
)

type AdminHomeHandler struct{}

func NewAdminHomeHandler() *AdminHomeHandler { return &AdminHomeHandler{} }

// @Tags PC端-管理后台首页
// @Summary 管理后台首页数据
// @Param days query int false "趋势天数" Enums(7,30) default(7)
// @Success 200 {object} response.Resp
func (h *AdminHomeHandler) AdminHome(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	days, _ := strconv.Atoi(strings.TrimSpace(c.Query("days")))
	data, err := dashboardservice.AdminHomeWithDaysContext(ctx, admin.ID, days)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags PC端-管理后台首页
// @Summary 清除推荐数据
// @Success 200 {object} response.Resp
func (h *AdminHomeHandler) ClearVouchData(ctx context.Context, c *app.RequestContext) {
	err := dashboardservice.ClearVouchDataContext(ctx)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
