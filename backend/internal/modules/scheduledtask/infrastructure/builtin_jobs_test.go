package infrastructure

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestWorkflowNotificationDispatchJobUsesRegisteredService(t *testing.T) {
	dispatcher := &fakeNotificationDispatcher{count: 3}
	job := NewWorkflowNotificationDispatchJob(dispatcher)
	result, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"limit":25}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "workflow.notification.dispatch_due" || dispatcher.limit != 25 || result.Data["dispatched"] != 3 {
		t.Fatalf("job result = %#v, dispatcher = %#v", result, dispatcher)
	}
}

func TestCleanupJobUsesServerRetentionPolicy(t *testing.T) {
	store := &fakeCleanupStore{}
	now := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	job := NewCleanupJob(store, 90, 30, func() time.Time { return now })
	_, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"batchSize":500}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if store.limit != 500 || store.runBefore != now.AddDate(0, 0, -90).UnixMilli() || store.logBefore != now.AddDate(0, 0, -30).UnixMilli() {
		t.Fatalf("cleanup args = %#v", store)
	}
}

type fakeNotificationDispatcher struct {
	limit int
	count int
}

func (dispatcher *fakeNotificationDispatcher) DispatchDueNotifications(_ context.Context, limit int) (int, error) {
	dispatcher.limit = limit
	return dispatcher.count, nil
}

type fakeCleanupStore struct {
	runBefore int64
	logBefore int64
	limit     int
}

func (store *fakeCleanupStore) Cleanup(_ context.Context, runBefore, logBefore int64, limit int) (CleanupResult, error) {
	store.runBefore = runBefore
	store.logBefore = logBefore
	store.limit = limit
	return CleanupResult{RunsDeleted: 2, LogsDeleted: 4}, nil
}
