package event

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin/backend/internal/app/service/event"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-赛事活动管理
// @Summary 发布活动动态(管理端)
// @Param eventId formData string true "活动ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) PostEventDynamic(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("eventId")
	title := c.PostForm("title")
	content := c.PostForm("content")
	images := c.PostForm("images")
	videos := c.PostForm("videos")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	userID := "admin_" + admin.Name
	addIP := c.ClientIP()
	if err := eventservice.PostEventDynamicForAdminContext(ctx, eventID, userID, title, content, images, videos, addIP, admin.ID); err != nil {
		response.Fail(c, "发布失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取活动动态列表(管理端)
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetEventDynamics(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := eventservice.GetEventDynamicsForAdminContext(ctx, eventID, 1, 100, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动动态
// @Param id formData string true "动态ID"
// @Param title formData string false "动态标题"
// @Param content formData string false "动态内容"
// @Param images formData string false "图片列表(JSON)"
// @Param videos formData string false "视频列表(JSON)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) EditEventDynamic(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	title := c.PostForm("title")
	content := c.PostForm("content")
	images := c.PostForm("images")
	videos := c.PostForm("videos")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	editIP := c.ClientIP()
	if err := eventservice.EditEventDynamicForAdminContext(ctx, id, title, content, images, videos, editIP, admin.ID); err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 删除活动动态
// @Param id formData string true "动态ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEventDynamic(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := eventservice.DelEventDynamicForAdminContext(ctx, id, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动动态
// @Param ids formData string true "动态ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEventDynamics(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := eventservice.DelEventDynamicsForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
