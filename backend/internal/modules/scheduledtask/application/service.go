package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/domain"
)

var (
	ErrTaskNotFound       = errors.New("scheduled task not found")
	ErrRunNotFound        = errors.New("scheduled task run not found")
	ErrVersionConflict    = errors.New("scheduled task version conflict")
	ErrDuplicateRunKey    = errors.New("scheduled task run already exists")
	ErrRunNotRetryable    = errors.New("scheduled task run is not retryable")
	ErrRunNotCancelable   = errors.New("scheduled task run is not cancelable")
	ErrInvalidTask        = errors.New("invalid scheduled task")
	ErrHandlerUnavailable = errors.New("scheduled task handler unavailable")
	ErrSystemTaskReadOnly = errors.New("system scheduled task is read-only")
)

var taskCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{1,99}$`)

type Store interface {
	CreateTask(context.Context, *scheduledtaskmodel.Task) error
	UpdateTask(context.Context, *scheduledtaskmodel.Task, int64) error
	GetTask(context.Context, uint64) (*scheduledtaskmodel.Task, error)
	CreateRun(context.Context, *scheduledtaskmodel.Run) error
	GetRun(context.Context, string) (*scheduledtaskmodel.Run, error)
	FindActiveRun(context.Context, uint64) (*scheduledtaskmodel.Run, error)
	FindWaitingRun(context.Context, uint64) (*scheduledtaskmodel.Run, error)
	IncrementRunCoalesced(context.Context, string, int, int64) error
	MarkRunDispatched(context.Context, string, string, int64) error
}

type HandlerConfigValidator interface {
	ValidateConfig(context.Context, string, json.RawMessage) error
}

type QueuePublisher interface {
	PublishRun(context.Context, string) (string, error)
}

type ServiceConfig struct {
	MinimumSecondInterval time.Duration
	Now                   func() time.Time
	NewRunID              func() string
}

type Service struct {
	store     Store
	validator HandlerConfigValidator
	publisher QueuePublisher
	config    ServiceConfig
}

type CreateTaskRequest struct {
	Code              string          `json:"code"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	HandlerType       string          `json:"handlerType"`
	HandlerConfigJSON json.RawMessage `json:"handlerConfigJson"`
	CronExpression    string          `json:"cronExpression"`
	CronPrecision     string          `json:"cronPrecision"`
	Timezone          string          `json:"timezone"`
	Enabled           bool            `json:"enabled"`
	MisfirePolicy     string          `json:"misfirePolicy"`
	MaxCatchUp        int             `json:"maxCatchUp"`
	ConcurrencyPolicy string          `json:"concurrencyPolicy"`
	TimeoutSeconds    int             `json:"timeoutSeconds"`
	MaxRetries        int             `json:"maxRetries"`
	RetryBackoffJSON  json.RawMessage `json:"retryBackoffJson"`
}

type UpdateTaskRequest struct {
	CreateTaskRequest
	Version int64 `json:"version"`
}

type ScheduledRunResult struct {
	Run    *scheduledtaskmodel.Run
	Merged bool
}

type DispatchResult struct {
	Run             *scheduledtaskmodel.Run `json:"run"`
	DispatchPending bool                    `json:"dispatchPending"`
}

