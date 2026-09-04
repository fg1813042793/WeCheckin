package notification

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	service "wecheckin/backend/internal/service/dingtalkh5/notification"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/response"
)

type Handler struct {
	service *service.Service
}

func NewHandler() *Handler {
	return &Handler{service: service.NewService(database.GetDB())}
}

func (h *Handler) List(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("pageSize")))
	data, err := h.service.List(ctx, user.ID, page, pageSize)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, data)
}

func (h *Handler) UnreadCount(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	count, err := h.service.UnreadCount(ctx, user.ID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, map[string]int64{"count": count})
}

func (h *Handler) MarkRead(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	notificationID, valid := parseNotificationID(c.Param("id"))
	if !valid {
		response.Fail(c, "通知不存在")
		return
	}
	if err := h.service.MarkRead(ctx, user.ID, notificationID); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			response.Fail(c, "通知不存在")
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func (h *Handler) MarkAllRead(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	if err := h.service.MarkAllRead(ctx, user.ID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

func parseNotificationID(value string) (uint, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}
