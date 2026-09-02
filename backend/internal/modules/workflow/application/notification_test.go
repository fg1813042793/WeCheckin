package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecheckin/backend/internal/workflowcore"
)

type notificationRepositoryStub struct {
	claimed          []NotificationRecord
	due              []NotificationRecord
	resetID          string
	inAppDelivered   []string
	sent             []string
	failed           []notificationFailure
	claimNow         int64
	claimStaleBefore int64
	dueLimit         int
	listed           NotificationQuery
	inAppErr         error
}

type notificationFailure struct {
	id          string
	attempts    int
	status      string
	nextRetryAt int64
	message     string
}

func (stub *notificationRepositoryStub) List(_ context.Context, query NotificationQuery) (*NotificationList, error) {
	stub.listed = query
	return &NotificationList{}, nil
}

func (stub *notificationRepositoryStub) ClaimByIDs(_ context.Context, _ []string, now, staleBefore int64) ([]NotificationRecord, error) {
	stub.claimNow = now
	stub.claimStaleBefore = staleBefore
	return append([]NotificationRecord(nil), stub.claimed...), nil
}

func (stub *notificationRepositoryStub) ClaimDue(_ context.Context, now, staleBefore int64, limit int) ([]NotificationRecord, error) {
	stub.claimNow = now
	stub.claimStaleBefore = staleBefore
	stub.dueLimit = limit
	return append([]NotificationRecord(nil), stub.due...), nil
}

func (stub *notificationRepositoryStub) DeliverInApp(_ context.Context, notification NotificationRecord, _ int64) error {
	stub.inAppDelivered = append(stub.inAppDelivered, notification.ID)
	return stub.inAppErr
}

func (stub *notificationRepositoryStub) MarkSent(_ context.Context, id, _ string, _ int64) error {
	stub.sent = append(stub.sent, id)
	return nil
}

func (stub *notificationRepositoryStub) MarkFailed(_ context.Context, id string, attempts int, status string, nextRetryAt int64, message string, _ int64) error {
	stub.failed = append(stub.failed, notificationFailure{id: id, attempts: attempts, status: status, nextRetryAt: nextRetryAt, message: message})
	return nil
}

func (stub *notificationRepositoryStub) ResetForRetry(_ context.Context, id string, _ int64) error {
	stub.resetID = id
	return nil
}

type notificationChannelStub struct {
	name    string
	errByID map[string]error
	seen    []string
}

func (stub *notificationChannelStub) Name() string { return stub.name }

func (stub *notificationChannelStub) Deliver(_ context.Context, notifications []NotificationRecord) []NotificationDeliveryResult {
	results := make([]NotificationDeliveryResult, 0, len(notifications))
	for _, notification := range notifications {
		stub.seen = append(stub.seen, notification.ID)
		results = append(results, NotificationDeliveryResult{ID: notification.ID, Err: stub.errByID[notification.ID]})
	}
	return results
}

func TestNotificationDispatcherKeepsChannelResultsIndependent(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	repository := &notificationRepositoryStub{claimed: []NotificationRecord{
		{ID: "in-app-1", Channel: workflowcore.NotificationChannelInApp},
		{ID: "ding-1", Channel: workflowcore.NotificationChannelDingTalkOA, Attempts: 0},
	}}
	dingTalk := &notificationChannelStub{
		name:    workflowcore.NotificationChannelDingTalkOA,
		errByID: map[string]error{"ding-1": errors.New("dingtalk unavailable")},
	}
	dispatcher := newNotificationDispatcherWithClock(repository, func() time.Time { return now }, dingTalk)

	count, err := dispatcher.Dispatch(context.Background(), []string{"in-app-1", "ding-1"})
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if count != 2 || len(repository.inAppDelivered) != 1 || repository.inAppDelivered[0] != "in-app-1" {
		t.Fatalf("in-app delivery = count %d ids %#v", count, repository.inAppDelivered)
	}
	if len(repository.sent) != 0 {
		t.Fatalf("in-app channel finalizes atomically and must not be marked twice: %#v", repository.sent)
	}
	if len(repository.failed) != 1 {
		t.Fatalf("failed deliveries = %#v", repository.failed)
	}
	failure := repository.failed[0]
	if failure.id != "ding-1" || failure.attempts != 1 || failure.status != NotificationStatusFailed {
		t.Fatalf("failure state = %#v", failure)
	}
	if failure.nextRetryAt != now.Add(time.Minute).UnixMilli() {
		t.Fatalf("first retry at = %d", failure.nextRetryAt)
	}
	if repository.claimStaleBefore != now.Add(-10*time.Minute).UnixMilli() {
		t.Fatalf("stale sending threshold = %d", repository.claimStaleBefore)
	}
}

func TestNotificationDispatcherMovesFifthFailureToDead(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	repository := &notificationRepositoryStub{due: []NotificationRecord{{
		ID: "ding-dead", Channel: workflowcore.NotificationChannelDingTalkOA, Attempts: 4,
	}}}
	dingTalk := &notificationChannelStub{name: workflowcore.NotificationChannelDingTalkOA, errByID: map[string]error{"ding-dead": errors.New("still unavailable")}}
	dispatcher := newNotificationDispatcherWithClock(repository, func() time.Time { return now }, dingTalk)

	count, err := dispatcher.DispatchDue(context.Background(), 0)
	if err != nil {
		t.Fatalf("DispatchDue() error = %v", err)
	}
	if count != 1 || repository.dueLimit != 100 {
		t.Fatalf("due dispatch = count %d limit %d", count, repository.dueLimit)
	}
	failure := repository.failed[0]
	if failure.attempts != 5 || failure.status != NotificationStatusDead || failure.nextRetryAt != 0 {
		t.Fatalf("dead failure = %#v", failure)
	}
}

func TestNotificationDispatcherSchedulesInAppFailureForRetry(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	repository := &notificationRepositoryStub{
		claimed:  []NotificationRecord{{ID: "in-app-failed", Channel: workflowcore.NotificationChannelInApp}},
		inAppErr: errors.New("notify table unavailable"),
	}
	dispatcher := newNotificationDispatcherWithClock(repository, func() time.Time { return now })

	count, err := dispatcher.Dispatch(context.Background(), []string{"in-app-failed"})
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if count != 1 || len(repository.failed) != 1 || repository.failed[0].id != "in-app-failed" || repository.failed[0].status != NotificationStatusFailed {
		t.Fatalf("in-app retry state = count %d failures %#v", count, repository.failed)
	}
}

func TestNotificationDispatcherRetriesSingleRecordImmediately(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	repository := &notificationRepositoryStub{claimed: []NotificationRecord{{ID: "ding-retry", Channel: workflowcore.NotificationChannelDingTalkOA}}}
	dingTalk := &notificationChannelStub{name: workflowcore.NotificationChannelDingTalkOA, errByID: map[string]error{}}
	dispatcher := newNotificationDispatcherWithClock(repository, func() time.Time { return now }, dingTalk)

	if err := dispatcher.Retry(context.Background(), "ding-retry"); err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
	if repository.resetID != "ding-retry" || len(repository.sent) != 1 || repository.sent[0] != "ding-retry" {
		t.Fatalf("retry state = reset %q sent %#v", repository.resetID, repository.sent)
	}
}

func TestNotificationRetryBackoffSchedule(t *testing.T) {
	want := []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 12 * time.Hour}
	for attempts, duration := range want {
		if got := notificationRetryDelay(attempts + 1); got != duration {
			t.Fatalf("attempt %d delay = %s, want %s", attempts+1, got, duration)
		}
	}
}
