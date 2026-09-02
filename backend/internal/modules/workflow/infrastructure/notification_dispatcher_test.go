package infrastructure

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/workflowcore"
)

type dingTalkNotificationResolverStub struct {
	targets map[string]DingTalkNotificationTarget
	errors  map[string]error
}

func (stub *dingTalkNotificationResolverStub) Resolve(_ context.Context, notification application.NotificationRecord) (DingTalkNotificationTarget, error) {
	if err := stub.errors[notification.ID]; err != nil {
		return DingTalkNotificationTarget{}, err
	}
	return stub.targets[notification.ID], nil
}

type dingTalkNotificationClientStub struct {
	calls []dingTalkNotificationCall
}

type dingTalkNotificationCall struct {
	config  configsvc.DingTalkH5CorpConfig
	userIDs []string
	payload configsvc.DingTalkWorkNotificationPayload
}

func (stub *dingTalkNotificationClientStub) SendWorkNotificationContext(_ context.Context, config configsvc.DingTalkH5CorpConfig, userIDs []string, payload configsvc.DingTalkWorkNotificationPayload) error {
	stub.calls = append(stub.calls, dingTalkNotificationCall{config: config, userIDs: append([]string(nil), userIDs...), payload: payload})
	return nil
}

func TestNotificationRecordFromModelDecodesPayloadSnapshot(t *testing.T) {
	record, err := notificationRecordFromModel(workflowmodel.NotificationOutbox{
		ID: "outbox-1", InstanceID: "instance-1", RecipientUserID: "7",
		PayloadJSON: `{"title":"待办","content":"请处理","starterId":"3"}`,
	})
	if err != nil {
		t.Fatalf("notificationRecordFromModel() error = %v", err)
	}
	if record.Payload.Title != "待办" || record.Payload.Content != "请处理" || record.Payload.StarterID != "3" {
		t.Fatalf("payload = %#v", record.Payload)
	}
	if _, err := notificationRecordFromModel(workflowmodel.NotificationOutbox{ID: "bad", PayloadJSON: "{"}); err == nil {
		t.Fatal("invalid payload snapshot should fail")
	}
}

func TestWorkflowInAppNotifyUsesInstanceSourceContract(t *testing.T) {
	record := application.NotificationRecord{
		ID: "outbox-1", InstanceID: "instance-1", Kind: workflowmodel.NotificationKindTaskArrived,
		RecipientUserID: "7", Payload: application.NotificationPayload{Title: "待办", Content: "请处理"},
	}

	notify := workflowInAppNotify(record, 1234)
	if notify.Type != "workflow" || notify.SourceType != "workflow_instance" || notify.SourceID != "instance-1" {
		t.Fatalf("workflow notify source = type %q sourceType %q sourceID %q", notify.Type, notify.SourceType, notify.SourceID)
	}
	if notify.UserID != "7" || notify.Title != "待办" || notify.Content != "请处理" || notify.AddTime != 1234 {
		t.Fatalf("workflow notify payload = %#v", notify)
	}
}

func TestSelectDingTalkNotificationTargetPrefersSmallestCommonEnabledCorp(t *testing.T) {
	bindings := []model.DingTalkH5UserBinding{
		{CorpID: "corp-c", DingTalkUserID: "user-c", Enabled: 1},
		{CorpID: "corp-b", DingTalkUserID: "user-b", Enabled: 1},
		{CorpID: "corp-a", DingTalkUserID: "user-a", Enabled: 1},
	}
	configs := map[string]model.DingTalkH5CorpConfig{
		"corp-a": {CorpID: "corp-a", Enabled: 1, NotifyEnabled: 1},
		"corp-b": {CorpID: "corp-b", Enabled: 1, NotifyEnabled: 1},
		"corp-c": {CorpID: "corp-c", Enabled: 1, NotifyEnabled: 1},
	}
	binding, config, err := selectDingTalkNotificationTarget([]string{"corp-b", "corp-a"}, bindings, configs)
	if err != nil {
		t.Fatalf("selectDingTalkNotificationTarget() error = %v", err)
	}
	if binding.CorpID != "corp-a" || binding.DingTalkUserID != "user-a" || config.CorpID != "corp-a" {
		t.Fatalf("selected binding/config = %#v %#v", binding, config)
	}
}

func TestSelectDingTalkNotificationTargetFallsBackToRecipientCorp(t *testing.T) {
	bindings := []model.DingTalkH5UserBinding{
		{CorpID: "corp-c", DingTalkUserID: "user-c", Enabled: 1},
		{CorpID: "corp-a", DingTalkUserID: "user-a", Enabled: 1},
	}
	configs := map[string]model.DingTalkH5CorpConfig{
		"corp-a": {CorpID: "corp-a", Enabled: 1, NotifyEnabled: 0},
		"corp-c": {CorpID: "corp-c", Enabled: 1, NotifyEnabled: 1},
	}
	binding, _, err := selectDingTalkNotificationTarget([]string{"corp-z"}, bindings, configs)
	if err != nil {
		t.Fatalf("selectDingTalkNotificationTarget() error = %v", err)
	}
	if binding.CorpID != "corp-c" {
		t.Fatalf("selected binding = %#v", binding)
	}
}

func TestDingTalkNotificationChannelBatchesSameCorpAndPayload(t *testing.T) {
	config := configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AppKey: "key", AppSecret: "secret", AgentID: "123", AppURL: "https://example.test/h5", Enabled: 1, NotifyEnabled: 1}
	resolver := &dingTalkNotificationResolverStub{
		targets: map[string]DingTalkNotificationTarget{
			"outbox-1": {Config: config, DingTalkUserID: "ding-1"},
			"outbox-2": {Config: config, DingTalkUserID: "ding-2"},
		},
		errors: map[string]error{"outbox-bad": errors.New("用户未绑定钉钉")},
	}
	client := &dingTalkNotificationClientStub{}
	channel := newDingTalkNotificationChannel(client, resolver)
	notifications := []application.NotificationRecord{
		{ID: "outbox-1", Channel: workflowcore.NotificationChannelDingTalkOA, Payload: application.NotificationPayload{Title: "待办", Content: "请审批"}},
		{ID: "outbox-2", Channel: workflowcore.NotificationChannelDingTalkOA, Payload: application.NotificationPayload{Title: "待办", Content: "请审批"}},
		{ID: "outbox-bad", Channel: workflowcore.NotificationChannelDingTalkOA},
	}

	results := channel.Deliver(context.Background(), notifications)
	if len(client.calls) != 1 {
		t.Fatalf("DingTalk calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if !reflect.DeepEqual(call.userIDs, []string{"ding-1", "ding-2"}) {
		t.Fatalf("batched user IDs = %#v", call.userIDs)
	}
	if call.payload.Title != "待办" || call.payload.Content != "请审批" || call.payload.URL != config.AppURL || call.payload.SourceName != "WeCheckin 流程" {
		t.Fatalf("payload = %#v", call.payload)
	}
	if len(results) != 3 || results[0].Err != nil || results[1].Err != nil || results[2].Err == nil {
		t.Fatalf("delivery results = %#v", results)
	}
}
