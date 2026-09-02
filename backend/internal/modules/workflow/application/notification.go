package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	NotificationStatusPending = "pending"
	NotificationStatusSending = "sending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
	NotificationStatusDead    = "dead"

	notificationDefaultDispatchLimit = 100
	notificationMaxDispatchLimit     = 500
	notificationSendingLease         = 10 * time.Minute
	notificationMaxAttempts          = 5
	notificationErrorMaxRunes        = 1000
)

var ErrNotificationChannelUnavailable = errors.New("通知渠道不可用")

type NotificationRecord struct {
	ID                string              `json:"id"`
	InstanceID        string              `json:"instanceId"`
	NodeID            string              `json:"nodeId"`
	TaskID            string              `json:"taskId"`
	RecipientUserID   string              `json:"recipientUserId"`
	Kind              string              `json:"kind"`
	Channel           string              `json:"channel"`
	Status            string              `json:"status"`
	Payload           NotificationPayload `json:"payload"`
	CorpID            string              `json:"corpId"`
	ProviderMessageID string              `json:"providerMessageId"`
	Attempts          int                 `json:"attempts"`
	NextRetryAt       int64               `json:"nextRetryAt"`
	LastError         string              `json:"lastError"`
	SentAt            int64               `json:"sentAt"`
	AddTime           int64               `json:"addTime"`
	EditTime          int64               `json:"editTime"`
}

type NotificationQuery struct {
	InstanceID      string
	RecipientUserID string
	Kind            string
	Channel         string
	Status          string
	Page            int
	PageSize        int
}

