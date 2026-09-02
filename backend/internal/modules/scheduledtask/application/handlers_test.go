package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

func TestHandlerRegistryRejectsDuplicatesAndUnknownTypes(t *testing.T) {
	registry := NewHandlerRegistry()
	handler := &fakeHandler{handlerType: "go"}
	if err := registry.Register(handler); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(registry.Register(handler), ErrHandlerAlreadyRegistered) {
		t.Fatal("duplicate handler type must fail")
	}
	if _, err := registry.Handler("missing"); !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("unknown handler error = %v", err)
	}
}

func TestHandlerRegistryValidatesAndExecutesTaskSnapshot(t *testing.T) {
	handler := &fakeHandler{handlerType: "go", result: HandlerResult{Summary: "ok"}}
	registry := NewHandlerRegistry()
	if err := registry.Register(handler); err != nil {
		t.Fatal(err)
	}
	if err := registry.ValidateConfig(context.Background(), "go", json.RawMessage(`{"handlerKey":"cleanup"}`)); err != nil {
		t.Fatal(err)
	}
	task := scheduledtaskmodel.Task{HandlerType: "go", HandlerConfigJSON: `{"handlerKey":"cleanup"}`}
	snapshot, _ := json.Marshal(task)
	run := &scheduledtaskmodel.Run{ID: "run-1", TriggerType: "manual", TaskSnapshotJSON: string(snapshot)}
	result, err := registry.Execute(context.Background(), run)
	if err != nil || result.Summary != "ok" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
	if handler.context.RunID != "run-1" || handler.context.Task.HandlerType != "go" {
		t.Fatalf("run context = %#v", handler.context)
	}
}

func TestHandlerRegistryPassesBoundRunLogger(t *testing.T) {
	handler := &fakeHandler{handlerType: "go"}
	registry := NewHandlerRegistry()
	if err := registry.Register(handler); err != nil {
		t.Fatal(err)
	}
	task := scheduledtaskmodel.Task{HandlerType: "go", HandlerConfigJSON: `{}`}
	snapshot, _ := json.Marshal(task)
	logger := NopRunLogger{}
	if _, err := registry.ExecuteWithLogger(context.Background(), &scheduledtaskmodel.Run{ID: "run-logger", TaskSnapshotJSON: string(snapshot)}, logger); err != nil {
		t.Fatal(err)
	}
	if handler.context.Logger == nil {
		t.Fatal("bound run logger was not passed to handler")
	}
}

type fakeHandler struct {
	handlerType string
	result      HandlerResult
	context     RunContext
}

func (handler *fakeHandler) Type() string { return handler.handlerType }
func (handler *fakeHandler) Metadata() HandlerMetadata {
	return HandlerMetadata{Type: handler.handlerType, Name: "Fake"}
}
func (handler *fakeHandler) ValidateConfig(context.Context, json.RawMessage) error { return nil }
func (handler *fakeHandler) Execute(_ context.Context, run RunContext) (HandlerResult, error) {
	handler.context = run
	return handler.result, nil
}
