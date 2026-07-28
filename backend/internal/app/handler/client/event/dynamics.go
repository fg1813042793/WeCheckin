package event

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin-backend/backend/internal/app/service/event"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags 客户端-赛事活动
// @Summary 发布活动动态
// @Param event_id formData string true "活动ID"
// @Param user_id formData string true "用户ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
// @Router /event/dynamic_post [post]
func (h *EventHandler) PostEventDynamic(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("event_id")
	userID := c.PostForm("user_id")
	title := c.PostForm("title")
	content := c.PostForm("content")
	images := c.PostForm("images")
	videos := c.PostForm("videos")
	addIP := c.ClientIP()
	if eventID == "" || userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	err := eventservice.PostEventDynamicContext(ctx, eventID, userID, title, content, images, videos, addIP)
	if err != nil {
		response.Fail(c, "发布失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 客户端-赛事活动
// @Summary 获取活动动态列表
// @Param event_id query string true "活动ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /event/dynamics [get]
func (h *EventHandler) GetEventDynamics(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("event_id")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := eventservice.GetEventDynamicsContext(ctx, eventID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}
