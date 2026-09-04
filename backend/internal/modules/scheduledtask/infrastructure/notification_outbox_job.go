package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type NotificationOutboxDispatcher interface {
	DispatchDue(context.Context, int) (int, error)
}

type NotificationOutboxDispatchJob struct {
	dispatcher NotificationOutboxDispatcher
}

func NewNotificationOutboxDispatchJob(dispatcher NotificationOutboxDispatcher) *NotificationOutboxDispatchJob {
	return &NotificationOutboxDispatchJob{dispatcher: dispatcher}
}

func (*NotificationOutboxDispatchJob) Key() string { return "notification.outbox.dispatch_due" }

func (*NotificationOutboxDispatchJob) Name() string { return "派发通用通知 Outbox" }

func (*NotificationOutboxDispatchJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","minimum":1,"maximum":500}}}`)
}

func (*NotificationOutboxDispatchJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeOutboxDispatchParams(raw)
	return err
}

func (job *NotificationOutboxDispatchJob) Execute(ctx context.Context, _ string, raw json.RawMessage, _ application.RunLogger) (application.HandlerResult, error) {
	if job == nil || job.dispatcher == nil {
		return application.HandlerResult{}, errors.New("notification outbox dispatcher is not initialized")
	}
	params, err := decodeOutboxDispatchParams(raw)
	if err != nil {
		return application.HandlerResult{}, err
	}
	count, err := job.dispatcher.DispatchDue(ctx, params.Limit)
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{
			Code: "notification_outbox_dispatch_failed", Summary: err.Error(), Temporary: true,
		}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("dispatched %d notification outbox records", count),
		Data:    map[string]any{"dispatched": count},
	}, nil
}

type outboxDispatchParams struct {
	Limit int `json:"limit"`
}

func decodeOutboxDispatchParams(raw json.RawMessage) (outboxDispatchParams, error) {
	params := outboxDispatchParams{Limit: 100}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, fmt.Errorf("decode notification outbox dispatch params: %w", err)
		}
	}
	if params.Limit < 1 || params.Limit > 500 {
		return params, errors.New("notification outbox dispatch limit must be between 1 and 500")
	}
	return params, nil
}

var _ GoJob = (*NotificationOutboxDispatchJob)(nil)
