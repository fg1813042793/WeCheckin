package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type WorkflowNotificationDispatcher interface {
	DispatchDueNotifications(context.Context, int) (int, error)
}

type WorkflowNotificationDispatchJob struct {
	dispatcher WorkflowNotificationDispatcher
}

func NewWorkflowNotificationDispatchJob(dispatcher WorkflowNotificationDispatcher) *WorkflowNotificationDispatchJob {
	return &WorkflowNotificationDispatchJob{dispatcher: dispatcher}
}

func (job *WorkflowNotificationDispatchJob) Key() string  { return "workflow.notification.dispatch_due" }
func (job *WorkflowNotificationDispatchJob) Name() string { return "派发到期流程通知" }
func (job *WorkflowNotificationDispatchJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","minimum":1,"maximum":1000}}}`)
}

func (job *WorkflowNotificationDispatchJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeNotificationDispatchParams(raw)
	return err
}

func (job *WorkflowNotificationDispatchJob) Execute(ctx context.Context, _ string, raw json.RawMessage, _ application.RunLogger) (application.HandlerResult, error) {
	if job == nil || job.dispatcher == nil {
		return application.HandlerResult{}, errors.New("workflow notification dispatcher is not initialized")
	}
	params, err := decodeNotificationDispatchParams(raw)
	if err != nil {
		return application.HandlerResult{}, err
	}
	count, err := job.dispatcher.DispatchDueNotifications(ctx, params.Limit)
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "notification_dispatch_failed", Summary: err.Error(), Temporary: true}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("dispatched %d workflow notifications", count),
		Data:    map[string]interface{}{"dispatched": count},
	}, nil
}

type notificationDispatchParams struct {
	Limit int `json:"limit"`
}

func decodeNotificationDispatchParams(raw json.RawMessage) (notificationDispatchParams, error) {
	params := notificationDispatchParams{Limit: 100}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, fmt.Errorf("decode notification dispatch params: %w", err)
		}
	}
	if params.Limit < 1 || params.Limit > 1000 {
		return params, errors.New("notification dispatch limit must be between 1 and 1000")
	}
	return params, nil
}

type CleanupResult struct {
	RunsDeleted int64 `json:"runsDeleted"`
	LogsDeleted int64 `json:"logsDeleted"`
}

type CleanupStore interface {
	Cleanup(context.Context, int64, int64, int) (CleanupResult, error)
}

type CleanupJob struct {
	store            CleanupStore
	runRetentionDays int
	logRetentionDays int
	now              func() time.Time
}

func NewCleanupJob(store CleanupStore, runRetentionDays, logRetentionDays int, now func() time.Time) *CleanupJob {
	if runRetentionDays <= 0 {
		runRetentionDays = 90
	}
	if logRetentionDays <= 0 {
		logRetentionDays = 30
	}
	if now == nil {
		now = time.Now
	}
	return &CleanupJob{store: store, runRetentionDays: runRetentionDays, logRetentionDays: logRetentionDays, now: now}
}

func (job *CleanupJob) Key() string  { return "scheduled-task.cleanup" }
func (job *CleanupJob) Name() string { return "清理定时任务历史" }
func (job *CleanupJob) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"batchSize":{"type":"integer","minimum":1,"maximum":5000}}}`)
}

func (job *CleanupJob) Validate(_ context.Context, raw json.RawMessage) error {
	_, err := decodeCleanupParams(raw)
	return err
}

func (job *CleanupJob) Execute(ctx context.Context, _ string, raw json.RawMessage, _ application.RunLogger) (application.HandlerResult, error) {
	if job == nil || job.store == nil {
		return application.HandlerResult{}, errors.New("scheduled task cleanup store is not initialized")
	}
	params, err := decodeCleanupParams(raw)
	if err != nil {
		return application.HandlerResult{}, err
	}
	now := job.now().UTC()
	result, err := job.store.Cleanup(
		ctx,
		now.AddDate(0, 0, -job.runRetentionDays).UnixMilli(),
		now.AddDate(0, 0, -job.logRetentionDays).UnixMilli(),
		params.BatchSize,
	)
	if err != nil {
		return application.HandlerResult{}, &application.HandlerError{Code: "cleanup_failed", Summary: err.Error(), Temporary: true}
	}
	return application.HandlerResult{
		Summary: fmt.Sprintf("deleted %d runs and %d log segments", result.RunsDeleted, result.LogsDeleted),
		Data: map[string]interface{}{
			"runsDeleted": result.RunsDeleted, "logsDeleted": result.LogsDeleted,
		},
	}, nil
}

type cleanupParams struct {
	BatchSize int `json:"batchSize"`
}

func decodeCleanupParams(raw json.RawMessage) (cleanupParams, error) {
	params := cleanupParams{BatchSize: 1000}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return params, fmt.Errorf("decode cleanup params: %w", err)
		}
	}
	if params.BatchSize < 1 || params.BatchSize > 5000 {
		return params, errors.New("cleanup batchSize must be between 1 and 5000")
	}
	return params, nil
}

var _ GoJob = (*WorkflowNotificationDispatchJob)(nil)
var _ GoJob = (*CleanupJob)(nil)
