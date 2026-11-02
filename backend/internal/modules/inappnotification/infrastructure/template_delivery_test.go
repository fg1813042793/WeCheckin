package infrastructure

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"wecheckin/backend/internal/modules/inappnotification/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/notificationstyle"
)

func TestTemplateDeliveryKeepsWorkflowAndCorpContextAcrossGroups(t *testing.T) {
	resolver := &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{Targets: []dingTalkNotificationTarget{
		{LocalUserID: 1, DingTalkUserID: "one", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AgentID: "101", AppURL: "https://example.test/app?tenant=a&corpId=old&view=old&key=wrong&page=outdated#/old"}},
		{LocalUserID: 2, DingTalkUserID: "two", Config: configsvc.DingTalkH5CorpConfig{CorpID: "corp-b", UnifiedAppID: "app-b", AppURL: "https://example.test/app?tenant=b"}},
	}}}
	template := &application.DingTalkSendTemplate{ID: 7, Version: 3, MessageType: "oa", Target: application.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17}, HeadColor: "FFAABBCC", ButtonTitle: "打开页面"}
	client := &templateMutatingClient{change: func() { template.Target.DefinitionID = 999; template.HeadColor = "FFFFFFFF" }}
	delivery := newDingTalkDelivery(client, resolver)
	delivery.styleLoader = rejectTemplateStyleLoader{t}
	result, err := delivery.DeliverDingTalk(context.Background(), application.DingTalkDeliveryBatch{Template: template, Title: "临时标题", Content: "临时正文", UserIDs: []uint{1, 2}})
	if err != nil || result.SentCount != 2 || len(client.calls) != 2 {
		t.Fatalf("result=%#v calls=%d err=%v", result, len(client.calls), err)
	}
	for i, call := range client.calls {
		parsed, err := url.Parse(call.payload.URL)
		if err != nil {
			t.Fatal(err)
		}
		redirect, err := url.Parse(parsed.Query().Get("redirect_url"))
		if err != nil {
			t.Fatal(err)
		}
		if redirect.Fragment != "/pages/index/index?view=workflow:start:17" || redirect.Query().Get("corpId") != call.corpID || redirect.Query().Get("view") != "workflow:start:17" || redirect.Query().Has("key") || redirect.Query().Has("page") || redirect.Query().Get("tenant") != []string{"a", "b"}[i] || parsed.Query().Get("corpid") != call.corpID {
			t.Fatalf("redirect=%s wrapper=%s", redirect, parsed)
		}
		if call.payload.MessageType != "oa" || call.payload.Title != "临时标题" || call.payload.Content != "临时正文" || call.payload.HeadColor != "FFAABBCC" {
			t.Fatalf("payload=%#v", call.payload)
		}
	}
}

func TestTemplateDeliveryPreservesCustomURLAndNeverFallsBack(t *testing.T) {
	for _, messageType := range []string{"link", "action_card"} {
		t.Run(messageType, func(t *testing.T) {
			client := &dingTalkClientStub{}
			delivery := newDingTalkDelivery(client, &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{Targets: []dingTalkNotificationTarget{{LocalUserID: 1, DingTalkUserID: "one", Config: configsvc.DingTalkH5CorpConfig{CorpID: "a", AgentID: "1"}}}}})
			target := "https://example.test/a?x=%2F&y=中文#anchor"
			_, err := delivery.DeliverDingTalk(context.Background(), application.DingTalkDeliveryBatch{Template: &application.DingTalkSendTemplate{MessageType: messageType, Target: application.DingTalkTemplateTarget{Kind: "url", URL: target}, ButtonTitle: "去填写"}, UserIDs: []uint{1}})
			if err != nil {
				t.Fatal(err)
			}
			parsed, _ := url.Parse(client.calls[0].payload.URL)
			if parsed.Query().Get("redirect_url") != target || client.calls[0].payload.MessageType != messageType || client.calls[0].payload.ButtonTitle != "去填写" {
				t.Fatalf("payload=%#v", client.calls[0].payload)
			}
		})
	}
	client := &dingTalkClientStub{}
	delivery := newDingTalkDelivery(client, &dingTalkTargetResolverStub{resolution: dingTalkTargetResolution{Targets: []dingTalkNotificationTarget{
		{LocalUserID: 1, DingTalkUserID: "one", Config: configsvc.DingTalkH5CorpConfig{CorpID: "a", AgentID: "1", AppURL: "https://example.test"}},
		{LocalUserID: 2, DingTalkUserID: "two", Config: configsvc.DingTalkH5CorpConfig{CorpID: "b", AgentID: "2"}},
	}}})
	_, err := delivery.DeliverDingTalk(context.Background(), application.DingTalkDeliveryBatch{Template: &application.DingTalkSendTemplate{Target: application.DingTalkTemplateTarget{Kind: "workflow_start", DefinitionID: 17}}, UserIDs: []uint{1, 2}})
	if !errors.Is(err, application.ErrTemplateTargetUnavailable) || len(client.calls) != 0 {
		t.Fatalf("error=%v calls=%d", err, len(client.calls))
	}
}

type rejectTemplateStyleLoader struct{ t *testing.T }

func (s rejectTemplateStyleLoader) NotificationStyles(context.Context) (notificationstyle.Config, error) {
	s.t.Error("template loaded global styles")
	return notificationstyle.Config{}, errors.New("must not load")
}

type templateMutatingClient struct {
	dingTalkClientStub
	change func()
}

func (c *templateMutatingClient) SendWorkNotificationContext(ctx context.Context, config configsvc.DingTalkH5CorpConfig, ids []string, payload configsvc.DingTalkWorkNotificationPayload) error {
	err := c.dingTalkClientStub.SendWorkNotificationContext(ctx, config, ids, payload)
	c.change()
	return err
}

func TestTemplateHomePreservesEntryAndRequiresAppIdentity(t *testing.T) {
	config := configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AgentID: "1", AppURL: "https://example.test/app?view=dashboard&key=entry&page=custom#/welcome"}
	template := application.DingTalkSendTemplate{MessageType: "oa", Target: application.DingTalkTemplateTarget{Kind: "app_home"}}
	payload, err := dingTalkSendTemplatePayload(config, template, "title", "content")
	if err != nil {
		t.Fatal(err)
	}
	wrapped, _ := url.Parse(payload.URL)
	if wrapped.Query().Get("redirect_url") != config.AppURL {
		t.Fatalf("home=%s", payload.URL)
	}
	config.AgentID = ""
	if _, err := dingTalkSendTemplatePayload(config, template, "title", "content"); !errors.Is(err, application.ErrTemplateTargetUnavailable) {
		t.Fatalf("missing app=%v", err)
	}
	config.AgentID = "1"
	config.CorpID = ""
	if _, err := dingTalkSendTemplatePayload(config, template, "title", "content"); !errors.Is(err, application.ErrTemplateTargetUnavailable) {
		t.Fatalf("missing corp=%v", err)
	}
}

func TestTemplateHomeKeepsExistingHTTPAppConfiguration(t *testing.T) {
	config := configsvc.DingTalkH5CorpConfig{CorpID: "corp-a", AgentID: "1", AppURL: "http://example.test/app?mode=development#/entry"}
	payload, err := dingTalkSendTemplatePayload(config, application.DingTalkSendTemplate{Target: application.DingTalkTemplateTarget{Kind: "app_home"}}, "title", "content")
	if err != nil {
		t.Fatal(err)
	}
	wrapped, _ := url.Parse(payload.URL)
	if wrapped.Query().Get("redirect_url") != config.AppURL {
		t.Fatalf("home=%s", payload.URL)
	}
}
