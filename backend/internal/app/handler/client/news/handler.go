package news

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"strconv"
	newsservice "wecheckin/backend/internal/app/service/news"
	"wecheckin/backend/pkg/response"
)

type NewsHandler struct{}

func NewNewsHandler() *NewsHandler { return &NewsHandler{} }

// @Tags 客户端-通知公告
// @Summary 获取通知公告列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /news/list [get]
func (h *NewsHandler) GetNewsList(ctx context.Context, c *app.RequestContext) {
	page, pageSize := 1, 10
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("pageSize")); err == nil && ps > 0 {
		pageSize = ps
	}
	keyword := c.Query("keyword")
	userID := c.Query("user_id")
	data, err := newsservice.GetNewsListContext(ctx, page, pageSize, keyword, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通知公告
// @Summary 查看通知公告详情
// @Param id query string true "通知公告ID"
// @Success 200 {object} response.Resp
// @Router /news/view [get]
func (h *NewsHandler) ViewNews(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := newsservice.ViewNewsContext(ctx, id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-通知公告
// @Summary 获取通知公告分类列表
// @Success 200 {object} response.Resp
// @Router /news/cate_list [get]
func (h *NewsHandler) GetNewsCateList(ctx context.Context, c *app.RequestContext) {
	list, err := newsservice.GetNewsCateListContext(ctx)
	if err != nil {
		response.JSON(c, []interface{}{})
		return
	}
	response.JSON(c, list)
}
