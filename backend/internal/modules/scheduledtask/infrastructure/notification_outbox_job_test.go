package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestNotificationOutboxDispatchJobUsesConfiguredLimit(t *testing.T) {
	dispatcher := &outboxDispatcherStub{count: 4}
	job := NewNotificationOutboxDispatchJob(dispatcher)
	result, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"limit":25}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "notification.outbox.dispatch_due" || dispatcher.limit != 25 || result.Data["dispatched"] != 4 {
		t.Fatalf("job=%q dispatcher=%#v result=%#v", job.Key(), dispatcher, result)
	}
}

func TestNotificationOutboxDispatchJobReturnsTemporaryFailure(t *testing.T) {
	dispatcher := &outboxDispatcherStub{err: errors.New("database unavailable")}
	job := NewNotificationOutboxDispatchJob(dispatcher)
	_, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{}`), nil)
	var handlerError *application.HandlerError
	if !errors.As(err, &handlerError) || !handlerError.Temporary || handlerError.Code != "notification_outbox_dispatch_failed" {
		t.Fatalf("error = %#v", err)
	}
}

type outboxDispatcherStub struct {
	limit int
	count int
	err   error
}

func (dispatcher *outboxDispatcherStub) DispatchDue(_ context.Context, limit int) (int, error) {
	dispatcher.limit = limit
	return dispatcher.count, dispatcher.err
}
