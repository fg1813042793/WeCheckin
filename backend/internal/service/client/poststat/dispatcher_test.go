package poststat

import (
	"context"
	"errors"
	"testing"

	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
)

func TestDispatchRuleNotificationsSupportsWebhookInternalAndBoth(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		want    int
	}{
		{name: "webhook", channel: "webhook", want: 1},
		{name: "internal", channel: "internal", want: 1},
		{name: "both", channel: "both", want: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dispatcher := &notificationDispatcherStub{}
			err := dispatchRuleNotifications(context.Background(), dispatcher, ruleNotificationInput{
				SurveyID: 12, ResponseID: 34, RuleIndex: 2, SurveyTitle: "问卷", Message: "内容",
				Rule: Rule{ID: "r1", NotifyChannel: test.channel, WebhookType: "dingtalk", WebhookURL: "https://example.com/hook", NotifyUserIds: "7,9"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(dispatcher.requests) != test.want {
				t.Fatalf("requests = %#v", dispatcher.requests)
			}
			keys := make(map[string]struct{}, len(dispatcher.requests))
			for _, request := range dispatcher.requests {
				if request.IdempotencyKey == "" {
					t.Fatal("idempotency key must not be empty")
				}
				keys[request.IdempotencyKey] = struct{}{}
			}
			if len(keys) != test.want {
				t.Fatalf("idempotency keys are not channel-specific: %#v", dispatcher.requests)
			}
		})
	}
}

func TestDispatchRuleNotificationsSkipsUnconfiguredDestinations(t *testing.T) {
	dispatcher := &notificationDispatcherStub{}
	err := dispatchRuleNotifications(context.Background(), dispatcher, ruleNotificationInput{
		SurveyID: 12, ResponseID: 34, Rule: Rule{NotifyChannel: "both"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dispatcher.requests) != 0 {
		t.Fatalf("requests = %#v", dispatcher.requests)
	}
}

func TestDispatchRuleNotificationsReturnsEnqueueFailure(t *testing.T) {
	want := errors.New("database unavailable")
	dispatcher := &notificationDispatcherStub{err: want}
	err := dispatchRuleNotifications(context.Background(), dispatcher, ruleNotificationInput{
		SurveyID: 12, ResponseID: 34, Rule: Rule{NotifyChannel: "webhook", WebhookURL: "https://example.com/hook"},
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func TestOutboxDispatcherMapsRequest(t *testing.T) {
	enqueuer := &outboxEnqueuerStub{}
	dispatcher := NewOutboxDispatcher(enqueuer)
	want := NotificationRequest{
		IdempotencyKey: "key", SourceType: "survey_response", SourceID: "34", Channel: "webhook",
		Recipient: map[string]string{"url": "https://example.com"}, Payload: map[string]string{"title": "问卷"},
	}
	if err := dispatcher.Enqueue(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	if enqueuer.request.IdempotencyKey != want.IdempotencyKey || enqueuer.request.Channel != want.Channel {
		t.Fatalf("request = %#v", enqueuer.request)
	}
}

type notificationDispatcherStub struct {
	requests []NotificationRequest
	err      error
}

func (dispatcher *notificationDispatcherStub) Enqueue(_ context.Context, request NotificationRequest) error {
	dispatcher.requests = append(dispatcher.requests, request)
	return dispatcher.err
}

type outboxEnqueuerStub struct {
	request notificationoutboxapp.EnqueueRequest
}

func (enqueuer *outboxEnqueuerStub) Enqueue(_ context.Context, request notificationoutboxapp.EnqueueRequest) error {
	enqueuer.request = request
	return nil
}