type NotificationList struct {
	List     []NotificationRecord `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type NotificationRepository interface {
	List(ctx context.Context, query NotificationQuery) (*NotificationList, error)
	ClaimByIDs(ctx context.Context, ids []string, now, staleBefore int64) ([]NotificationRecord, error)
	ClaimDue(ctx context.Context, now, staleBefore int64, limit int) ([]NotificationRecord, error)
	DeliverInApp(ctx context.Context, notification NotificationRecord, now int64) error
	MarkSent(ctx context.Context, id, providerMessageID string, now int64) error
	MarkFailed(ctx context.Context, id string, attempts int, status string, nextRetryAt int64, message string, now int64) error
	ResetForRetry(ctx context.Context, id string, now int64) error
}

type NotificationDeliveryResult struct {
	ID                string
	ProviderMessageID string
	Err               error
}

type NotificationChannel interface {
	Name() string
	Deliver(ctx context.Context, notifications []NotificationRecord) []NotificationDeliveryResult
}

type NotificationDispatcher interface {
	List(ctx context.Context, query NotificationQuery) (*NotificationList, error)
	Dispatch(ctx context.Context, ids []string) (int, error)
	DispatchDue(ctx context.Context, limit int) (int, error)
	Retry(ctx context.Context, id string) error
}

type notificationDispatcher struct {
	repository NotificationRepository
	channels   map[string]NotificationChannel
	now        func() time.Time
}

func NewNotificationDispatcher(repository NotificationRepository, channels ...NotificationChannel) NotificationDispatcher {
	return newNotificationDispatcherWithClock(repository, time.Now, channels...)
}

func newNotificationDispatcherWithClock(repository NotificationRepository, now func() time.Time, channels ...NotificationChannel) *notificationDispatcher {
	registered := make(map[string]NotificationChannel, len(channels))
	for _, channel := range channels {
		if channel == nil || strings.TrimSpace(channel.Name()) == "" {
			continue
		}
		registered[channel.Name()] = channel
	}
	return &notificationDispatcher{repository: repository, channels: registered, now: now}
}

func (dispatcher *notificationDispatcher) List(ctx context.Context, query NotificationQuery) (*NotificationList, error) {
	return dispatcher.repository.List(ctx, query)
}

func (dispatcher *notificationDispatcher) Dispatch(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	now := dispatcher.now()
	notifications, err := dispatcher.repository.ClaimByIDs(ctx, ids, now.UnixMilli(), now.Add(-notificationSendingLease).UnixMilli())
	if err != nil {
		return 0, err
	}
	return dispatcher.deliver(ctx, notifications, now.UnixMilli())
}

func (dispatcher *notificationDispatcher) DispatchDue(ctx context.Context, limit int) (int, error) {
	limit = normalizeNotificationDispatchLimit(limit)
	now := dispatcher.now()
	notifications, err := dispatcher.repository.ClaimDue(ctx, now.UnixMilli(), now.Add(-notificationSendingLease).UnixMilli(), limit)
	if err != nil {
		return 0, err
	}
	return dispatcher.deliver(ctx, notifications, now.UnixMilli())
}

func (dispatcher *notificationDispatcher) Retry(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("通知 ID 不能为空")
	}
	if err := dispatcher.repository.ResetForRetry(ctx, id, dispatcher.now().UnixMilli()); err != nil {
		return err
	}
	_, err := dispatcher.Dispatch(ctx, []string{id})
	return err
}

func (dispatcher *notificationDispatcher) deliver(ctx context.Context, notifications []NotificationRecord, now int64) (int, error) {
	byChannel := make(map[string][]NotificationRecord)
	var firstErr error
	for _, notification := range notifications {
		if notification.Channel == "in_app" {
			if err := dispatcher.repository.DeliverInApp(ctx, notification, now); err != nil {
				if markErr := dispatcher.markFailure(ctx, notification, err, now); markErr != nil && firstErr == nil {
					firstErr = markErr
				}
			}
			continue
		}
		byChannel[notification.Channel] = append(byChannel[notification.Channel], notification)
	}

	for channelName, channelNotifications := range byChannel {
		channel, exists := dispatcher.channels[channelName]
		if !exists {
			err := fmt.Errorf("%w: %s", ErrNotificationChannelUnavailable, channelName)
			for _, notification := range channelNotifications {
				if markErr := dispatcher.markFailure(ctx, notification, err, now); markErr != nil && firstErr == nil {
					firstErr = markErr
				}
			}
			continue
		}

		results := channel.Deliver(ctx, channelNotifications)
		resultByID := make(map[string]NotificationDeliveryResult, len(results))
		for _, result := range results {
			resultByID[result.ID] = result
		}
		for _, notification := range channelNotifications {
			result, exists := resultByID[notification.ID]
			if !exists {
				result = NotificationDeliveryResult{ID: notification.ID, Err: errors.New("通知渠道未返回投递结果")}
			}
			if result.Err != nil {
				if err := dispatcher.markFailure(ctx, notification, result.Err, now); err != nil && firstErr == nil {
					firstErr = err
				}
				continue
			}
			if err := dispatcher.repository.MarkSent(ctx, notification.ID, result.ProviderMessageID, now); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return len(notifications), firstErr
}

func (dispatcher *notificationDispatcher) markFailure(ctx context.Context, notification NotificationRecord, deliveryErr error, now int64) error {
	attempts := notification.Attempts + 1
	status := NotificationStatusFailed
	nextRetryAt := time.UnixMilli(now).Add(notificationRetryDelay(attempts)).UnixMilli()
	if attempts >= notificationMaxAttempts {
		status = NotificationStatusDead
		nextRetryAt = 0
	}
	return dispatcher.repository.MarkFailed(ctx, notification.ID, attempts, status, nextRetryAt, truncateNotificationError(deliveryErr), now)
}

func normalizeNotificationDispatchLimit(limit int) int {
	if limit <= 0 {
		return notificationDefaultDispatchLimit
	}
	if limit > notificationMaxDispatchLimit {
		return notificationMaxDispatchLimit
	}
	return limit
}

func notificationRetryDelay(attempts int) time.Duration {
	switch {
	case attempts <= 1:
		return time.Minute
	case attempts == 2:
		return 5 * time.Minute
	case attempts == 3:
		return 30 * time.Minute
	case attempts == 4:
		return 2 * time.Hour
	default:
		return 12 * time.Hour
	}
}

func truncateNotificationError(err error) string {
	if err == nil {
		return ""
	}
	runes := []rune(strings.TrimSpace(err.Error()))
	if len(runes) > notificationErrorMaxRunes {
		runes = runes[:notificationErrorMaxRunes]
	}
	return string(runes)
}
