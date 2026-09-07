package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
)

const DingTalkNotificationJobKey = "notification.dingtalk.send"

type DingTalkNotificationSender interface {
	SendDingTalk(context.Context, inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error)
}

type DingTalkNotificationJob struct {
	sender DingTalkNotificationSender
}

func NewDingTalkNotificationJob(sender DingTalkNotificationSender) *DingTalkNotificationJob {
	return &DingTalkNotificationJob{sender: sender}
}

func (job *DingTalkNotificationJob) Key() string  { return DingTalkNotificationJobKey }
func (job *DingTalkNotificationJob) Name() string { return "发送钉钉通知" }
func (job *DingTalkNotificationJob) ConfigSchema() json.RawMessage {
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

func (job *DingTalkNotificationJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeInAppNotificationParams(raw, "validation")
	return err
}

func (job *DingTalkNotificationJob) Execute(ctx context.Context, runID string, raw json.RawMessage, logger scheduledtaskapp.RunLogger) (scheduledtaskapp.HandlerResult, error) {
	if job == nil || job.sender == nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "handler_unavailable", Summary: "DingTalk notification service is not initialized"}
	}
	input, err := decodeInAppNotificationParams(raw, runID)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	result, err := job.sender.SendDingTalk(ctx, input)
	if err != nil {
		if errors.Is(err, inappnotificationapp.ErrNoRecipients) {
			return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "no_recipients", Summary: err.Error()}
		}
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "notification_delivery_failed", Summary: err.Error(), Temporary: true}
	}
	if logger == nil {
		logger = scheduledtaskapp.NopRunLogger{}
	}
	_ = logger.Log(ctx, "info", "delivery", fmt.Sprintf(
		"planned=%d sent=%d skipped=%d failed=%d",
		result.PlannedCount,
		result.SentCount,
		result.SkippedCount,
		result.FailedCount,
	))
	return scheduledtaskapp.HandlerResult{
		Summary: fmt.Sprintf("sent %d DingTalk notifications", result.SentCount),
		Data: map[string]interface{}{
			"sourceId":     result.SourceID,
			"plannedCount": result.PlannedCount,
			"sentCount":    result.SentCount,
			"skippedCount": result.SkippedCount,
			"failedCount":  result.FailedCount,
		},
	}, nil
}

var _ GoJob = (*DingTalkNotificationJob)(nil)
