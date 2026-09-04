package infrastructure

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"wecheckin/backend/internal/modules/inappnotification/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
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

type dingTalkTargetResolverStub struct {
	resolution dingTalkTargetResolution
	err        error
}

func (stub *dingTalkTargetResolverStub) ResolveTargets(context.Context, []uint) (dingTalkTargetResolution, error) {
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
