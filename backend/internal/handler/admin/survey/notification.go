package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	surveyservice "wecheckin/backend/internal/service/client/survey"
	"wecheckin/backend/pkg/response"
)

// NotifyList GET /admin/notify/list
func (h *AdminSurveyHandler) NotifyList(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	result, err := h.survey.NotificationListContext(ctx, surveyservice.NotificationQuery{
		SourceType: c.Query("sourceType"),
		SourceID:   c.Query("sourceId"),
		UserID:     c.Query("userId"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, surveyNotificationListResponse{List: result.List, Total: result.Total})
}

// NotifyRead POST /admin/notify/read
func (h *AdminSurveyHandler) NotifyRead(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var req struct {
		ID     uint   `json:"id"`
		All    bool   `json:"all"`
		UserID string `json:"userId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if err := h.survey.MarkNotificationsReadContext(ctx, surveyservice.NotificationReadInput{
		ID:     req.ID,
		All:    req.All,
		UserID: req.UserID,
	}); err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// NotifyUnreadCount GET /admin/notify/unread_count
func (h *AdminSurveyHandler) NotifyUnreadCount(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	userID := c.Query("userId")
	count, err := h.survey.NotificationUnreadCountContext(ctx, userID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]int64{"count": count})
}
