package application

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationmodel "wecheckin/backend/internal/model/notification"
)

func TestServiceEnqueuePersistsPendingNotification(t *testing.T) {
	store := &storeStub{}
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	service := newService(store, func() time.Time { return now })

	err := service.Enqueue(context.Background(), EnqueueRequest{
		IdempotencyKey: "survey:12:answer:34:rule:r1:webhook",
		SourceType:     "survey_response", SourceID: "34", Channel: "webhook",
		Recipient: map[string]any{"url": "https://example.com/hook"},
		Payload:   map[string]any{"title": "问卷", "content": "已提交"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.enqueued.IdempotencyKey == "" || store.enqueued.Status != notificationmodel.StatusPending {
		t.Fatalf("enqueued = %#v", store.enqueued)
	}
	if store.enqueued.NextRetryAt != now.UnixMilli() || store.enqueued.Attempts != 0 {
		t.Fatalf("initial delivery state = %#v", store.enqueued)
	}
	if store.enqueued.RecipientJSON == "" || store.enqueued.PayloadJSON == "" {
		t.Fatalf("serialized fields = %#v", store.enqueued)
	}
}

func TestServiceDispatchDueMarksSuccessfulDeliverySent(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	store := &storeStub{due: []notificationmodel.Outbox{{ID: 7, Channel: "internal", Status: notificationmodel.StatusSending}}}
	channel := &channelStub{name: "internal"}
	service := newService(store, func() time.Time { return now }, channel)

	count, err := service.DispatchDue(context.Background(), 25)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || store.sentID != 7 || channel.delivered != 1 {
		t.Fatalf("count=%d sent=%d delivered=%d", count, store.sentID, channel.delivered)
	}
}

func TestServiceDispatchDueSchedulesFailureForRetry(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	store := &storeStub{due: []notificationmodel.Outbox{{ID: 8, Channel: "webhook", Status: notificationmodel.StatusSending}}}
	service := newService(store, func() time.Time { return now }, &channelStub{name: "webhook", err: errors.New("temporary failure")})

	count, err := service.DispatchDue(context.Background(), 25)
	if count != 1 || err == nil {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if store.failed.ID != 8 || store.failed.Attempts != 1 || store.failed.Status != notificationmodel.StatusFailed {
		t.Fatalf("failed = %#v", store.failed)
	}
	if store.failed.NextRetryAt != now.Add(time.Minute).UnixMilli() {
		t.Fatalf("next retry = %d", store.failed.NextRetryAt)
	}
}

func TestServiceDispatchDueMovesFifthFailureToDead(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	store := &storeStub{due: []notificationmodel.Outbox{{ID: 9, Channel: "missing", Status: notificationmodel.StatusSending, Attempts: 4}}}
	service := newService(store, func() time.Time { return now })

	_, err := service.DispatchDue(context.Background(), 25)
	if err == nil {
		t.Fatal("missing channel must fail")
	}
	if store.failed.Attempts != 5 || store.failed.Status != notificationmodel.StatusDead || store.failed.NextRetryAt != 0 {
		t.Fatalf("failed = %#v", store.failed)
	}
}

type storeStub struct {
	enqueued notificationmodel.Outbox
	due      []notificationmodel.Outbox
	sentID   uint64
	failed   Failure
}

func (store *storeStub) Enqueue(_ context.Context, row notificationmodel.Outbox) (bool, error) {
	store.enqueued = row
	return true, nil
}

func (store *storeStub) ClaimDue(_ context.Context, _, _ int64, _ int) ([]notificationmodel.Outbox, error) {
	return store.due, nil
}

func (store *storeStub) MarkSent(_ context.Context, id uint64, _ int64) error {
	store.sentID = id
	return nil
}

func (store *storeStub) MarkFailed(_ context.Context, failure Failure) error {
	store.failed = failure
	return nil
}

type channelStub struct {
	name      string
	err       error
	delivered int
}

func (channel *channelStub) Name() string { return channel.name }

func (channel *channelStub) Deliver(context.Context, notificationmodel.Outbox) error {
	channel.delivered++
	return channel.err
}
