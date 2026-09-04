package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
)

const InAppNotificationJobKey = "notification.in_app.send"

type InAppNotificationSender interface {
	Send(context.Context, inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error)
}

type InAppNotificationJob struct {
	sender InAppNotificationSender
}

func NewInAppNotificationJob(sender InAppNotificationSender) *InAppNotificationJob {
	return &InAppNotificationJob{sender: sender}
}

func (job *InAppNotificationJob) Key() string  { return InAppNotificationJobKey }
func (job *InAppNotificationJob) Name() string { return "发送站内信" }
func (job *InAppNotificationJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"required":["title","content","scope"],
		"properties":{
			"title":{"type":"string","minLength":1,"maxLength":255},
			"content":{"type":"string","minLength":1,"maxLength":5000},
			"scope":{"type":"string","enum":["all","departments","users"]},
			"userIds":{"type":"array","uniqueItems":true,"items":{"type":"integer","minimum":1}},
			"departmentIds":{"type":"array","uniqueItems":true,"items":{"type":"integer","minimum":1}}
		}
	}`)
}

func (job *InAppNotificationJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeInAppNotificationParams(raw, "validation")
	return err
}

func (job *InAppNotificationJob) Execute(ctx context.Context, runID string, raw json.RawMessage, logger scheduledtaskapp.RunLogger) (scheduledtaskapp.HandlerResult, error) {
	if job == nil || job.sender == nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "handler_unavailable", Summary: "in-app notification service is not initialized"}
	}
	input, err := decodeInAppNotificationParams(raw, runID)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	result, err := job.sender.Send(ctx, input)
	if err != nil {
		if errors.Is(err, inappnotificationapp.ErrNoRecipients) {
			return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "no_recipients", Summary: err.Error()}
		}
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "notification_delivery_failed", Summary: err.Error(), Temporary: true}
	}
	if logger == nil {
		logger = scheduledtaskapp.NopRunLogger{}
	}
	logContent := fmt.Sprintf(
		"planned=%d sent=%d skipped=%d replayed=%t",
		result.PlannedCount,
		result.SentCount,
		result.SkippedCount,
		result.Replayed,
	)
	_ = logger.Log(ctx, "info", "delivery", logContent)
	return scheduledtaskapp.HandlerResult{
		Summary: fmt.Sprintf("sent %d in-app notifications", result.SentCount),
		Data: map[string]interface{}{
			"sourceId":     result.SourceID,
			"plannedCount": result.PlannedCount,
			"sentCount":    result.SentCount,
			"skippedCount": result.SkippedCount,
			"replayed":     result.Replayed,
		},
	}, nil
}

func decodeInAppNotificationParams(raw json.RawMessage, sourceID string) (inappnotificationapp.SendInput, error) {
	var input inappnotificationapp.SendInput
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return input, fmt.Errorf("decode in-app notification params: %w", err)
	}
	input.SourceType = inappnotificationapp.SourceScheduledTaskRun
	input.SourceID = sourceID
	if err := inappnotificationapp.ValidateSendInput(input); err != nil {
		return input, err
	}
	return input, nil
}

var _ GoJob = (*InAppNotificationJob)(nil)
