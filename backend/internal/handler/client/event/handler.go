package event

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin/backend/internal/service/client/event"
	"wecheckin/backend/pkg/response"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler { return &EventHandler{} }

// @Tags 客户端-赛事活动
// @Summary 获取活动列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param user_id query string false "用户ID"
// @Param keyword query string false "搜索关键词"
// @Param type query string false "活动类型"
// @Success 200 {object} response.Resp
// @Router /event/list [get]
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
	data, err := eventservice.GetEventListContext(ctx, page, pageSize, userID, keyword, typ)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-赛事活动
// @Summary 查看活动详情
// @Param id query string true "活动ID"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /event/view [get]
func (h *EventHandler) ViewEvent(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	userID := c.Query("user_id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	event, err := eventservice.ViewEventContext(ctx, id, userID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	response.JSON(c, event)
}

// @Tags 客户端-赛事活动
// @Summary 参与活动报名
// @Param event_id formData string true "活动ID"
// @Param user_id formData string true "用户ID"
// @Param forms formData string false "报名表单数据(JSON)"
// @Success 200 {object} response.Resp
// @Router /event/participate [post]
func (h *EventHandler) EventParticipate(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("event_id")
	userID := c.PostForm("user_id")
	if userID == "" {
		userID = c.PostForm("token")
	}
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	err := eventservice.EventParticipateContext(ctx, eventID, userID, forms, addIP)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags 客户端-赛事活动
// @Summary 获取我的活动列表
// @Param user_id query string true "用户ID"
// @Param type query string false "活动类型"
// @Param status query string false "活动状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /event/my_list [get]
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
	result, err := eventservice.GetMyEventListContext(ctx, userID, typ, status, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}

// @Tags 客户端-赛事活动
// @Summary 获取我的活动角色
// @Param user_id query string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /event/my_roles [get]
func (h *EventHandler) GetMyEventRoles(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	if userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	data, err := eventservice.GetMyEventRolesContext(ctx, userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 客户端-赛事活动
// @Summary 获取我管理的活动列表
// @Param user_id query string true "用户ID"
// @Param type query string false "活动类型"
// @Param status query string false "活动状态"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /event/my_managed [get]
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
	result, err := eventservice.GetMyManagedListContext(ctx, userID, typ, status, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}

// @Tags 客户端-赛事活动
// @Summary 获取活动参与成员列表
// @Param event_id query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /event/participant_list [get]
func (h *EventHandler) GetEventParticipantList(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("event_id")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := eventservice.GetEventParticipantListContext(ctx, eventID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, participantListResponse{List: list})
}
