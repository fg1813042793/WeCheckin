package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

var (
	ErrGoJobNotFound          = errors.New("registered Go task not found")
	ErrGoJobAlreadyRegistered = errors.New("registered Go task already exists")
)

type GoJob interface {
	Key() string
	Name() string
	ConfigSchema() json.RawMessage
	Validate(context.Context, json.RawMessage) error
	Execute(context.Context, string, json.RawMessage, application.RunLogger) (application.HandlerResult, error)
}

type GoHandler struct {
	mu   sync.RWMutex
	jobs map[string]GoJob
}

type goHandlerConfig struct {
	HandlerKey string          `json:"handlerKey"`
	Params     json.RawMessage `json:"params"`
}

func NewGoHandler() *GoHandler {
	return &GoHandler{jobs: make(map[string]GoJob)}
}

func (handler *GoHandler) Type() string { return "go" }

func (handler *GoHandler) Metadata() application.HandlerMetadata {
	handler.mu.RLock()
	keys := make([]string, 0, len(handler.jobs))
	labels := make(map[string]string, len(handler.jobs))
	for key, job := range handler.jobs {
		keys = append(keys, key)
		name := strings.TrimSpace(job.Name())
		if name == "" {
			name = key
		}
		labels[key] = name
	}
	handler.mu.RUnlock()
	sort.Strings(keys)
	schema, _ := json.Marshal(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"handlerKey": map[string]interface{}{"type": "string", "enum": keys, "x-enum-labels": labels},
			"params":     map[string]interface{}{"type": "object"},
		},
		"required": []string{"handlerKey"},
	})
	return application.HandlerMetadata{
		Type: "go", Name: "Registered Go handler", Description: "Executes a server-registered Go task",
		RiskLevel: "low", ConfigSchema: schema,
	}
}

func (handler *GoHandler) Register(job GoJob) error {
	if handler == nil || job == nil {
		return errors.New("Go task registration is required")
	}
	key := strings.TrimSpace(job.Key())
	if key == "" {
		return errors.New("Go task key is required")
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	if _, exists := handler.jobs[key]; exists {
		return fmt.Errorf("%w: %s", ErrGoJobAlreadyRegistered, key)
	}
	handler.jobs[key] = job
	return nil
}

func (handler *GoHandler) ValidateConfig(ctx context.Context, raw json.RawMessage) error {
	config, job, err := handler.resolve(raw)
	if err != nil {
		return err
	}
	return job.Validate(ctx, config.Params)
}

func (handler *GoHandler) Execute(ctx context.Context, run application.RunContext) (application.HandlerResult, error) {
	config, job, err := handler.resolve(json.RawMessage(run.Task.HandlerConfigJSON))
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	logger := run.Logger
	if logger == nil {
		logger = application.NopRunLogger{}
	}
	return job.Execute(ctx, run.RunID, config.Params, logger)
}

func (handler *GoHandler) resolve(raw json.RawMessage) (goHandlerConfig, GoJob, error) {
	var config goHandlerConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, nil, fmt.Errorf("decode Go task config: %w", err)
	}
	config.HandlerKey = strings.TrimSpace(config.HandlerKey)
	if config.HandlerKey == "" {
		return config, nil, errors.New("Go task handlerKey is required")
	}
	if len(config.Params) == 0 {
		config.Params = json.RawMessage(`{}`)
	}
	handler.mu.RLock()
	job := handler.jobs[config.HandlerKey]
	handler.mu.RUnlock()
	if job == nil {
		return config, nil, fmt.Errorf("%w: %s", ErrGoJobNotFound, config.HandlerKey)
	}
	return config, job, nil
}

var _ application.Handler = (*GoHandler)(nil)
