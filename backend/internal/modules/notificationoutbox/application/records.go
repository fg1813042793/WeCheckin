package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	Enqueue(context.Context, notificationmodel.Outbox) (bool, error)
	ListWebhookRecords(context.Context, WebhookRecordQuery) ([]notificationmodel.Outbox, int64, error)
	GetWebhookRecord(context.Context, uint64) (*notificationmodel.Outbox, error)
	DeleteWebhookRecord(context.Context, uint64) error
}

var (
	ErrWebhookRecordNotFound = errors.New("webhook notification record not found")
	ErrWebhookRecordBusy     = errors.New("webhook notification record is being sent")
)

type WebhookRecord struct {
	ID                 uint64 `json:"id"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	SourceType         string `json:"sourceType"`
	SourceID           string `json:"sourceId"`
	WebhookType        string `json:"webhookType"`
	Target             string `json:"target"`
	Status             string `json:"status"`
	Attempts           int    `json:"attempts"`
	NextRetryAt        int64  `json:"nextRetryAt"`
	LastError          string `json:"lastError"`
	SentAt             int64  `json:"sentAt"`
	ResponseStatusCode int    `json:"responseStatusCode"`
	ResponseBody       string `json:"responseBody"`
	AddTime            int64  `json:"addTime"`
	EditTime           int64  `json:"editTime"`
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

func (service *Service) GetWebhookRecord(ctx context.Context, id uint64) (WebhookRecord, error) {
	store, ok := service.store.(WebhookRecordStore)
	if !ok {
		return WebhookRecord{}, ErrOutboxStoreUnavailable
	}
	row, err := store.GetWebhookRecord(ctx, id)
	if err != nil {
		return WebhookRecord{}, err
	}
	return webhookRecordFromOutbox(*row), nil
}

func (service *Service) DeleteWebhookRecord(ctx context.Context, id uint64) error {
	store, ok := service.store.(WebhookRecordStore)
	if !ok {
		return ErrOutboxStoreUnavailable
	}
	return store.DeleteWebhookRecord(ctx, id)
}

func (service *Service) RetryWebhookRecord(ctx context.Context, id uint64) (WebhookRecord, error) {
	store, ok := service.store.(WebhookRecordStore)
	if !ok {
		return WebhookRecord{}, ErrOutboxStoreUnavailable
	}
	row, err := store.GetWebhookRecord(ctx, id)
	if err != nil {
		return WebhookRecord{}, err
	}
	if row.Status == notificationmodel.StatusSending {
		return WebhookRecord{}, ErrWebhookRecordBusy
	}
	now := service.now().UnixMilli()
	row.ID = 0
	row.IdempotencyKey = fmt.Sprintf("%s:manual:%d", row.IdempotencyKey, now)
	row.Status = notificationmodel.StatusPending
	row.Attempts = 0
	row.NextRetryAt = now
	row.LastError = ""
	row.SentAt = 0
	row.ResponseStatusCode = 0
	row.ResponseBody = ""
	row.AddTime = now
	row.EditTime = now
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	created, err := store.Enqueue(ctx, *row)
	if err != nil {
		return WebhookRecord{}, err
	}
	if !created {
		return WebhookRecord{}, errors.New("webhook retry record was not created")
	}
	return webhookRecordFromOutbox(*row), nil
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
		LastError: row.LastError, SentAt: row.SentAt,
		ResponseStatusCode: row.ResponseStatusCode, ResponseBody: row.ResponseBody,
		AddTime: row.AddTime, EditTime: row.EditTime,
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
