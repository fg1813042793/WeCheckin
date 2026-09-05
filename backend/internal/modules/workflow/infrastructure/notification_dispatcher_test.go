package infrastructure

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/notificationstyle"
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

func TestNotificationRecordsUseRecipientNamesWithoutIDFallback(t *testing.T) {
	rows := []workflowmodel.NotificationOutbox{
		{ID: "named", RecipientUserID: "7", PayloadJSON: `{}`},
		{ID: "missing", RecipientUserID: "404", PayloadJSON: `{}`},
	}
	users := []model.User{{ID: 7, Name: " 张三 "}}

	records, err := notificationRecordsFromModelsWithUsers(rows, users)
	if err != nil {
		t.Fatalf("notificationRecordsFromModelsWithUsers() error = %v", err)
	}
	if records[0].RecipientUserName != "张三" || records[1].RecipientUserName != "" {
		t.Fatalf("notification recipient names = %#v", records)
	}
}

func TestWorkflowInAppNotifyUsesInstanceSourceContract(t *testing.T) {
	record := application.NotificationRecord{
		ID: "outbox-1", InstanceID: "instance-1", Kind: workflowmodel.NotificationKindTaskArrived,
		RecipientUserID: "7", Payload: application.NotificationPayload{Title: "待办", Content: "请处理"},
	}

	notify := workflowInAppNotify(record, 1234)
	if notify.Type != workflowmodel.NotificationKindTaskArrived || notify.SourceType != "workflow_instance" || notify.SourceID != "instance-1" {
		t.Fatalf("workflow notify source = type %q sourceType %q sourceID %q", notify.Type, notify.SourceType, notify.SourceID)
	}
	if notify.UserID != "7" || notify.Title != "待办" || notify.Content != "请处理" || notify.AddTime != 1234 {
		t.Fatalf("workflow notify payload = %#v", notify)
	}
	legacy := workflowInAppNotify(application.NotificationRecord{InstanceID: "instance-2"}, 1234)
	if legacy.Type != "workflow" {
		t.Fatalf("legacy workflow notify type = %q", legacy.Type)
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
		{ID: "outbox-1", InstanceID: "instance-1", Channel: workflowcore.NotificationChannelDingTalkOA, Payload: application.NotificationPayload{Title: "待办", Content: "请审批", MessageType: configsvc.DingTalkMessageTypeActionCard}},
		{ID: "outbox-2", InstanceID: "instance-1", Channel: workflowcore.NotificationChannelDingTalkOA, Payload: application.NotificationPayload{Title: "待办", Content: "请审批", MessageType: configsvc.DingTalkMessageTypeActionCard}},
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
	if call.payload.Title != "待办" || call.payload.Content != "请审批" || call.payload.URL != buildWorkflowNotificationURL(config.AppURL, config, "instance-1") || call.payload.SourceName != "WeCheckin 流程" || call.payload.MessageType != configsvc.DingTalkMessageTypeActionCard {
		t.Fatalf("payload = %#v", call.payload)
	}
	if len(results) != 3 || results[0].Err != nil || results[1].Err != nil || results[2].Err == nil {
		t.Fatalf("delivery results = %#v", results)
	}
}

func TestBuildWorkflowNotificationURLWrapsInstanceDeepLinkForDingTalk(t *testing.T) {
	got := buildWorkflowNotificationURL(
		"https://oa.example.com/dingtalk-h5/?corpId=corp-a",
		configsvc.DingTalkH5CorpConfig{CorpID: "ding-corp", AgentID: "123456"},
		"instance-1",
	)

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse workflow notification url: %v", err)
	}
	if parsed.Scheme != "dingtalk" || parsed.Host != "dingtalkclient" || parsed.Path != "/action/openapp" {
		t.Fatalf("workflow notification url = %q, want DingTalk openapp link", got)
	}
	query := parsed.Query()
	if query.Get("corpid") != "ding-corp" || query.Get("app_id") != "0_123456" {
		t.Fatalf("openapp query = %s", parsed.RawQuery)
	}
	redirect, err := url.Parse(query.Get("redirect_url"))
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if redirect.Query().Get("corpId") != "corp-a" {
		t.Fatalf("redirect corpId = %q, want corp-a", redirect.Query().Get("corpId"))
	}
	if redirect.Query().Get("view") != "workflow:instance:instance-1" {
		t.Fatalf("redirect view = %q, want workflow instance route", redirect.Query().Get("view"))
	}
}

func TestDingTalkNotificationGroupKeyIncludesTypeSpecificFields(t *testing.T) {
	base := configsvc.DingTalkWorkNotificationPayload{
		MessageType: configsvc.DingTalkMessageTypeActionCard,
		Title:       "审批待办",
		Content:     "请及时处理",
		URL:         "https://example.test/h5",
		MediaID:     "media-a",
		Duration:    10,
		ButtonTitle: "立即处理",
		HeadColor:   "FF1677FF",
	}
	baseKey := dingTalkNotificationGroupKey("corp-a", base)
	cases := []configsvc.DingTalkWorkNotificationPayload{
		func() configsvc.DingTalkWorkNotificationPayload {
			value := base
			value.MediaID = "media-b"
			return value
		}(),
		func() configsvc.DingTalkWorkNotificationPayload { value := base; value.Duration = 20; return value }(),
		func() configsvc.DingTalkWorkNotificationPayload {
			value := base
			value.ButtonTitle = "查看详情"
			return value
		}(),
		func() configsvc.DingTalkWorkNotificationPayload {
			value := base
			value.HeadColor = "FFFF0000"
			return value
		}(),
	}
	for index, payload := range cases {
		if key := dingTalkNotificationGroupKey("corp-a", payload); key == baseKey {
			t.Fatalf("case %d should produce a distinct group key", index)
		}
	}
}

