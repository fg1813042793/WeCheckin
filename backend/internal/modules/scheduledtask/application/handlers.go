package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

var (
	ErrHandlerNotFound          = errors.New("scheduled task handler not found")
	ErrHandlerAlreadyRegistered = errors.New("scheduled task handler already registered")
)

type TaskSnapshot = scheduledtaskmodel.Task

type QueueMessage struct {
	MessageID string
	RunID     string
}

type WorkerHeartbeat struct {
	WorkerID      string `json:"workerId"`
	Role          string `json:"role"`
	Version       string `json:"version"`
	StartedAt     int64  `json:"startedAt"`
	LastHeartbeat int64  `json:"lastHeartbeat"`
	CurrentRuns   int    `json:"currentRuns"`
	WorkerCount   int    `json:"workerCount"`
}

type HandlerMetadata struct {
	Type         string          `json:"type"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	RiskLevel    string          `json:"riskLevel"`
	ConfigSchema json.RawMessage `json:"configSchema"`
}

type HandlerResult struct {
	Summary string                 `json:"summary"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type HandlerError struct {
	Code      string
	Summary   string
	Temporary bool
}

func (err *HandlerError) Error() string {
	if err == nil {
		return ""
	}
	if err.Summary != "" {
		return err.Summary
	}
	return err.Code
}

type RunLogger interface {
	Log(context.Context, string, string, string) error
}

type RunContext struct {
	RunID       string
	TriggerType string
	Attempt     int
	Task        TaskSnapshot
	Logger      RunLogger
}

type Handler interface {
	Type() string
	Metadata() HandlerMetadata
	ValidateConfig(context.Context, json.RawMessage) error
	Execute(context.Context, RunContext) (HandlerResult, error)
}

type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{handlers: make(map[string]Handler)}
}

func (registry *HandlerRegistry) Register(handler Handler) error {
	if registry == nil || handler == nil {
		return errors.New("scheduled task handler is required")
	}
	handlerType := strings.TrimSpace(handler.Type())
	if handlerType == "" {
		return errors.New("scheduled task handler type is required")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.handlers[handlerType]; exists {
		return fmt.Errorf("%w: %s", ErrHandlerAlreadyRegistered, handlerType)
	}
	registry.handlers[handlerType] = handler
	return nil
}

func (registry *HandlerRegistry) Handler(handlerType string) (Handler, error) {
	if registry == nil {
		return nil, ErrHandlerNotFound
	}
	registry.mu.RLock()
	handler := registry.handlers[strings.TrimSpace(handlerType)]
	registry.mu.RUnlock()
	if handler == nil {
		return nil, fmt.Errorf("%w: %s", ErrHandlerNotFound, handlerType)
	}
	return handler, nil
}

func (registry *HandlerRegistry) Metadata() []HandlerMetadata {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	result := make([]HandlerMetadata, 0, len(registry.handlers))
	for _, handler := range registry.handlers {
		metadata := handler.Metadata()
		metadata.Type = handler.Type()
		result = append(result, metadata)
	}
	registry.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].Type < result[j].Type })
	return result
}

func (registry *HandlerRegistry) ValidateConfig(ctx context.Context, handlerType string, raw json.RawMessage) error {
	handler, err := registry.Handler(handlerType)
	if err != nil {
		return err
	}
	return handler.ValidateConfig(ctx, raw)
}

func (registry *HandlerRegistry) Execute(ctx context.Context, run *scheduledtaskmodel.Run) (HandlerResult, error) {
	return registry.ExecuteWithLogger(ctx, run, nil)
}

func (registry *HandlerRegistry) ExecuteWithLogger(ctx context.Context, run *scheduledtaskmodel.Run, logger RunLogger) (HandlerResult, error) {
	if run == nil {
		return HandlerResult{}, errors.New("scheduled task run is required")
	}
	var task scheduledtaskmodel.Task
	if err := json.Unmarshal([]byte(run.TaskSnapshotJSON), &task); err != nil {
		return HandlerResult{}, fmt.Errorf("decode scheduled task snapshot: %w", err)
	}
	handler, err := registry.Handler(task.HandlerType)
	if err != nil {
		return HandlerResult{}, err
	}
	return handler.Execute(ctx, RunContext{
		RunID: run.ID, TriggerType: run.TriggerType, Attempt: run.Attempt, Task: task,
		Logger: logger,
	})
}

type NopRunLogger struct{}

func (NopRunLogger) Log(context.Context, string, string, string) error { return nil }

var _ HandlerConfigValidator = (*HandlerRegistry)(nil)
