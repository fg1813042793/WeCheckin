package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	notificationmodel "wecheckin/backend/internal/model/notification"
)

const (
	defaultDispatchLimit = 100
	maxDispatchLimit     = 500
	sendingLease         = 10 * time.Minute
	maxAttempts          = 5
	maxErrorRunes        = 1000
	maxResponseRunes     = 4000
	responseTruncatedTag = "\n...[TRUNCATED]"
)

var (
	secretAssignmentPattern = regexp.MustCompile(`(?i)(access[_-]?token|app[_-]?secret|token)(\s*[=:]\s*)([^&\s"']+)`)
	bearerPattern           = regexp.MustCompile(`(?i)(bearer\s+)([a-z0-9._~+/=-]+)`)
	jsonSecretPattern       = regexp.MustCompile(`(?i)("(?:access[_-]?token|refresh[_-]?token|app[_-]?secret|client[_-]?secret|authorization|password|token)"\s*:\s*")[^"]*(")`)
)

type EnqueueRequest struct {
	IdempotencyKey string
	SourceType     string
	SourceID       string
	Channel        string
	Recipient      any
	Payload        any
}

type Failure struct {
	ID                 uint64
	Attempts           int
	Status             string
	NextRetryAt        int64
	LastError          string
	ResponseStatusCode int
	ResponseBody       string
	EditTime           int64
}

type DeliveryResult struct {
	ResponseStatusCode int
	ResponseBody       string
}

type Store interface {
	Enqueue(context.Context, notificationmodel.Outbox) (bool, error)
	ClaimDue(context.Context, int64, int64, int) ([]notificationmodel.Outbox, error)
	MarkSent(context.Context, uint64, int64, DeliveryResult) error
	MarkFailed(context.Context, Failure) error
}

type Channel interface {
	Name() string
	Deliver(context.Context, notificationmodel.Outbox) (DeliveryResult, error)
}

type Service struct {
	store    Store
	channels map[string]Channel
	now      func() time.Time
}

func NewService(store Store, channels ...Channel) *Service {
	return newService(store, time.Now, channels...)
}

func newService(store Store, now func() time.Time, channels ...Channel) *Service {
	registered := make(map[string]Channel, len(channels))
	for _, channel := range channels {
		if channel == nil || strings.TrimSpace(channel.Name()) == "" {
			continue
		}
		registered[strings.TrimSpace(channel.Name())] = channel
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, channels: registered, now: now}
}

func (service *Service) Enqueue(ctx context.Context, request EnqueueRequest) error {
	if service == nil || service.store == nil {
		return errors.New("notification outbox store is not initialized")
	}
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	request.SourceType = strings.TrimSpace(request.SourceType)
	request.SourceID = strings.TrimSpace(request.SourceID)
	request.Channel = strings.TrimSpace(request.Channel)
	if request.IdempotencyKey == "" || request.SourceType == "" || request.SourceID == "" || request.Channel == "" {
		return errors.New("notification idempotency key, source and channel are required")
	}
	recipient, err := json.Marshal(request.Recipient)
	if err != nil {
		return fmt.Errorf("encode notification recipient: %w", err)
	}
	payload, err := json.Marshal(request.Payload)
	if err != nil {
		return fmt.Errorf("encode notification payload: %w", err)
	}
	now := service.now().UnixMilli()
	_, err = service.store.Enqueue(ctx, notificationmodel.Outbox{
		IdempotencyKey: request.IdempotencyKey,
		SourceType:     request.SourceType,
		SourceID:       request.SourceID,
		Channel:        request.Channel,
		RecipientJSON:  string(recipient),
		PayloadJSON:    string(payload),
		Status:         notificationmodel.StatusPending,
		NextRetryAt:    now,
		AddTime:        now,
		EditTime:       now,
	})
	return err
}

func (service *Service) DispatchDue(ctx context.Context, limit int) (int, error) {
	if service == nil || service.store == nil {
		return 0, errors.New("notification outbox store is not initialized")
	}
	limit = normalizeDispatchLimit(limit)
	now := service.now()
	rows, err := service.store.ClaimDue(ctx, now.UnixMilli(), now.Add(-sendingLease).UnixMilli(), limit)
	if err != nil {
		return 0, err
	}
	var deliveryErrors []error
	for _, row := range rows {
		channel := service.channels[row.Channel]
		deliveryResult := DeliveryResult{}
		var deliveryErr error
		if channel == nil {
			deliveryErr = fmt.Errorf("notification channel %q is unavailable", row.Channel)
		} else {
			deliveryResult, deliveryErr = channel.Deliver(ctx, row)
		}
		deliveryResult = sanitizeDeliveryResult(deliveryResult)
		if deliveryErr == nil {
			if err := service.store.MarkSent(ctx, row.ID, now.UnixMilli(), deliveryResult); err != nil {
				deliveryErrors = append(deliveryErrors, err)
			}
			continue
		}
		attempts := row.Attempts + 1
		status := notificationmodel.StatusFailed
		nextRetryAt := now.Add(retryDelay(attempts)).UnixMilli()
		if attempts >= maxAttempts {
			status = notificationmodel.StatusDead
			nextRetryAt = 0
		}
		markErr := service.store.MarkFailed(ctx, Failure{
			ID: row.ID, Attempts: attempts, Status: status, NextRetryAt: nextRetryAt,
			LastError: sanitizeError(deliveryErr), ResponseStatusCode: deliveryResult.ResponseStatusCode,
			ResponseBody: deliveryResult.ResponseBody, EditTime: now.UnixMilli(),
		})
		if markErr != nil {
			deliveryErrors = append(deliveryErrors, errors.Join(deliveryErr, markErr))
		} else {
			deliveryErrors = append(deliveryErrors, deliveryErr)
		}
	}
	return len(rows), errors.Join(deliveryErrors...)
}

func normalizeDispatchLimit(limit int) int {
	if limit <= 0 {
		return defaultDispatchLimit
	}
	if limit > maxDispatchLimit {
		return maxDispatchLimit
	}
	return limit
}

func retryDelay(attempts int) time.Duration {
	switch attempts {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 30 * time.Minute
	case 4:
		return 2 * time.Hour
	default:
		return 12 * time.Hour
	}
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	message = redactSecrets(message)
	runes := []rune(message)
	if len(runes) > maxErrorRunes {
		runes = runes[:maxErrorRunes]
	}
	return string(runes)
}

func sanitizeDeliveryResult(result DeliveryResult) DeliveryResult {
	result.ResponseBody = sanitizeResponseBody(result.ResponseBody)
	return result
}

func sanitizeResponseBody(body string) string {
	message := strings.TrimSpace(strings.ToValidUTF8(body, "�"))
	message = redactSecrets(message)
	runes := []rune(message)
	marker := []rune(responseTruncatedTag)
	if len(runes) > maxResponseRunes {
		limit := maxResponseRunes - len(marker)
		if limit < 0 {
			limit = 0
		}
		runes = append(runes[:limit], marker...)
	}
	return string(runes)
}

func redactSecrets(message string) string {
	message = jsonSecretPattern.ReplaceAllString(message, "${1}[REDACTED]${2}")
	message = secretAssignmentPattern.ReplaceAllString(message, "${1}${2}[REDACTED]")
	return bearerPattern.ReplaceAllString(message, "${1}[REDACTED]")
}
