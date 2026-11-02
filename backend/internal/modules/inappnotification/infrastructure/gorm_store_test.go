package infrastructure

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	notificationmodel "wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/modules/inappnotification/application"
)

func TestExpandDepartmentIDsIncludesSelectedDepartmentsAndDescendants(t *testing.T) {
	departments := []model.Department{
		{ID: 1, ParentID: 0},
		{ID: 2, ParentID: 1},
		{ID: 3, ParentID: 2},
		{ID: 4, ParentID: 0},
	}

	got := expandDepartmentIDs([]uint{2}, departments)
	if want := []uint{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expandDepartmentIDs() = %v, want %v", got, want)
	}
}

func TestDingTalkRecordKeyIsStableAndSourceSpecific(t *testing.T) {
	first := dingTalkRecordKey("scheduled_task_run", "run-1")
	second := dingTalkRecordKey("scheduled_task_run", "run-1")
	other := dingTalkRecordKey("scheduled_task_run", "run-2")
	if first != second || first == other || len(first) != 64 {
		t.Fatalf("keys first=%q second=%q other=%q", first, second, other)
	}
}

func TestDingTalkRecordFromOutboxDecodesCountsAndTemplate(t *testing.T) {
	record := dingTalkRecordFromOutbox(notificationmodel.Outbox{
		ID: 7, SourceType: "scheduled_task_run", SourceID: "run-1", Status: application.DingTalkDeliveryStatusPartial,
		RecipientJSON: `{"plannedCount":4,"sentCount":2,"skippedCount":1,"failedCount":1}`,
		PayloadJSON:   `{"title":"月度绩效","content":"请填写","notificationType":"scheduled_task","templateId":9,"templateVersion":3,"templateName":"绩效提醒"}`,
		Attempts:      2, SentAt: 100, AddTime: 90, EditTime: 100,
	})
	if record.ID != 7 || record.SourceID != "run-1" || record.Status != application.DingTalkDeliveryStatusPartial ||
		record.PlannedCount != 4 || record.SentCount != 2 || record.SkippedCount != 1 || record.FailedCount != 1 ||
		record.TemplateID != 9 || record.TemplateVersion != 3 || record.TemplateName != "绩效提醒" {
		t.Fatalf("record=%#v", record)
	}
}

func TestDingTalkRecordFromOutboxReadsFeedbackQueueCounts(t *testing.T) {
	for _, test := range []struct {
		name, status, response string
		want                   dingTalkRecordCounts
	}{
		{"pending", notificationmodel.StatusPending, "", dingTalkRecordCounts{PlannedCount: 1}},
		{"sending", notificationmodel.StatusSending, "", dingTalkRecordCounts{PlannedCount: 1}},
		{"sent", notificationmodel.StatusSent, `{"plannedCount":1,"sentCount":1,"skippedCount":0,"failedCount":0}`, dingTalkRecordCounts{PlannedCount: 1, SentCount: 1}},
		{"failed", notificationmodel.StatusFailed, `{"plannedCount":1,"sentCount":0,"skippedCount":0,"failedCount":1}`, dingTalkRecordCounts{PlannedCount: 1, FailedCount: 1}},
		{"skipped", notificationmodel.StatusFailed, `{"plannedCount":1,"sentCount":0,"skippedCount":1,"failedCount":0}`, dingTalkRecordCounts{PlannedCount: 1, SkippedCount: 1}},
		{"dead", notificationmodel.StatusDead, `{"plannedCount":1,"sentCount":0,"skippedCount":0,"failedCount":1}`, dingTalkRecordCounts{PlannedCount: 1, FailedCount: 1}},
		{"malformed response", notificationmodel.StatusFailed, `{"sentCount":1,`, dingTalkRecordCounts{PlannedCount: 1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			record := dingTalkRecordFromOutbox(notificationmodel.Outbox{
				ID: 8, Channel: notificationmodel.ChannelDingTalk, SourceType: "user_feedback", SourceID: "91", Status: test.status,
				RecipientJSON: `{"userIds":[7]}`, ResponseBody: test.response,
				PayloadJSON: `{"title":"反馈新消息","content":"问题已修复","sourceType":"user_feedback","sourceId":"91","notificationType":"feedback_status"}`,
			})
			got := dingTalkRecordCounts{PlannedCount: record.PlannedCount, SentCount: record.SentCount, SkippedCount: record.SkippedCount, FailedCount: record.FailedCount}
			if got != test.want || record.Status != test.status || record.Title != "反馈新消息" || record.NotificationType != "feedback_status" {
				t.Fatalf("record=%#v want counts=%#v", record, test.want)
			}
		})
	}
}

func TestDingTalkRecordFromOutboxKeepsLegacyCountsAuthoritative(t *testing.T) {
	for _, sourceType := range []string{"scheduled_task_run", "user_feedback"} {
		record := dingTalkRecordFromOutbox(notificationmodel.Outbox{
			Channel: notificationmodel.ChannelDingTalk, SourceType: sourceType,
			RecipientJSON: `{"plannedCount":4,"sentCount":2,"skippedCount":1,"failedCount":1}`,
			ResponseBody:  `{"plannedCount":1,"sentCount":1,"skippedCount":0,"failedCount":0}`,
		})
		if record.PlannedCount != 4 || record.SentCount != 2 || record.SkippedCount != 1 || record.FailedCount != 1 {
			t.Fatalf("legacy record=%#v", record)
		}
	}
}

func TestExpandDepartmentIDsIgnoresMissingAndDuplicateSelections(t *testing.T) {
	departments := []model.Department{{ID: 2, ParentID: 1}, {ID: 3, ParentID: 2}}

	got := expandDepartmentIDs([]uint{2, 2, 99}, departments)
	if want := []uint{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expandDepartmentIDs() = %v, want %v", got, want)
	}
}

func TestDeliveryKeyIsStableAndRecipientSpecific(t *testing.T) {
	first := deliveryKey("scheduled_task_run", "run-1", 7)
	second := deliveryKey("scheduled_task_run", "run-1", 7)
	otherRecipient := deliveryKey("scheduled_task_run", "run-1", 8)

	if first != second || first == otherRecipient || len(first) != 64 {
		t.Fatalf("delivery keys first=%q second=%q other=%q", first, second, otherRecipient)
	}
}

func TestBuildDepartmentOptionsKeepsHierarchyAndPromotesOrphans(t *testing.T) {
	departments := []model.Department{
		{ID: 2, Name: "研发", ParentID: 1},
		{ID: 3, Name: "平台组", ParentID: 2},
		{ID: 4, Name: "行政", ParentID: 0},
	}

	got := buildDepartmentOptions(departments)
	if len(got) != 2 || got[0].ID != 2 || len(got[0].Children) != 1 || got[0].Children[0].ID != 3 || got[1].ID != 4 {
		t.Fatalf("department options = %#v", got)
	}
}
