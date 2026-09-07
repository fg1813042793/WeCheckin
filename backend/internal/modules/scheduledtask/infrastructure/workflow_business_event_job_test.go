package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	scheduledtaskapp "wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestWorkflowBusinessEventJobUsesConfiguredLimit(t *testing.T) {
	dispatcher := &fakeWorkflowBusinessEventDispatcher{count: 3}
	job := NewWorkflowBusinessEventJob(dispatcher)

	result, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"limit":25}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if job.Key() != "workflow.business-event.dispatch_due" || dispatcher.limit != 25 || result.Data["dispatched"] != 3 {
		t.Fatalf("job result=%#v dispatcher=%#v", result, dispatcher)
	}
}

func TestWorkflowBusinessEventJobReturnsTemporaryFailure(t *testing.T) {
	dispatcher := &fakeWorkflowBusinessEventDispatcher{err: errors.New("database unavailable")}
	job := NewWorkflowBusinessEventJob(dispatcher)

	_, err := job.Execute(context.Background(), "run-1", json.RawMessage(`{"limit":100}`), nil)
	var handlerErr *scheduledtaskapp.HandlerError
	if !errors.As(err, &handlerErr) || !handlerErr.Temporary || handlerErr.Code != "workflow_business_event_dispatch_failed" {
		t.Fatalf("err=%#v", err)
	}
}

type fakeWorkflowBusinessEventDispatcher struct {
	limit int
	count int
	err   error
}

func (dispatcher *fakeWorkflowBusinessEventDispatcher) DispatchDueBusinessEvents(_ context.Context, limit int) (int, error) {
	dispatcher.limit = limit
	return dispatcher.count, dispatcher.err
}
