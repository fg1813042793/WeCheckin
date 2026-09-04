package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	notificationmodel "wecheckin/backend/internal/model/notification"
	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/internal/support/outboundhttp"
)

func TestInternalChannelUsesAllScopeForNotifyAdmin(t *testing.T) {
	sender := &inAppSenderStub{}
	channel := NewInternalChannel(sender)
	row := notificationRow(t, notificationoutboxapp.InternalRecipient{NotifyAdmin: true, UserIDs: []uint{7}}, notificationoutboxapp.MessagePayload{
		Title: "统计", Content: "内容", SourceType: "survey", SourceID: "12",
	})

	if err := channel.Deliver(context.Background(), row); err != nil {
		t.Fatal(err)
	}
	if sender.input.Scope != inappnotificationapp.ScopeAll || sender.input.SourceType != "survey" || sender.input.SourceID != "12" || sender.input.DeliveryKey != "outbox-key" {
		t.Fatalf("send input = %#v", sender.input)
	}
}

func TestInternalChannelUsesSelectedUsers(t *testing.T) {
	sender := &inAppSenderStub{}
	channel := NewInternalChannel(sender)
	row := notificationRow(t, notificationoutboxapp.InternalRecipient{UserIDs: []uint{7, 9}}, notificationoutboxapp.MessagePayload{Title: "统计", Content: "内容"})
	row.SourceType, row.SourceID = "survey_response", "34"

	if err := channel.Deliver(context.Background(), row); err != nil {
		t.Fatal(err)
	}
	if sender.input.Scope != inappnotificationapp.ScopeUsers || len(sender.input.UserIDs) != 2 {
		t.Fatalf("send input = %#v", sender.input)
	}
}

func TestWebhookChannelKeepsDingTalkPayloadFormat(t *testing.T) {
	client := &outboundClientStub{response: outboundhttp.Response{StatusCode: 200}}
	channel := NewWebhookChannel(client)
	row := notificationRow(t, notificationoutboxapp.WebhookRecipient{Type: "dingtalk", URL: "https://example.com/hook"}, notificationoutboxapp.MessagePayload{
		Title: "问卷", Content: "第一行\n第二行",
	})

	if err := channel.Deliver(context.Background(), row); err != nil {
		t.Fatal(err)
	}
	if client.request.Method != "POST" || client.request.URL != "https://example.com/hook" || client.request.Timeout <= 0 {
		t.Fatalf("request = %#v", client.request)
	}
	var payload struct {
		MsgType  string `json:"msgtype"`
		Markdown struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		} `json:"markdown"`
	}
	if err := json.Unmarshal(client.request.Body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.MsgType != "markdown" || payload.Markdown.Title != "问卷" || payload.Markdown.Text != "第一行\n\n第二行" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestWebhookChannelRejectsNonSuccessStatus(t *testing.T) {
	client := &outboundClientStub{response: outboundhttp.Response{StatusCode: 500, Body: []byte("failed")}}
	channel := NewWebhookChannel(client)
	row := notificationRow(t, notificationoutboxapp.WebhookRecipient{URL: "https://example.com/hook"}, notificationoutboxapp.MessagePayload{Title: "问卷", Content: "内容"})

	err := channel.Deliver(context.Background(), row)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("error = %v", err)
	}
}

func notificationRow(t *testing.T, recipient any, payload any) notificationmodel.Outbox {
	t.Helper()
	recipientJSON, err := json.Marshal(recipient)
	if err != nil {
		t.Fatal(err)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return notificationmodel.Outbox{ID: 1, IdempotencyKey: "outbox-key", RecipientJSON: string(recipientJSON), PayloadJSON: string(payloadJSON)}
}

type inAppSenderStub struct {
	input inappnotificationapp.SendInput
}

func (sender *inAppSenderStub) Send(_ context.Context, input inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error) {
	sender.input = input
	return inappnotificationapp.SendResult{SentCount: 1}, nil
}

type outboundClientStub struct {
	request  outboundhttp.Request
	response outboundhttp.Response
	err      error
}

func (client *outboundClientStub) Do(_ context.Context, request outboundhttp.Request) (outboundhttp.Response, error) {
	client.request = request
	return client.response, client.err
}
