package application

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	notificationmodel "wecheckin/backend/internal/model/notification"
)

type WebhookRecordQuery struct {
	Status      string
	SourceType  string
	WebhookType string
	AddTimeFrom int64
	AddTimeTo   int64
	Page        int
	PageSize    int
}

type WebhookRecordStore interface {
	ListWebhookRecords(context.Context, WebhookRecordQuery) ([]notificationmodel.Outbox, int64, error)
}

type WebhookRecord struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	SourceType  string `json:"sourceType"`
	SourceID    string `json:"sourceId"`
	WebhookType string `json:"webhookType"`
	Target      string `json:"target"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	NextRetryAt int64  `json:"nextRetryAt"`
	LastError   string `json:"lastError"`
	SentAt      int64  `json:"sentAt"`
	AddTime     int64  `json:"addTime"`
	EditTime    int64  `json:"editTime"`
}

type WebhookRecordList struct {
	List     []WebhookRecord `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

func (service *Service) ListWebhookRecords(ctx context.Context, query WebhookRecordQuery) (WebhookRecordList, error) {
	store, ok := service.store.(WebhookRecordStore)
	if !ok {
		return WebhookRecordList{}, ErrOutboxStoreUnavailable
	}
	query.Page, query.PageSize = normalizeRecordPagination(query.Page, query.PageSize)
	rows, total, err := store.ListWebhookRecords(ctx, query)
	if err != nil {
		return WebhookRecordList{}, err
	}
	result := WebhookRecordList{List: make([]WebhookRecord, 0, len(rows)), Total: total, Page: query.Page, PageSize: query.PageSize}
	for _, row := range rows {
		result.List = append(result.List, webhookRecordFromOutbox(row))
	}
	return result, nil
}

var ErrOutboxStoreUnavailable = errOutboxStoreUnavailable{}

type errOutboxStoreUnavailable struct{}

func (errOutboxStoreUnavailable) Error() string {
	return "notification outbox record store is not initialized"
}

func normalizeRecordPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func webhookRecordFromOutbox(row notificationmodel.Outbox) WebhookRecord {
	var recipient WebhookRecipient
	_ = json.Unmarshal([]byte(row.RecipientJSON), &recipient)
	var payload MessagePayload
	_ = json.Unmarshal([]byte(row.PayloadJSON), &payload)
	return WebhookRecord{
		ID: row.ID, Title: payload.Title, Content: payload.Content,
		SourceType: row.SourceType, SourceID: row.SourceID,
		WebhookType: strings.TrimSpace(recipient.Type), Target: maskWebhookTarget(recipient.URL),
		Status: row.Status, Attempts: row.Attempts, NextRetryAt: row.NextRetryAt,
		LastError: row.LastError, SentAt: row.SentAt, AddTime: row.AddTime, EditTime: row.EditTime,
	}
}

func maskWebhookTarget(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "已配置目标"
	}
	return parsed.Scheme + "://" + parsed.Host + "/***"
}

func webhookStatusLabel(status string) string {
	switch status {
	case notificationmodel.StatusPending:
		return "待发送"
	case notificationmodel.StatusSending:
		return "发送中"
	case notificationmodel.StatusSent:
		return "已发送"
	case notificationmodel.StatusFailed:
		return "待重试"
	case notificationmodel.StatusDead:
		return "已终止"
	default:
		return "未知状态 " + strconv.Quote(status)
	}
}
