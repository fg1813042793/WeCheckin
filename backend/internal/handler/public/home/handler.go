package home

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	setupservice "wecheckin/backend/internal/service/admin/setup"
	homeservice "wecheckin/backend/internal/service/client/home"
	"wecheckin/backend/pkg/response"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler { return &HomeHandler{} }

// @Tags 客户端-首页
// @Summary 获取系统设置
// @Param key query string true "设置键名"
// @Success 200 {object} response.Resp
// @Router /home/setup_get [get]
func (h *HomeHandler) GetSetup(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	setup, err := setupservice.GetSetupContext(ctx, key)
	if err != nil || setup == nil {
		response.JSON(c, nil)
		return
	}
	// Return the value directly (rich text JSON or plain text)
	response.JSON(c, setup.Value)
}

// @Tags 客户端-首页
// @Summary 获取首页列表
// @Success 200 {object} response.Resp
// @Router /home/list [get]
func (h *HomeHandler) GetHomeList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := homeservice.GetHomeListContext(ctx, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
