package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/domain"
)

type TaskQuery struct {
	Keyword     string
	HandlerType string
	Enabled     *bool
	Page        int
	PageSize    int
}

type TaskList struct {
	List     []scheduledtaskmodel.Task `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

type RunQuery struct {
	TaskID      uint64
	Status      string
	TriggerType string
	WorkerID    string
	StartTime   int64
	EndTime     int64
	Page        int
	PageSize    int
}

type RunList struct {
	List     []scheduledtaskmodel.Run `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

type RunDetail struct {
	Run  *scheduledtaskmodel.Run     `json:"run"`
	Logs []scheduledtaskmodel.RunLog `json:"logs"`
}

type CronPreviewRequest struct {
	Expression string    `json:"expression"`
	Precision  string    `json:"precision"`
	Timezone   string    `json:"timezone"`
	Count      int       `json:"count"`
	After      time.Time `json:"-"`
}

type CronPreviewResult struct {
	Occurrences []domain.PreviewOccurrence `json:"occurrences"`
}

type ManagementStore interface {
	ListTasks(context.Context, TaskQuery) ([]scheduledtaskmodel.Task, int64, error)
	DeleteTask(context.Context, uint64, uint64, int64) error
	ListRuns(context.Context, RunQuery) ([]scheduledtaskmodel.Run, int64, error)
	ListRunLogs(context.Context, string) ([]scheduledtaskmodel.RunLog, error)
	CancelRun(context.Context, string, bool, uint64, int64) error
}

func (service *Service) GetTask(ctx context.Context, taskID uint64) (*scheduledtaskmodel.Task, error) {
	return service.store.GetTask(ctx, taskID)
}

func (service *Service) ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error) {
	store, err := service.managementStore()
	if err != nil {
		return nil, err
	}
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.HandlerType = strings.TrimSpace(query.HandlerType)
	items, total, err := store.ListTasks(ctx, query)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]scheduledtaskmodel.Task, 0)
	}
	return &TaskList{List: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (service *Service) DeleteTask(ctx context.Context, taskID, actorID uint64) error {
	if taskID == 0 {
		return ErrTaskNotFound
	}
	task, err := service.store.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if isSystemTaskCode(task.Code) {
		return ErrSystemTaskReadOnly
	}
	store, err := service.managementStore()
	if err != nil {
		return err
	}
	return store.DeleteTask(ctx, taskID, actorID, service.config.Now().UnixMilli())
}

func (service *Service) SetTaskEnabled(ctx context.Context, taskID, actorID uint64, enabled bool, expectedVersion int64) (*scheduledtaskmodel.Task, error) {
	if expectedVersion < 1 {
		return nil, fmt.Errorf("%w: version is required", ErrInvalidTask)
	}
	task, err := service.store.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if isSystemTaskCode(task.Code) {
		return nil, ErrSystemTaskReadOnly
	}
	nextRunAt := int64(0)
	if enabled {
		schedule, err := domain.ParseSchedule(task.CronPrecision, task.CronExpression, task.Timezone, service.config.MinimumSecondInterval)
		if err != nil {
			return nil, err
		}
		nextRunAt = schedule.Next(service.config.Now().UTC()).UnixMilli()
	}
	task.Enabled = boolInt(enabled)
	task.NextRunAt = nextRunAt
	task.Version = expectedVersion + 1
	task.UpdatedBy = actorID
	task.EditTime = service.config.Now().UnixMilli()
	if err := service.store.UpdateTask(ctx, task, expectedVersion); err != nil {
		return nil, err
	}
	return task, nil
}

func (service *Service) ListRuns(ctx context.Context, query RunQuery) (*RunList, error) {
	store, err := service.managementStore()
	if err != nil {
		return nil, err
	}
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	query.Status = strings.TrimSpace(query.Status)
	query.TriggerType = strings.TrimSpace(query.TriggerType)
	query.WorkerID = strings.TrimSpace(query.WorkerID)
	items, total, err := store.ListRuns(ctx, query)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]scheduledtaskmodel.Run, 0)
	}
	return &RunList{List: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (service *Service) GetRunDetail(ctx context.Context, runID string) (*RunDetail, error) {
	run, err := service.store.GetRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	store, err := service.managementStore()
	if err != nil {
		return nil, err
	}
	logs, err := store.ListRunLogs(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		logs = make([]scheduledtaskmodel.RunLog, 0)
	}
	return &RunDetail{Run: run, Logs: logs}, nil
}

func (service *Service) CancelRun(ctx context.Context, runID string, actorID uint64) (*scheduledtaskmodel.Run, error) {
	run, err := service.store.GetRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	running := run.Status == scheduledtaskmodel.RunStatusRunning
	switch run.Status {
	case scheduledtaskmodel.RunStatusWaiting, scheduledtaskmodel.RunStatusQueued, scheduledtaskmodel.RunStatusRetryWait, scheduledtaskmodel.RunStatusRunning:
	default:
		return nil, ErrRunNotCancelable
	}
	store, err := service.managementStore()
	if err != nil {
		return nil, err
	}
	now := service.config.Now().UnixMilli()
	if err := store.CancelRun(ctx, run.ID, running, actorID, now); err != nil {
		return nil, err
	}
	run.CancelRequestedAt = now
	if !running {
		run.Status = scheduledtaskmodel.RunStatusCanceled
		run.FinishedAt = now
	}
	return run, nil
}

func (service *Service) PreviewCron(request CronPreviewRequest) (CronPreviewResult, error) {
	if request.Precision == "" {
		request.Precision = scheduledtaskmodel.CronPrecisionMinute
	}
	if request.Timezone == "" {
		request.Timezone = "Asia/Shanghai"
	}
	if request.Count <= 0 {
		request.Count = 5
	}
	if request.Count > 20 {
		request.Count = 20
	}
	if request.After.IsZero() {
		request.After = service.config.Now().UTC()
	}
	schedule, err := domain.ParseSchedule(request.Precision, strings.TrimSpace(request.Expression), request.Timezone, service.config.MinimumSecondInterval)
	if err != nil {
		return CronPreviewResult{}, err
	}
	return CronPreviewResult{Occurrences: schedule.Preview(request.After, request.Count)}, nil
}

func (service *Service) HandlerMetadata() []HandlerMetadata {
	provider, ok := service.validator.(interface{ Metadata() []HandlerMetadata })
	if !ok {
		return []HandlerMetadata{}
	}
	return provider.Metadata()
}

func (service *Service) managementStore() (ManagementStore, error) {
	store, ok := service.store.(ManagementStore)
	if !ok {
		return nil, errors.New("scheduled task management store is unavailable")
	}
	return store, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
