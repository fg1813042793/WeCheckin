package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// NotifyList GET /admin/notify/list
func (h *AdminSurveyHandler) NotifyList(ctx context.Context, c *app.RequestContext) {
	sourceType := c.Query("sourceType")
	sourceID := c.Query("sourceId")
	userID := c.Query("userId")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Notify{})
	if sourceType != "" {
		q = q.Where("`notify_source_type` = ?", sourceType)
	}
	if sourceID != "" {
		q = q.Where("`notify_source_id` = ?", sourceID)
	}
	if userID != "" {
		q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", userID)
	}
	var total int64
	q.Count(&total)
	var list []model.Notify
	q.Order("`notify_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, surveyNotificationListResponse{List: list, Total: total})
}

// NotifyRead POST /admin/notify/read
func (h *AdminSurveyHandler) NotifyRead(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID     uint   `json:"id"`
		All    bool   `json:"all"`
		UserID string `json:"userId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if req.All {
		q := db.Model(&model.Notify{}).Where("`notify_is_read` = 0")
		if req.UserID != "" {
			q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", req.UserID)
		}
		q.UpdateColumn("notify_is_read", 1)
	} else if req.ID > 0 {
		db.Model(&model.Notify{}).Where("`notify_id` = ?", req.ID).UpdateColumn("notify_is_read", 1)
	}
	response.JSON(c, nil)
}

// NotifyUnreadCount GET /admin/notify/unread_count
func (h *AdminSurveyHandler) NotifyUnreadCount(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("userId")
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Notify{}).Where("`notify_is_read` = 0")
	if userID != "" {
		q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", userID)
	}
	var cnt int64
	q.Count(&cnt)
	response.JSON(c, map[string]int64{"count": cnt})
}
