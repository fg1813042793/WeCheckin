package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestGoHandlerUsesOnlyRegisteredTaskKeys(t *testing.T) {
	handler := NewGoHandler()
	job := &fakeGoJob{key: "scheduled-task.cleanup"}
	if err := handler.Register(job); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(handler.Register(job), ErrGoJobAlreadyRegistered) {
		t.Fatal("duplicate Go job must fail")
	}
	if err := handler.ValidateConfig(context.Background(), json.RawMessage(`{"handlerKey":"missing"}`)); !errors.Is(err, ErrGoJobNotFound) {
		t.Fatalf("unknown Go job error = %v", err)
	}
	result, err := handler.Execute(context.Background(), application.RunContext{
		RunID: "run-1", Task: application.TaskSnapshot{HandlerConfigJSON: `{"handlerKey":"scheduled-task.cleanup","params":{"days":30}}`},
	})
	if err != nil || result.Summary != "cleaned" || job.runID != "run-1" || string(job.params) != `{"days":30}` {
		t.Fatalf("result/job = %#v / %#v, err = %v", result, job, err)
	}
}

type fakeGoJob struct {
	key    string
	runID  string
	params json.RawMessage
}

func (job *fakeGoJob) Key() string                                     { return job.key }
func (job *fakeGoJob) Name() string                                    { return "Cleanup" }
func (job *fakeGoJob) ConfigSchema() json.RawMessage                   { return json.RawMessage(`{"type":"object"}`) }
func (job *fakeGoJob) Validate(context.Context, json.RawMessage) error { return nil }
func (job *fakeGoJob) Execute(_ context.Context, runID string, params json.RawMessage, _ application.RunLogger) (application.HandlerResult, error) {
	job.runID = runID
	job.params = append(json.RawMessage(nil), params...)
	return application.HandlerResult{Summary: "cleaned"}, nil
}