func NewService(store Store, validator HandlerConfigValidator, publisher QueuePublisher, cfg ServiceConfig) *Service {
	if cfg.MinimumSecondInterval <= 0 {
		cfg.MinimumSecondInterval = 5 * time.Second
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.NewRunID == nil {
		cfg.NewRunID = randomRunID
	}
	return &Service{store: store, validator: validator, publisher: publisher, config: cfg}
}

func (service *Service) CreateTask(ctx context.Context, actorID uint64, request CreateTaskRequest) (*scheduledtaskmodel.Task, error) {
	if isSystemTaskCode(request.Code) {
		return nil, ErrSystemTaskReadOnly
	}
	normalized, schedule, err := service.validateTaskRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	now := service.config.Now().UTC()
	task := &scheduledtaskmodel.Task{
		Code: normalized.Code, Name: normalized.Name, Description: normalized.Description,
		HandlerType: normalized.HandlerType, HandlerConfigJSON: string(normalized.HandlerConfigJSON),
		CronExpression: normalized.CronExpression, CronPrecision: normalized.CronPrecision, Timezone: normalized.Timezone,
		Enabled: boolInt(normalized.Enabled), MisfirePolicy: normalized.MisfirePolicy, MaxCatchUp: normalized.MaxCatchUp,
		ConcurrencyPolicy: normalized.ConcurrencyPolicy, TimeoutSeconds: normalized.TimeoutSeconds, MaxRetries: normalized.MaxRetries,
		RetryBackoffJSON: string(normalized.RetryBackoffJSON), Version: 1, CreatedBy: actorID, UpdatedBy: actorID,
		AddTime: now.UnixMilli(), EditTime: now.UnixMilli(),
	}
	if normalized.Enabled {
		task.NextRunAt = schedule.Next(now).UnixMilli()
	}
	if err := service.store.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (service *Service) UpdateTask(ctx context.Context, taskID, actorID uint64, request UpdateTaskRequest) (*scheduledtaskmodel.Task, error) {
	if request.Version < 1 {
		return nil, fmt.Errorf("%w: version is required", ErrInvalidTask)
	}
	existing, err := service.store.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if isSystemTaskCode(existing.Code) || isSystemTaskCode(request.Code) {
		return nil, ErrSystemTaskReadOnly
	}
	normalized, schedule, err := service.validateTaskRequest(ctx, request.CreateTaskRequest)
	if err != nil {
		return nil, err
	}
	now := service.config.Now().UTC()
	existing.Code = normalized.Code
	existing.Name = normalized.Name
	existing.Description = normalized.Description
	existing.HandlerType = normalized.HandlerType
	existing.HandlerConfigJSON = string(normalized.HandlerConfigJSON)
	existing.CronExpression = normalized.CronExpression
	existing.CronPrecision = normalized.CronPrecision
	existing.Timezone = normalized.Timezone
	existing.Enabled = boolInt(normalized.Enabled)
	existing.MisfirePolicy = normalized.MisfirePolicy
	existing.MaxCatchUp = normalized.MaxCatchUp
	existing.ConcurrencyPolicy = normalized.ConcurrencyPolicy
	existing.TimeoutSeconds = normalized.TimeoutSeconds
	existing.MaxRetries = normalized.MaxRetries
	existing.RetryBackoffJSON = string(normalized.RetryBackoffJSON)
	existing.UpdatedBy = actorID
	existing.EditTime = now.UnixMilli()
	existing.Version = request.Version + 1
	existing.NextRunAt = 0
	if normalized.Enabled {
		existing.NextRunAt = schedule.Next(now).UnixMilli()
	}
	if err := service.store.UpdateTask(ctx, existing, request.Version); err != nil {
		return nil, err
	}
	return existing, nil
}

func (service *Service) CreateScheduledRun(ctx context.Context, task *scheduledtaskmodel.Task, scheduledAt time.Time, coalescedCount int, triggerType string) (ScheduledRunResult, error) {
	if task == nil || task.ID == 0 {
		return ScheduledRunResult{}, fmt.Errorf("%w: task is required", ErrInvalidTask)
	}
	if coalescedCount < 1 {
		coalescedCount = 1
	}
	active, err := service.store.FindActiveRun(ctx, task.ID)
	if err != nil {
		return ScheduledRunResult{}, err
	}
	status := scheduledtaskmodel.RunStatusQueued
	summary := ""
	if active != nil {
		switch task.ConcurrencyPolicy {
		case scheduledtaskmodel.ConcurrencyPolicySkip:
			status = scheduledtaskmodel.RunStatusSkipped
			summary = "skipped because another run is active"
		case scheduledtaskmodel.ConcurrencyPolicyQueue:
			waiting, err := service.store.FindWaitingRun(ctx, task.ID)
			if err != nil {
				return ScheduledRunResult{}, err
			}
			if waiting != nil {
				if err := service.store.IncrementRunCoalesced(ctx, waiting.ID, coalescedCount, service.config.Now().UnixMilli()); err != nil {
					return ScheduledRunResult{}, err
				}
				return ScheduledRunResult{Run: waiting, Merged: true}, nil
			}
			status = scheduledtaskmodel.RunStatusWaiting
		}
	}
	run, err := service.newRun(task, scheduledAt, triggerType, status, "", coalescedCount)
	if err != nil {
		return ScheduledRunResult{}, err
	}
	run.ResultSummary = summary
	if err := service.store.CreateRun(ctx, run); err != nil {
		return ScheduledRunResult{}, err
	}
	return ScheduledRunResult{Run: run}, nil
}

func (service *Service) RunNow(ctx context.Context, taskID, _ uint64) (DispatchResult, error) {
	task, err := service.store.GetTask(ctx, taskID)
	if err != nil {
		return DispatchResult{}, err
	}
	run, err := service.newRun(task, service.config.Now().UTC(), scheduledtaskmodel.TriggerTypeManual, scheduledtaskmodel.RunStatusQueued, "", 1)
	if err != nil {
		return DispatchResult{}, err
	}
	if err := service.store.CreateRun(ctx, run); err != nil {
		return DispatchResult{}, err
	}
	return service.dispatch(ctx, run), nil
}

func (service *Service) RetryRun(ctx context.Context, runID string, _ uint64) (DispatchResult, error) {
	parent, err := service.store.GetRun(ctx, runID)
	if err != nil {
		return DispatchResult{}, err
	}
	if parent.Status != scheduledtaskmodel.RunStatusFailed {
		return DispatchResult{}, ErrRunNotRetryable
	}
	task, err := service.store.GetTask(ctx, parent.TaskID)
	if err != nil {
		return DispatchResult{}, err
	}
	run, err := service.newRun(task, service.config.Now().UTC(), scheduledtaskmodel.TriggerTypeManualRetry, scheduledtaskmodel.RunStatusQueued, parent.ID, 1)
	if err != nil {
		return DispatchResult{}, err
	}
	if err := service.store.CreateRun(ctx, run); err != nil {
		return DispatchResult{}, err
	}
	return service.dispatch(ctx, run), nil
}

func (service *Service) dispatch(ctx context.Context, run *scheduledtaskmodel.Run) DispatchResult {
	result := DispatchResult{Run: run}
	if service.publisher == nil {
		result.DispatchPending = true
		return result
	}
	messageID, err := service.publisher.PublishRun(ctx, run.ID)
	if err != nil {
		result.DispatchPending = true
		return result
	}
	if messageID != "" {
		if err := service.store.MarkRunDispatched(ctx, run.ID, messageID, service.config.Now().UnixMilli()); err != nil {
			result.DispatchPending = true
		}
	}
	return result
}

func (service *Service) newRun(task *scheduledtaskmodel.Task, scheduledAt time.Time, triggerType, status, parentRunID string, coalescedCount int) (*scheduledtaskmodel.Run, error) {
	snapshot, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	runID := service.config.NewRunID()
	runKey := fmt.Sprintf("task:%d:%s:%s", task.ID, triggerType, runID)
	if triggerType == scheduledtaskmodel.TriggerTypeScheduled || triggerType == scheduledtaskmodel.TriggerTypeMisfire {
		runKey = fmt.Sprintf("task:%d:scheduled:%d", task.ID, scheduledAt.UTC().UnixMilli())
	}
	nowMillis := service.config.Now().UnixMilli()
	run := &scheduledtaskmodel.Run{
		ID: runID, RunKey: runKey, TaskID: task.ID, ParentRunID: parentRunID,
		TriggerType: triggerType, Status: status, TaskSnapshotJSON: string(snapshot),
		ScheduledAt: scheduledAt.UTC().UnixMilli(), CoalescedCount: coalescedCount,
		AddTime: nowMillis, EditTime: nowMillis,
	}
	if status == scheduledtaskmodel.RunStatusQueued {
		run.QueuedAt = nowMillis
	}
	if status == scheduledtaskmodel.RunStatusSkipped {
		run.FinishedAt = nowMillis
	}
	return run, nil
}

func (service *Service) validateTaskRequest(ctx context.Context, request CreateTaskRequest) (CreateTaskRequest, *domain.Schedule, error) {
	request.Code = strings.TrimSpace(request.Code)
	request.Name = strings.TrimSpace(request.Name)
	request.HandlerType = strings.TrimSpace(request.HandlerType)
	request.CronExpression = strings.TrimSpace(request.CronExpression)
	request.CronPrecision = strings.TrimSpace(request.CronPrecision)
	request.Timezone = strings.TrimSpace(request.Timezone)
	if !taskCodePattern.MatchString(request.Code) {
		return request, nil, fmt.Errorf("%w: invalid code", ErrInvalidTask)
	}
	if request.Name == "" || len(request.Name) > 200 {
		return request, nil, fmt.Errorf("%w: invalid name", ErrInvalidTask)
	}
	if len(request.HandlerConfigJSON) == 0 {
		request.HandlerConfigJSON = json.RawMessage(`{}`)
	}
	var configValue map[string]interface{}
	if err := json.Unmarshal(request.HandlerConfigJSON, &configValue); err != nil {
		return request, nil, fmt.Errorf("%w: handler config must be a JSON object", ErrInvalidTask)
	}
	if service.validator == nil {
		return request, nil, ErrHandlerUnavailable
	}
	if err := service.validator.ValidateConfig(ctx, request.HandlerType, request.HandlerConfigJSON); err != nil {
		return request, nil, fmt.Errorf("%w: %v", ErrInvalidTask, err)
	}
	if request.CronPrecision == "" {
		request.CronPrecision = scheduledtaskmodel.CronPrecisionMinute
	}
	if request.Timezone == "" {
		request.Timezone = "Asia/Shanghai"
	}
	schedule, err := domain.ParseSchedule(request.CronPrecision, request.CronExpression, request.Timezone, service.config.MinimumSecondInterval)
	if err != nil {
		return request, nil, err
	}
	if request.MisfirePolicy == "" {
		request.MisfirePolicy = scheduledtaskmodel.MisfirePolicySkip
	}
	if err := domain.ValidateMisfirePolicy(request.MisfirePolicy); err != nil {
		return request, nil, err
	}
	if request.MaxCatchUp <= 0 {
		request.MaxCatchUp = 1
	}
	if request.ConcurrencyPolicy == "" {
		request.ConcurrencyPolicy = scheduledtaskmodel.ConcurrencyPolicySkip
	}
	if err := domain.ValidateConcurrencyPolicy(request.ConcurrencyPolicy); err != nil {
		return request, nil, err
	}
	if request.TimeoutSeconds == 0 {
		request.TimeoutSeconds = 300
	}
	if request.TimeoutSeconds < 1 || request.TimeoutSeconds > 86400 {
		return request, nil, fmt.Errorf("%w: timeout seconds out of range", ErrInvalidTask)
	}
	if request.MaxRetries < 0 || request.MaxRetries > 5 {
		return request, nil, fmt.Errorf("%w: max retries out of range", ErrInvalidTask)
	}
	if len(request.RetryBackoffJSON) == 0 {
		request.RetryBackoffJSON = json.RawMessage(`{"type":"fixed","seconds":30}`)
	}
	if !json.Valid(request.RetryBackoffJSON) {
		return request, nil, fmt.Errorf("%w: invalid retry backoff JSON", ErrInvalidTask)
	}
	return request, schedule, nil
}

func randomRunID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	return "run-" + hex.EncodeToString(value[:])
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func isSystemTaskCode(code string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(code)), "system.")
}
