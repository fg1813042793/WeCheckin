package application

import (
	"testing"

	notificationmodel "wecheckin/backend/internal/model/notification"
)

func TestWebhookRecordFromOutboxMasksTargetAndDecodesPayload(t *testing.T) {
	record := webhookRecordFromOutbox(notificationmodel.Outbox{
		ID: 7, SourceType: "survey_response", SourceID: "42", Status: notificationmodel.StatusSent,
		Attempts: 1, RecipientJSON: `{"type":"dingtalk","url":"https://example.com/hook?access_token=secret"}`,
		PayloadJSON: `{"title":"问卷统计通知: DISC","content":"D 个数 2"}`,
	})

	if record.ID != 7 || record.Title != "问卷统计通知: DISC" || record.WebhookType != "dingtalk" {
		t.Fatalf("record = %#v", record)
	}
	if record.Target != "https://example.com/***" {
		t.Fatalf("target = %q", record.Target)
	}
}
