package infrastructure

import (
	"context"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"wecheckin/backend/internal/modules/inappnotification/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/notificationstyle"
)

func TestDingTalkDeliveryGroupsRecipientsByCorpAndKeepsPartialFailures(t *testing.T) {
	resolver := &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{
		Targets: []dingTalkNotificationTarget{
			{LocalUserID: 4, DingTalkUserID: "ding-4", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", CorpName: "企业A", AppURL: "https://a.example/app"}},
			{LocalUserID: 9, DingTalkUserID: "ding-9", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", CorpName: "企业A", AppURL: "https://a.example/app"}},
			{LocalUserID: 11, DingTalkUserID: "ding-11", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-b", CorpName: "企业B", AppURL: "https://b.example/app"}},
		},
		SkippedCount: 1,
	}}
	client := &dingTalkClientStub{failures: map[string]error{"corp-b": errors.New("remote unavailable")}}
	delivery := newDingTalkDelivery(client, resolver)

	result, err := delivery.DeliverDingTalk(context.Background(), application.DingTalkDeliveryBatch{
		Title: "系统通知", Content: "维护公告", UserIDs: []uint{11, 9, 4},
	})
	if err != nil {
		t.Fatalf("DeliverDingTalk() error = %v", err)
	}
	if result.SentCount != 2 || result.SkippedCount != 1 || result.FailedCount != 1 {
		t.Fatalf("delivery result = %#v", result)
	}
	if len(client.calls) != 2 {
		t.Fatalf("client calls = %d, want 2", len(client.calls))
	}
	if !reflect.DeepEqual(client.calls[0].userIDs, []string{"ding-4", "ding-9"}) {
		t.Fatalf("corp-a recipients = %v", client.calls[0].userIDs)
	}
	if client.calls[0].payload.Title != "系统通知" || client.calls[0].payload.Content != "维护公告" || client.calls[0].payload.URL != "https://a.example/app" {
		t.Fatalf("corp-a payload = %#v", client.calls[0].payload)
	}
}

func TestDingTalkManualPayloadIncludesTitleInTextMode(t *testing.T) {
	payload := dingTalkManualNotificationPayload(configsvc.DingTalkH5CorpConfig{}, "系统通知", "维护公告")
	if payload.Title != "系统通知" || payload.Content != "系统通知\n维护公告" {
		t.Fatalf("text payload = %#v", payload)
	}
}

func TestDingTalkManualPayloadWrapsConfiguredAppURLForInAppOpen(t *testing.T) {
	payload := dingTalkManualNotificationPayload(configsvc.DingTalkH5CorpConfig{
		CorpID: "ding-corp", AgentID: "123456", AppURL: "https://oa.example.com/dingtalk-h5/?corpId=ding-corp",
	}, "测试通知", "测试内容")

	parsed, err := url.Parse(payload.URL)
	if err != nil {
		t.Fatalf("parse notification url: %v", err)
	}
	if parsed.Scheme != "dingtalk" || parsed.Host != "dingtalkclient" || parsed.Path != "/action/openapp" {
		t.Fatalf("notification url = %q, want DingTalk openapp link", payload.URL)
	}
	query := parsed.Query()
	if query.Get("corpid") != "ding-corp" || query.Get("app_id") != "0_123456" {
		t.Fatalf("openapp query = %s", parsed.RawQuery)
	}
	redirect, err := url.Parse(query.Get("redirect_url"))
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if redirect.Host != "oa.example.com" || redirect.Path != "/dingtalk-h5/" {
		t.Fatalf("redirect_url = %q", query.Get("redirect_url"))
	}
}

func TestDingTalkDeliveryAppliesConfiguredTemplateForNotificationType(t *testing.T) {
	resolver := &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{Targets: []dingTalkNotificationTarget{{
		LocalUserID: 4, DingTalkUserID: "ding-4", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AppURL: "https://example.test/app"},
	}}}}
	client := &dingTalkClientStub{}
	delivery := newDingTalkDelivery(client, resolver)
	delivery.styleLoader = &dingTalkStyleLoaderStub{config: notificationstyle.Config{Styles: []notificationstyle.Style{{
		Type: notificationstyle.TypeTaskReminder, Label: "处理提醒", Icon: "bell", Tone: notificationstyle.ToneWarning,
		DingTalk: notificationstyle.DingTalkTemplate{
			MessageType: notificationstyle.DingTalkMessageTypeMarkdown,
			Title:       "【提醒】{{title}}", Content: "## {{title}}\n{{content}}",
		},
	}}}}

	_, err := delivery.DeliverDingTalk(context.Background(), application.DingTalkDeliveryBatch{
		Title: "请审批", Content: "单据已到达", NotificationType: notificationstyle.TypeTaskReminder, UserIDs: []uint{4},
	})
	if err != nil {
		t.Fatalf("DeliverDingTalk() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("client calls = %d", len(client.calls))
	}
	payload := client.calls[0].payload
	if payload.MessageType != configsvc.DingTalkMessageTypeMarkdown || payload.Title != "【提醒】请审批" || payload.Content != "## 请审批\n单据已到达" {
		t.Fatalf("templated payload = %#v", payload)
	}
}

type dingTalkTargetResolverStub struct {
	resolution dingTalkTargetResolution
	err        error
	ctx        context.Context
	userIDs    []uint
	calls      int
}

type dingTalkStyleLoaderStub struct {
	config notificationstyle.Config
}

func (stub *dingTalkStyleLoaderStub) NotificationStyles(context.Context) (notificationstyle.Config, error) {
	return stub.config, nil
}

func (stub *dingTalkTargetResolverStub) ResolveTargets(ctx context.Context, userIDs []uint) (dingTalkTargetResolution, error) {
	stub.ctx, stub.userIDs = ctx, append([]uint(nil), userIDs...)
	stub.calls++
	return stub.resolution, stub.err
}

type dingTalkClientCall struct {
	corpID  string
	userIDs []string
	payload configsvc.DingTalkWorkNotificationPayload
}

type dingTalkClientStub struct {
	calls    []dingTalkClientCall
	failures map[string]error
}

func (stub *dingTalkClientStub) SendWorkNotificationContext(_ context.Context, config configsvc.DingTalkH5CorpConfig, userIDs []string, payload configsvc.DingTalkWorkNotificationPayload) error {
	stub.calls = append(stub.calls, dingTalkClientCall{corpID: config.CorpID, userIDs: append([]string(nil), userIDs...), payload: payload})
	return stub.failures[config.CorpID]
}

func TestDingTalkCanDeliverToChecksSingleReadyTargetWithoutSending(t *testing.T) {
	readyTarget := dingTalkNotificationTarget{
		LocalUserID: 7, DingTalkUserID: "ding-7",
		Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AppKey: "test-key", AppSecret: "test-secret", AgentID: "123", Enabled: 1, NotifyEnabled: 1},
	}
	for _, test := range []struct {
		name   string
		mutate func(*dingTalkNotificationTarget)
		want   bool
	}{
		{"ready agent", func(*dingTalkNotificationTarget) {}, true},
		{"wrong local user", func(target *dingTalkNotificationTarget) { target.LocalUserID = 9 }, false},
		{"blank binding", func(target *dingTalkNotificationTarget) { target.DingTalkUserID = " " }, false},
		{"blank corp", func(target *dingTalkNotificationTarget) { target.Config.CorpID = " " }, false},
		{"disabled corp", func(target *dingTalkNotificationTarget) { target.Config.Enabled = 0 }, false},
		{"disabled notifications", func(target *dingTalkNotificationTarget) { target.Config.NotifyEnabled = 0 }, false},
		{"missing key", func(target *dingTalkNotificationTarget) { target.Config.AppKey = " " }, false},
		{"missing secret", func(target *dingTalkNotificationTarget) { target.Config.AppSecret = " " }, false},
		{"missing agent", func(target *dingTalkNotificationTarget) { target.Config.AgentID = " " }, false},
		{"invalid agent", func(target *dingTalkNotificationTarget) { target.Config.AgentID = "invalid" }, false},
		{"zero agent", func(target *dingTalkNotificationTarget) { target.Config.AgentID = "0" }, false},
		{"robot defaults code to app key", func(target *dingTalkNotificationTarget) {
			target.Config.NotifyMode = "robot"
			target.Config.AgentID = ""
		}, true},
		{"robot inferred from code", func(target *dingTalkNotificationTarget) {
			target.Config.RobotCode = "robot-code"
			target.Config.AgentID = ""
		}, true},
		{"agent fallback", func(target *dingTalkNotificationTarget) { target.Config.NotifyMode = "agent-fallback" }, true},
		{"fallback needs valid agent", func(target *dingTalkNotificationTarget) {
			target.Config.NotifyMode = "agent_fallback"
			target.Config.AgentID = ""
		}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := readyTarget
			test.mutate(&target)
			resolver := &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{Targets: []dingTalkNotificationTarget{target}}}
			client := &dingTalkClientStub{}
			delivery := newDingTalkDelivery(client, resolver)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, err := delivery.CanDeliverTo(ctx, 7)
			if err != nil || got != test.want {
				t.Fatalf("ready=%v want=%v err=%v", got, test.want, err)
			}
			if resolver.calls != 1 || resolver.ctx != ctx || !reflect.DeepEqual(resolver.userIDs, []uint{7}) {
				t.Fatalf("resolver calls=%d userIDs=%v", resolver.calls, resolver.userIDs)
			}
			if len(client.calls) != 0 {
				t.Fatal("preflight sent a notification")
			}
		})
	}
	for _, resolution := range []dingTalkTargetResolution{
		{}, {SkippedCount: 1}, {Targets: []dingTalkNotificationTarget{readyTarget, readyTarget}},
		{Targets: []dingTalkNotificationTarget{readyTarget}, SkippedCount: 1},
	} {
		client := &dingTalkClientStub{}
		delivery := newDingTalkDelivery(client, &dingTalkTargetResolverStub{resolution: resolution})
		if ready, err := delivery.CanDeliverTo(context.Background(), 7); ready || err != nil || len(client.calls) != 0 {
			t.Fatalf("ready=%v err=%v sends=%d", ready, err, len(client.calls))
		}
	}
}

func TestDingTalkCanDeliverToHandlesUnavailableAndResolverErrors(t *testing.T) {
	for _, delivery := range []*DingTalkDelivery{nil, {}, {client: &dingTalkClientStub{}}, {resolver: &dingTalkTargetResolverStub{}}} {
		if ready, err := delivery.CanDeliverTo(context.Background(), 7); ready || !errors.Is(err, application.ErrDingTalkDeliveryUnavailable) {
			t.Fatalf("ready=%v err=%v", ready, err)
		}
	}
	resolver := &dingTalkTargetResolverStub{err: errors.New("resolver unavailable")}
	client := &dingTalkClientStub{}
	delivery := newDingTalkDelivery(client, resolver)
	if ready, err := delivery.CanDeliverTo(context.Background(), 0); ready || err != nil || resolver.calls != 0 {
		t.Fatalf("zero user: ready=%v err=%v calls=%d", ready, err, resolver.calls)
	}
	if ready, err := delivery.CanDeliverTo(context.Background(), 7); ready || !errors.Is(err, resolver.err) {
		t.Fatalf("ready=%v err=%v", ready, err)
	}
	if len(client.calls) != 0 {
		t.Fatal("preflight sent a notification")
	}
}