func TestDingTalkNotificationChannelAppliesKindTemplate(t *testing.T) {
	config := configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AppURL: "https://example.test/h5", Enabled: 1, NotifyEnabled: 1}
	resolver := &dingTalkNotificationResolverStub{targets: map[string]DingTalkNotificationTarget{
		"outbox-1": {Config: config, DingTalkUserID: "ding-1", PicURL: "https://static.example.test/workflow-logo.png"},
	}, errors: map[string]error{}}
	client := &dingTalkNotificationClientStub{}
	styles := &workflowDingTalkStyleLoaderStub{config: notificationstyle.Config{Styles: []notificationstyle.Style{{
		Type: notificationstyle.TypeTaskArrived, Label: "待处理", Icon: "clock", Tone: notificationstyle.ToneWarning,
		DingTalk: notificationstyle.DingTalkTemplate{
			MessageType: notificationstyle.DingTalkMessageTypeActionCard,
			Title:       "{{title}}", Content: "## {{title}}\n{{content}}", URL: "{{url}}", ButtonTitle: "立即处理",
		},
	}}}}
	channel := newDingTalkNotificationChannel(client, resolver, styles)

	results := channel.Deliver(context.Background(), []application.NotificationRecord{{
		ID: "outbox-1", Kind: notificationstyle.TypeTaskArrived,
		Payload: application.NotificationPayload{Title: "审批待办", Content: "请及时处理"},
	}})
	if len(results) != 1 || results[0].Err != nil || len(client.calls) != 1 {
		t.Fatalf("delivery results=%#v calls=%d", results, len(client.calls))
	}
	payload := client.calls[0].payload
	if payload.MessageType != configsvc.DingTalkMessageTypeActionCard || payload.Content != "## 审批待办\n请及时处理" || payload.ButtonTitle != "立即处理" {
		t.Fatalf("templated workflow payload = %#v", payload)
	}
}

func TestDingTalkNotificationChannelAppliesWorkflowLogoToOATemplate(t *testing.T) {
	config := configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AppURL: "https://example.test/h5", Enabled: 1, NotifyEnabled: 1}
	resolver := &dingTalkNotificationResolverStub{targets: map[string]DingTalkNotificationTarget{
		"outbox-1": {Config: config, DingTalkUserID: "ding-1", PicURL: "https://static.example.test/workflow-logo.png"},
	}, errors: map[string]error{}}
	client := &dingTalkNotificationClientStub{}
	styles := &workflowDingTalkStyleLoaderStub{config: notificationstyle.Config{Styles: []notificationstyle.Style{{
		Type: notificationstyle.TypeTaskArrived, Label: "待处理", Icon: "clock", Tone: notificationstyle.ToneWarning,
		DingTalk: notificationstyle.DingTalkTemplate{
			MessageType: notificationstyle.DingTalkMessageTypeOA,
			Title:       "{{title}}", Content: "{{content}}", URL: "{{url}}", PicURL: "{{picUrl}}", SourceName: "{{sourceName}}",
		},
	}}}}
	channel := newDingTalkNotificationChannel(client, resolver, styles)

	results := channel.Deliver(context.Background(), []application.NotificationRecord{{
		ID: "outbox-1", Kind: notificationstyle.TypeTaskArrived,
		Payload: application.NotificationPayload{Title: "审批待办", Content: "请及时处理"},
	}})
	if len(results) != 1 || results[0].Err != nil || len(client.calls) != 1 {
		t.Fatalf("delivery results=%#v calls=%d", results, len(client.calls))
	}
	payload := client.calls[0].payload
	if payload.MessageType != configsvc.DingTalkMessageTypeOA || payload.PicURL != "https://static.example.test/workflow-logo.png" {
		t.Fatalf("templated workflow OA payload = %#v", payload)
	}
}

func TestWorkflowNotificationVersionMetadataLogo(t *testing.T) {
	logoURL, recorded, err := workflowNotificationVersionLogoURL(`{"name":"绩效审批","logoUrl":"/uploads/workflow-logos/performance.png"}`)
	if err != nil {
		t.Fatalf("workflowNotificationVersionLogoURL() error = %v", err)
	}
	if !recorded || logoURL != "/uploads/workflow-logos/performance.png" {
		t.Fatalf("logoURL = %q, recorded = %t", logoURL, recorded)
	}
	if _, recorded, err := workflowNotificationVersionLogoURL(""); err != nil || recorded {
		t.Fatalf("empty metadata should not be recorded: recorded=%t err=%v", recorded, err)
	}
	if _, _, err := workflowNotificationVersionLogoURL("{"); err == nil {
		t.Fatal("invalid metadata should fail")
	}
}

type workflowDingTalkStyleLoaderStub struct {
	config notificationstyle.Config
}

func (stub *workflowDingTalkStyleLoaderStub) NotificationStyles(context.Context) (notificationstyle.Config, error) {
	return stub.config, nil
}
