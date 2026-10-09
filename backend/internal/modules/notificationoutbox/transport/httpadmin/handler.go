package httpadmin

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/pkg/response"
)

type Service interface {
	ListWebhookRecords(context.Context, application.WebhookRecordQuery) (application.WebhookRecordList, error)
	GetWebhookRecord(context.Context, uint64) (application.WebhookRecord, error)
	DeleteWebhookRecord(context.Context, uint64) error
	RetryWebhookRecord(context.Context, uint64) (application.WebhookRecord, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List returns sanitized webhook delivery records; recipient URL credentials are never returned.
func (handler *Handler) List(ctx context.Context, c *app.RequestContext) {
	query := application.WebhookRecordQuery{
		Status: c.Query("status"), SourceType: c.Query("sourceType"), WebhookType: c.Query("webhookType"),
		Page: queryInt(c, "page"), PageSize: queryInt(c, "pageSize"),
	}
	query.AddTimeFrom, _ = strconv.ParseInt(c.Query("addTimeFrom"), 10, 64)
	query.AddTimeTo, _ = strconv.ParseInt(c.Query("addTimeTo"), 10, 64)
	result, err := handler.service.ListWebhookRecords(ctx, query)
	if err != nil {
		response.FailInternal(ctx, c, "admin.notification.webhook_list", "查询 Webhook 投递记录失败，请稍后重试", err)
		return
	}
	response.JSON(c, result)
}

func (handler *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "Webhook 记录 ID 无效")
		return
	}
	result, err := handler.service.GetWebhookRecord(ctx, id)
	if err != nil {
		response.FailInternal(ctx, c, "admin.notification.webhook_detail", "查询 Webhook 记录失败，请稍后重试", err)
		return
	}
	response.JSON(c, result)
}

func (handler *Handler) Delete(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "Webhook 记录 ID 无效")
		return
	}
	if err := handler.service.DeleteWebhookRecord(ctx, id); err != nil {
		response.FailInternal(ctx, c, "admin.notification.webhook_delete", "删除 Webhook 记录失败，请稍后重试", err)
		return
	}
	response.JSON(c, map[string]uint64{"id": id})
}

func (handler *Handler) Retry(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "Webhook 记录 ID 无效")
		return
	}
	result, err := handler.service.RetryWebhookRecord(ctx, id)
	if err != nil {
		response.FailInternal(ctx, c, "admin.notification.webhook_retry", "重发 Webhook 失败，请稍后重试", err)
		return
	}
	response.JSON(c, result)
}

func queryInt(c *app.RequestContext, key string) int {
	value, _ := strconv.Atoi(c.Query(key))
	return value
}
