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

func TestGoHandlerMetadataIncludesRegisteredTaskNames(t *testing.T) {
	handler := NewGoHandler()
	if err := handler.Register(&fakeGoJob{key: "scheduled-task.cleanup", name: "清理定时任务历史"}); err != nil {
		t.Fatal(err)
	}

	var schema struct {
		Properties map[string]struct {
			Enum       []string          `json:"enum"`
			EnumLabels map[string]string `json:"x-enum-labels"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(handler.Metadata().ConfigSchema, &schema); err != nil {
		t.Fatal(err)
	}
	handlerKey := schema.Properties["handlerKey"]
	if len(handlerKey.Enum) != 1 || handlerKey.Enum[0] != "scheduled-task.cleanup" {
		t.Fatalf("handlerKey enum = %#v", handlerKey.Enum)
	}
	if got := handlerKey.EnumLabels["scheduled-task.cleanup"]; got != "清理定时任务历史" {
		t.Fatalf("handlerKey label = %q", got)
	}
}

type fakeGoJob struct {
	key    string
	name   string
	runID  string
	params json.RawMessage
}

func (job *fakeGoJob) Key() string                                     { return job.key }
func (job *fakeGoJob) Name() string                                    { return job.name }
func (job *fakeGoJob) ConfigSchema() json.RawMessage                   { return json.RawMessage(`{"type":"object"}`) }
func (job *fakeGoJob) Validate(context.Context, json.RawMessage) error { return nil }
func (job *fakeGoJob) Execute(_ context.Context, runID string, params json.RawMessage, _ application.RunLogger) (application.HandlerResult, error) {
	job.runID = runID
	job.params = append(json.RawMessage(nil), params...)
	return application.HandlerResult{Summary: "cleaned"}, nil
}
