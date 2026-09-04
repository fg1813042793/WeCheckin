package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
)

func TestInAppNotificationJobSendsWithRunIDAndLogsOnlyDeliveryCounts(t *testing.T) {
	sender := &notificationSenderStub{result: inappnotificationapp.SendResult{
		SourceID: "run-1", PlannedCount: 3, SentCount: 2, SkippedCount: 1,
	}}
	logger := &captureRunLogger{}
	job := NewInAppNotificationJob(sender)
	params := json.RawMessage(`{"title":"定时通知","content":"敏感正文","scope":"departments","departmentIds":[3]}`)

	result, err := job.Execute(context.Background(), "run-1", params, logger)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "notification.in_app.send" || job.Name() != "发送站内信" {
		t.Fatalf("job metadata = %q/%q", job.Key(), job.Name())
	}
	if sender.input.SourceType != inappnotificationapp.SourceScheduledTaskRun || sender.input.SourceID != "run-1" {
		t.Fatalf("send source = %q/%q", sender.input.SourceType, sender.input.SourceID)
	}
	if result.Data["sentCount"] != 2 || !strings.Contains(logger.content, "sent=2") {
		t.Fatalf("result=%#v log=%q", result, logger.content)
	}
	if strings.Contains(logger.content, "敏感正文") || strings.Contains(logger.content, "定时通知") {
		t.Fatalf("delivery log leaked notification body: %q", logger.content)
	}
}

func TestInAppNotificationJobValidatesRecipientConfiguration(t *testing.T) {
	job := NewInAppNotificationJob(&notificationSenderStub{})
	err := job.Validate(context.Background(), json.RawMessage(`{"title":"定时通知","content":"正文","scope":"users"}`))
	if err == nil || !strings.Contains(err.Error(), "recipients") {
		t.Fatalf("Validate() error = %v", err)
	}
}

type notificationSenderStub struct {
	input  inappnotificationapp.SendInput
	result inappnotificationapp.SendResult
	err    error
}

func (stub *notificationSenderStub) Send(_ context.Context, input inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error) {
	stub.input = input
	return stub.result, stub.err
}

type captureRunLogger struct {
	level   string
	stage   string
	content string
}

func (logger *captureRunLogger) Log(_ context.Context, level, stage, content string) error {
	logger.level = level
	logger.stage = stage
	logger.content = content
	return nil
}
