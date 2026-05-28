package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler { return &HomeHandler{} }

// @Tags 首页
// @Summary 获取系统设置
// @Param key query string true "设置键名"
// @Success 200 {object} response.Resp
// @Router /home/setup_get [get]
func (h *HomeHandler) GetSetup(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	setup, err := service.GetSetup(key)
	if err != nil {
		response.Fail(c, "参数错误")
		return
	}
	response.JSON(c, setup)
}

// @Tags 首页
// @Summary 获取首页列表
// @Success 200 {object} response.Resp
// @Router /home/list [get]
func (h *HomeHandler) GetHomeList(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetHomeList()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
