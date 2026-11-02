package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	notificationmodel "wecheckin/backend/internal/model/notification"
	inappnotificationapp "wecheckin/backend/internal/modules/inappnotification/application"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
)

func TestDingTalkChannelDeliversFeedbackToOneUser(t *testing.T) {
	delivery := &dingTalkDeliveryStub{result: inappnotificationapp.DingTalkDeliveryResult{SentCount: 1}}
	channel := NewDingTalkChannel(delivery)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := channel.Deliver(ctx, feedbackDingTalkRow(t))
	if err != nil || channel.Name() != notificationmodel.ChannelDingTalk {
		t.Fatalf("name=%q err=%v", channel.Name(), err)
	}
	want := inappnotificationapp.DingTalkDeliveryBatch{
		Title: "反馈新消息", Content: "问题已修复", NotificationType: "feedback_status", UserIDs: []uint{7},
	}
	if delivery.calls != 1 || delivery.ctx != ctx || !reflect.DeepEqual(delivery.batch, want) {
		t.Fatalf("calls=%d batch=%#v", delivery.calls, delivery.batch)
	}
	assertDingTalkCounts(t, result, 1, 1, 0, 0)
}

func TestDingTalkChannelRejectsInvalidRowsWithoutDelivery(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*notificationmodel.Outbox)
	}{
		{"malformed recipient", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{` }},
		{"null recipient", func(row *notificationmodel.Outbox) { row.RecipientJSON = `null` }},
		{"missing users", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{}` }},
		{"empty users", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[]}` }},
		{"zero user", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[0]}` }},
		{"multiple users", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[7,9]}` }},
		{"duplicate users", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[7,7]}` }},
		{"broadcast", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"notifyAdmin":true,"userIds":[7]}` }},
		{"negative user", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[-7]}` }},
		{"string user", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":["7"]}` }},
		{"legacy record", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"plannedCount":1,"sentCount":0}` }},
		{"unknown recipient fields", func(row *notificationmodel.Outbox) { row.RecipientJSON = `{"userIds":[7],"corpId":"other"}` }},
		{"trailing recipient", func(row *notificationmodel.Outbox) { row.RecipientJSON += `{}` }},
		{"wrong channel", func(row *notificationmodel.Outbox) { row.Channel = notificationmodel.ChannelInternal }},
		{"wrong source", func(row *notificationmodel.Outbox) { row.SourceType = "scheduled_task_run" }},
		{"missing source", func(row *notificationmodel.Outbox) { row.SourceType = "" }},
		{"missing source id", func(row *notificationmodel.Outbox) { row.SourceID = "" }},
		{"source id mismatch", func(row *notificationmodel.Outbox) { row.SourceID = "92" }},
		{"malformed payload", func(row *notificationmodel.Outbox) { row.PayloadJSON = `{` }},
		{"null payload", func(row *notificationmodel.Outbox) { row.PayloadJSON = `null` }},
	} {
		t.Run(test.name, func(t *testing.T) {
			row := feedbackDingTalkRow(t)
			test.mutate(&row)
			delivery := &dingTalkDeliveryStub{}
			if _, err := NewDingTalkChannel(delivery).Deliver(context.Background(), row); err == nil {
				t.Fatal("invalid row accepted")
			}
			if delivery.calls != 0 {
				t.Fatalf("sent invalid row: %d calls", delivery.calls)
			}
		})
	}
	for _, test := range []struct {
		name   string
		mutate func(*notificationoutboxapp.MessagePayload)
	}{
		{"wrong payload source", func(p *notificationoutboxapp.MessagePayload) { p.SourceType = "survey" }},
		{"missing payload source", func(p *notificationoutboxapp.MessagePayload) { p.SourceType = "" }},
		{"missing payload id", func(p *notificationoutboxapp.MessagePayload) { p.SourceID = "" }},
		{"wrong notification type", func(p *notificationoutboxapp.MessagePayload) { p.NotificationType = "admin_manual" }},
		{"missing notification type", func(p *notificationoutboxapp.MessagePayload) { p.NotificationType = "" }},
		{"blank title", func(p *notificationoutboxapp.MessagePayload) { p.Title = "  " }},
		{"blank content", func(p *notificationoutboxapp.MessagePayload) { p.Content = "\n " }},
	} {
		t.Run(test.name, func(t *testing.T) {
			row := feedbackDingTalkRow(t)
			var payload notificationoutboxapp.MessagePayload
			if err := json.Unmarshal([]byte(row.PayloadJSON), &payload); err != nil {
				t.Fatal(err)
			}
			test.mutate(&payload)
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			row.PayloadJSON = string(body)
			delivery := &dingTalkDeliveryStub{}
			if _, err := NewDingTalkChannel(delivery).Deliver(context.Background(), row); err == nil || delivery.calls != 0 {
				t.Fatalf("err=%v calls=%d", err, delivery.calls)
			}
		})
	}
}

func TestDingTalkChannelRejectsUnavailableDelivery(t *testing.T) {
	for _, channel := range []*DingTalkChannel{nil, NewDingTalkChannel(nil)} {
		if _, err := channel.Deliver(context.Background(), feedbackDingTalkRow(t)); !errors.Is(err, inappnotificationapp.ErrDingTalkDeliveryUnavailable) {
			t.Fatalf("error=%v", err)
		}
	}
}

func TestDingTalkChannelTreatsUnsuccessfulCountsAsRetryableFailure(t *testing.T) {
	downstreamErr := errors.New("delivery unavailable")
	for _, test := range []struct {
		name   string
		result inappnotificationapp.DingTalkDeliveryResult
		err    error
	}{
		{"failed", inappnotificationapp.DingTalkDeliveryResult{FailedCount: 1}, nil},
		{"unbound", inappnotificationapp.DingTalkDeliveryResult{SkippedCount: 1}, nil},
		{"zero sent", inappnotificationapp.DingTalkDeliveryResult{}, nil},
		{"partial failed", inappnotificationapp.DingTalkDeliveryResult{SentCount: 1, FailedCount: 1}, nil},
		{"partial skipped", inappnotificationapp.DingTalkDeliveryResult{SentCount: 1, SkippedCount: 1}, nil},
		{"too many sent", inappnotificationapp.DingTalkDeliveryResult{SentCount: 2}, nil},
		{"invalid counts", inappnotificationapp.DingTalkDeliveryResult{SentCount: 1, FailedCount: -1}, nil},
		{"downstream error", inappnotificationapp.DingTalkDeliveryResult{}, downstreamErr},
		{"error after send", inappnotificationapp.DingTalkDeliveryResult{SentCount: 1}, downstreamErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			delivery := &dingTalkDeliveryStub{result: test.result, err: test.err}
			result, err := NewDingTalkChannel(delivery).Deliver(context.Background(), feedbackDingTalkRow(t))
			if err == nil || (test.err != nil && !errors.Is(err, test.err)) {
				t.Fatalf("error=%v", err)
			}
			assertDingTalkCounts(t, result, 1, test.result.SentCount, test.result.SkippedCount, test.result.FailedCount)
		})
	}
}

func TestDingTalkChannelUsesOutboxRetryAndSingleRecord(t *testing.T) {
	store := &dingTalkOutboxStoreStub{row: feedbackDingTalkRow(t)}
	delivery := &dingTalkDeliveryStub{result: inappnotificationapp.DingTalkDeliveryResult{FailedCount: 1}}
	service := notificationoutboxapp.NewService(store, NewDingTalkChannel(delivery))
	if count, err := service.DispatchDue(context.Background(), 1); count != 1 || err == nil {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if store.sent != 0 || len(store.failures) != 1 {
		t.Fatalf("sent=%d failures=%v", store.sent, store.failures)
	}
	failure := store.failures[0]
	if failure.ID != store.row.ID || failure.Status != notificationmodel.StatusFailed || failure.Attempts != 1 || failure.NextRetryAt <= failure.EditTime {
		t.Fatalf("failure=%#v", failure)
	}
	delivery.result = inappnotificationapp.DingTalkDeliveryResult{SentCount: 1}
	if count, err := service.DispatchDue(context.Background(), 1); count != 1 || err != nil {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if delivery.calls != 2 || store.sent != 1 || store.enqueued != 0 {
		t.Fatalf("calls=%d sent=%d extra records=%d", delivery.calls, store.sent, store.enqueued)
	}
	assertDingTalkCounts(t, store.result, 1, 1, 0, 0)
}

func feedbackDingTalkRow(t *testing.T) notificationmodel.Outbox {
	t.Helper()
	row := notificationRow(t, notificationoutboxapp.InternalRecipient{UserIDs: []uint{7}}, notificationoutboxapp.MessagePayload{
		Title: "反馈新消息", Content: "问题已修复", SourceType: "user_feedback", SourceID: "91", NotificationType: "feedback_status",
	})
	row.Channel, row.SourceType, row.SourceID = notificationmodel.ChannelDingTalk, "user_feedback", "91"
	return row
}

func assertDingTalkCounts(t *testing.T, result notificationoutboxapp.DeliveryResult, planned, sent, skipped, failed int) {
	t.Helper()
	var counts map[string]int
	if err := json.Unmarshal([]byte(result.ResponseBody), &counts); err != nil {
		t.Fatalf("response counts: %v", err)
	}
	want := map[string]int{"plannedCount": planned, "sentCount": sent, "skippedCount": skipped, "failedCount": failed}
	if !reflect.DeepEqual(counts, want) {
		t.Fatalf("counts=%v want=%v", counts, want)
	}
}

type dingTalkDeliveryStub struct {
	ctx    context.Context
	batch  inappnotificationapp.DingTalkDeliveryBatch
	calls  int
	result inappnotificationapp.DingTalkDeliveryResult
	err    error
}

func (stub *dingTalkDeliveryStub) DeliverDingTalk(ctx context.Context, batch inappnotificationapp.DingTalkDeliveryBatch) (inappnotificationapp.DingTalkDeliveryResult, error) {
	stub.ctx, stub.batch = ctx, batch
	stub.calls++
	return stub.result, stub.err
}

type dingTalkOutboxStoreStub struct {
	row            notificationmodel.Outbox
	failures       []notificationoutboxapp.Failure
	sent, enqueued int
	result         notificationoutboxapp.DeliveryResult
}

func (store *dingTalkOutboxStoreStub) Enqueue(context.Context, notificationmodel.Outbox) (bool, error) {
	store.enqueued++
	return true, nil
}
func (store *dingTalkOutboxStoreStub) ClaimDue(context.Context, int64, int64, int) ([]notificationmodel.Outbox, error) {
	return []notificationmodel.Outbox{store.row}, nil
}
func (store *dingTalkOutboxStoreStub) MarkSent(_ context.Context, _ uint64, _ int64, result notificationoutboxapp.DeliveryResult) error {
	store.sent++
	store.result = result
	return nil
}
func (store *dingTalkOutboxStoreStub) MarkFailed(_ context.Context, failure notificationoutboxapp.Failure) error {
	store.failures = append(store.failures, failure)
	store.row.Attempts = failure.Attempts
	return nil
}
