package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestDingTalkNotificationJobSendsWithRunIDAndLogsOnlyDeliveryCounts(t *testing.T) {
	sender := &dingTalkNotificationSenderStub{result: inappnotificationapp.SendResult{
		SourceID: "run-1", PlannedCount: 4, SentCount: 2, SkippedCount: 1, FailedCount: 1,
	}}
	logger := &captureRunLogger{}
	job := NewDingTalkNotificationJob(sender)
	params := json.RawMessage(`{"title":"定时钉钉通知","content":"敏感正文","scope":"users","userIds":[4,9]}`)

	result, err := job.Execute(context.Background(), "run-1", params, logger)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "notification.dingtalk.send" || job.Name() != "发送钉钉通知" {
		t.Fatalf("job metadata = %q/%q", job.Key(), job.Name())
	}
	if sender.input.SourceType != inappnotificationapp.SourceScheduledTaskRun || sender.input.SourceID != "run-1" {
		t.Fatalf("send source = %q/%q", sender.input.SourceType, sender.input.SourceID)
	}
	if result.Data["plannedCount"] != 4 || result.Data["sentCount"] != 2 || result.Data["skippedCount"] != 1 || result.Data["failedCount"] != 1 {
		t.Fatalf("result = %#v", result)
	}
	for _, sensitive := range []string{"定时钉钉通知", "敏感正文"} {
		if strings.Contains(logger.content, sensitive) {
			t.Fatalf("delivery log leaked notification content %q: %q", sensitive, logger.content)
		}
	}
	for _, count := range []string{"planned=4", "sent=2", "skipped=1", "failed=1"} {
		if !strings.Contains(logger.content, count) {
			t.Fatalf("delivery log missing %q: %q", count, logger.content)
		}
	}
}

func TestDingTalkNotificationJobValidatesRecipientConfiguration(t *testing.T) {
	job := NewDingTalkNotificationJob(&dingTalkNotificationSenderStub{})
	err := job.Validate(context.Background(), json.RawMessage(`{"title":"定时钉钉通知","content":"正文","scope":"departments"}`))
	if err == nil || !strings.Contains(err.Error(), "recipients") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDingTalkNotificationJobMapsMissingRecipients(t *testing.T) {
	job := NewDingTalkNotificationJob(&dingTalkNotificationSenderStub{err: inappnotificationapp.ErrNoRecipients})
	_, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"title":"通知","content":"正文","scope":"all"}`), nil)
	var handlerErr *scheduledtaskapp.HandlerError
	if !errors.As(err, &handlerErr) || handlerErr.Code != "no_recipients" {
		t.Fatalf("Execute() error = %v", err)
	}
}

type dingTalkNotificationSenderStub struct {
	input  inappnotificationapp.SendInput
	result inappnotificationapp.SendResult
	err    error
}

func (stub *dingTalkNotificationSenderStub) SendDingTalk(_ context.Context, input inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error) {
	stub.input = input
	return stub.result, stub.err
}
