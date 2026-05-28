package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type NewsHandler struct{}

func NewNewsHandler() *NewsHandler { return &NewsHandler{} }

// @Tags 新闻
// @Summary 获取新闻列表
// @Success 200 {object} response.Resp
// @Router /news/list [get]
func (h *NewsHandler) GetNewsList(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetNewsList()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 新闻
// @Summary 查看新闻详情
// @Param id query string true "新闻ID"
// @Success 200 {object} response.Resp
// @Router /news/view [get]
func (h *NewsHandler) ViewNews(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.ViewNews(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
