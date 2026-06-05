package client

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler { return &EventHandler{} }

func (h *EventHandler) GetEventList(ctx context.Context, c *app.RequestContext) {
	page, pageSize := 1, 10
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("pageSize")); err == nil && ps > 0 {
		pageSize = ps
	}
	userID := c.Query("user_id")
	keyword := c.Query("keyword")
	typ := c.Query("type")
	data, err := service.GetEventList(page, pageSize, userID, keyword, typ)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *EventHandler) ViewEvent(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	userID := c.Query("user_id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	event, err := service.ViewEvent(id, userID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	response.JSON(c, event)
}

func (h *EventHandler) EventParticipate(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("event_id")
	userID := c.PostForm("user_id")
	if userID == "" {
		userID = c.PostForm("token")
	}
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	err := service.EventParticipate(eventID, userID, forms, addIP)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *EventHandler) GetMyEventList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	typ := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := service.GetMyEventList(userID, typ, status, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}

func (h *EventHandler) GetMyEventRoles(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	if userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	data, err := service.GetMyEventRoles(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

func (h *EventHandler) GetMyManagedList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	typ := c.Query("type")
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := service.GetMyManagedList(userID, typ, status, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}

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
	err := service.PostEventDynamic(eventID, userID, title, content, images, videos, addIP)
	if err != nil {
		response.Fail(c, "发布失败")
		return
	}
	response.JSON(c, nil)
}

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
	result, err := service.GetEventDynamics(eventID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}

func (h *EventHandler) GetEventParticipantList(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("event_id")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := service.GetEventParticipantList(eventID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list})
}

func (h *EventHandler) SaveEventScore(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("event_id")
	participantID := c.PostForm("participant_id")
	score := c.PostForm("score")
	judgeID := c.PostForm("judge_id")
	if eventID == "" || participantID == "" || score == "" || judgeID == "" {
		response.Fail(c, "参数错误")
		return
	}
	err := service.SaveEventScore(eventID, participantID, score, judgeID)
	if err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}

func (h *EventHandler) GetEventScores(ctx context.Context, c *app.RequestContext) {
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
	result, err := service.GetEventScores(eventID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}
