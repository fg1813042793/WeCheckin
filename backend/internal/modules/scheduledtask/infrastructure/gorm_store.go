package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/application"
	"wecheckin/backend/internal/modules/scheduledtask/domain"
	scheduledtaskruntime "wecheckin/backend/internal/modules/scheduledtask/runtime"
	"wecheckin/backend/pkg/database"
)

type GormStore struct {
	db    *gorm.DB
	txCtx context.Context
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (store *GormStore) InTransaction(ctx context.Context, fn func(*GormStore) error) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(&GormStore{db: tx, txCtx: db.Statement.Context})
	})
}

func (store *GormStore) CreateTask(ctx context.Context, task *scheduledtaskmodel.Task) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return translateStoreError(db.Create(task).Error)
}

func (store *GormStore) UpdateTask(ctx context.Context, task *scheduledtaskmodel.Task, expectedVersion int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&scheduledtaskmodel.Task{}).
		Where("id = ? AND deleted_at = 0", task.ID).
		Where("version = ?", expectedVersion).
		Select(
			"code", "name", "description", "handler_type", "handler_config_json",
			"cron_expression", "cron_precision", "timezone", "enabled", "misfire_policy",
			"max_catch_up", "concurrency_policy", "timeout_seconds", "max_retries",
			"retry_backoff_json", "last_scheduled_at", "next_run_at", "version",
			"updated_by", "edit_time", "updated_at",
		).Updates(task)
	if result.Error != nil {
		return translateStoreError(result.Error)
	}
	if result.RowsAffected == 0 {
		return application.ErrVersionConflict
	}
	return nil
}

func (store *GormStore) GetTask(ctx context.Context, taskID uint64) (*scheduledtaskmodel.Task, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var task scheduledtaskmodel.Task
	if err := db.First(&task, "id = ? AND deleted_at = 0", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, application.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (store *GormStore) CreateRun(ctx context.Context, run *scheduledtaskmodel.Run) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return translateStoreError(db.Create(run).Error)
}

func (store *GormStore) GetRun(ctx context.Context, runID string) (*scheduledtaskmodel.Run, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var run scheduledtaskmodel.Run
	if err := db.First(&run, "id = ?", runID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, application.ErrRunNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (store *GormStore) FindActiveRun(ctx context.Context, taskID uint64) (*scheduledtaskmodel.Run, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var run scheduledtaskmodel.Run
	err = db.Where("task_id = ? AND run_status IN ?", taskID, []string{
		scheduledtaskmodel.RunStatusWaiting,
		scheduledtaskmodel.RunStatusQueued,
		scheduledtaskmodel.RunStatusRunning,
		scheduledtaskmodel.RunStatusRetryWait,
	}).Order("add_time ASC").First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (store *GormStore) FindWaitingRun(ctx context.Context, taskID uint64) (*scheduledtaskmodel.Run, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var run scheduledtaskmodel.Run
	err = db.Where("task_id = ? AND run_status = ?", taskID, scheduledtaskmodel.RunStatusWaiting).
		Order("add_time ASC").First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (store *GormStore) IncrementRunCoalesced(ctx context.Context, runID string, count int, now int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&scheduledtaskmodel.Run{}).
		Where("id = ? AND run_status = ?", runID, scheduledtaskmodel.RunStatusWaiting).
		Updates(map[string]interface{}{
			"coalesced_count": gorm.Expr("coalesced_count + ?", count),
			"edit_time":       now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrVersionConflict
	}
	return nil
}

func (store *GormStore) MarkRunDispatched(ctx context.Context, runID, messageID string, now int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Model(&scheduledtaskmodel.Run{}).Where("id = ?", runID).Updates(map[string]interface{}{
		"redis_message_id": messageID,
		"queued_at":        now,
		"edit_time":        now,
	}).Error
}

func (store *GormStore) LockDueTasks(ctx context.Context, now int64, limit int) ([]scheduledtaskmodel.Task, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var tasks []scheduledtaskmodel.Task
	err = db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("enabled = 1 AND deleted_at = 0 AND next_run_at > 0 AND next_run_at <= ?", now).
		Order("next_run_at ASC").Limit(limit).Find(&tasks).Error
	return tasks, err
}

func (store *GormStore) TryClaimRun(ctx context.Context, runID, workerID string, now int64) (bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	result := db.Model(&scheduledtaskmodel.Run{}).
		Where("id = ?", runID).
		Where("run_status = ?", scheduledtaskmodel.RunStatusQueued).
		Updates(map[string]interface{}{
			"run_status":   scheduledtaskmodel.RunStatusRunning,
			"worker_id":    workerID,
			"started_at":   now,
			"heartbeat_at": now,
			"edit_time":    now,
		})
	return result.RowsAffected == 1, result.Error
}

func (store *GormStore) ListTasks(ctx context.Context, query application.TaskQuery) ([]scheduledtaskmodel.Task, int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer cancel()
	statement := db.Model(&scheduledtaskmodel.Task{}).Where("deleted_at = 0")
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		statement = statement.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if handlerType := strings.TrimSpace(query.HandlerType); handlerType != "" {
		statement = statement.Where("handler_type = ?", handlerType)
	}
	if query.Enabled != nil {
		statement = statement.Where("enabled = ?", boolToInt(*query.Enabled))
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]scheduledtaskmodel.Task, 0)
	err = statement.Order("id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}

func (store *GormStore) DeleteTask(ctx context.Context, taskID, actorID uint64, now int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&scheduledtaskmodel.Task{}).
		Where("id = ? AND deleted_at = 0", taskID).
		Updates(map[string]interface{}{
			"enabled":     0,
			"next_run_at": 0,
			"deleted_at":  now,
			"updated_by":  actorID,
			"edit_time":   now,
			"version":     gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrTaskNotFound
	}
	return nil
}

func (store *GormStore) ListRuns(ctx context.Context, query application.RunQuery) ([]scheduledtaskmodel.Run, int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer cancel()
	statement := db.Model(&scheduledtaskmodel.Run{})
	if query.TaskID > 0 {
		statement = statement.Where("task_id = ?", query.TaskID)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		statement = statement.Where("run_status = ?", status)
	}
	if triggerType := strings.TrimSpace(query.TriggerType); triggerType != "" {
		statement = statement.Where("trigger_type = ?", triggerType)
	}
	if workerID := strings.TrimSpace(query.WorkerID); workerID != "" {
		statement = statement.Where("worker_id = ?", workerID)
	}
	if query.StartTime > 0 {
		statement = statement.Where("scheduled_at >= ?", query.StartTime)
	}
	if query.EndTime > 0 {
		statement = statement.Where("scheduled_at <= ?", query.EndTime)
	}
	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]scheduledtaskmodel.Run, 0)
	err = statement.Order("scheduled_at DESC, id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}

func (store *GormStore) ListRunLogs(ctx context.Context, runID string) ([]scheduledtaskmodel.RunLog, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	items := make([]scheduledtaskmodel.RunLog, 0)
	err = db.Where("run_id = ?", runID).Order("log_sequence ASC, id ASC").Find(&items).Error
	return items, err
}

func (store *GormStore) CancelRun(ctx context.Context, runID string, running bool, _ uint64, now int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	updates := map[string]interface{}{
		"cancel_requested_at": now,
		"edit_time":           now,
	}
	statuses := []string{scheduledtaskmodel.RunStatusRunning}
	if !running {
		statuses = []string{
			scheduledtaskmodel.RunStatusWaiting,
			scheduledtaskmodel.RunStatusQueued,
			scheduledtaskmodel.RunStatusRetryWait,
		}
		updates["run_status"] = scheduledtaskmodel.RunStatusCanceled
		updates["finished_at"] = now
	}
	result := db.Model(&scheduledtaskmodel.Run{}).
		Where("id = ? AND run_status IN ?", runID, statuses).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrRunNotCancelable
	}
	return nil
}

func (store *GormStore) GenerateDueRuns(ctx context.Context, now int64, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	runIDs := make([]string, 0)
	err := store.InTransaction(ctx, func(tx *GormStore) error {
		tasks, err := tx.LockDueTasks(ctx, now, limit)
		if err != nil {
			return err
		}
		service := application.NewService(tx, nil, nil, application.ServiceConfig{})
		nowTime := time.UnixMilli(now).UTC()
		for index := range tasks {
			task := &tasks[index]
			schedule, err := domain.ParseSchedule(task.CronPrecision, task.CronExpression, task.Timezone, 0)
			if err != nil {
				return fmt.Errorf("parse scheduled task %d: %w", task.ID, err)
			}
			cursor := time.UnixMilli(task.LastScheduledAt).UTC()
			if task.LastScheduledAt <= 0 {
				cursor = time.UnixMilli(task.NextRunAt).UTC().Add(-time.Nanosecond)
			}
			due, err := schedule.ComputeDue(cursor, nowTime, task.MisfirePolicy, task.MaxCatchUp)
			if err != nil {
				return fmt.Errorf("compute scheduled task %d: %w", task.ID, err)
			}
			for _, occurrence := range due.Runs {
				triggerType := scheduledtaskmodel.TriggerTypeScheduled
				if occurrence.CoalescedCount > 1 || due.SkippedCount > 0 {
					triggerType = scheduledtaskmodel.TriggerTypeMisfire
				}
				result, err := service.CreateScheduledRun(ctx, task, occurrence.ScheduledAt, occurrence.CoalescedCount, triggerType)
				if errors.Is(err, application.ErrDuplicateRunKey) {
					continue
				}
				if err != nil {
					return err
				}
				if !result.Merged && result.Run != nil && result.Run.Status == scheduledtaskmodel.RunStatusQueued {
					runIDs = append(runIDs, result.Run.ID)
				}
			}
			expectedVersion := task.Version
			task.LastScheduledAt = now
			task.NextRunAt = 0
			if !due.Next.IsZero() {
				task.NextRunAt = due.Next.UnixMilli()
			}
			task.Version++
			task.EditTime = now
			if err := tx.UpdateTask(ctx, task, expectedVersion); err != nil {
				return err
			}
		}
		return nil
	})
	return runIDs, err
}

func (store *GormStore) ListUndeliveredRuns(ctx context.Context, queuedBefore int64, limit int) ([]string, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var ids []string
	err = db.Model(&scheduledtaskmodel.Run{}).
		Where("run_status = ? AND redis_message_id = '' AND queued_at <= ?", scheduledtaskmodel.RunStatusQueued, queuedBefore).
		Order("queued_at ASC").Limit(limit).Pluck("id", &ids).Error
	return ids, err
}

func (store *GormStore) WakeRetryRuns(ctx context.Context, now int64, limit int) ([]string, error) {
	return store.transitionRuns(ctx, scheduledtaskmodel.RunStatusRetryWait, now, limit,
		func(db *gorm.DB) *gorm.DB { return db.Where("next_retry_at > 0 AND next_retry_at <= ?", now) })
}

func (store *GormStore) WakeWaitingRuns(ctx context.Context, now int64, limit int) ([]string, error) {
	return store.transitionRuns(ctx, scheduledtaskmodel.RunStatusWaiting, now, limit, func(db *gorm.DB) *gorm.DB {
		return db.Where(`NOT EXISTS (
			SELECT 1 FROM scheduled_task_runs active
			WHERE active.task_id = scheduled_task_runs.task_id
			AND active.id <> scheduled_task_runs.id
			AND active.run_status IN ?
		)`, []string{scheduledtaskmodel.RunStatusQueued, scheduledtaskmodel.RunStatusRunning, scheduledtaskmodel.RunStatusRetryWait})
	})
}

func (store *GormStore) transitionRuns(
	ctx context.Context,
	fromStatus string,
	now int64,
	limit int,
	filter func(*gorm.DB) *gorm.DB,
) ([]string, error) {
	ids := make([]string, 0)
	err := store.InTransaction(ctx, func(tx *GormStore) error {
		db, cancel, err := tx.contextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()
		query := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("run_status = ?", fromStatus)
		if filter != nil {
			query = filter(query)
		}
		var runs []scheduledtaskmodel.Run
		if err := query.Order("scheduled_at ASC").Limit(limit).Find(&runs).Error; err != nil {
			return err
		}
		for _, run := range runs {
			result := db.Model(&scheduledtaskmodel.Run{}).
				Where("id = ? AND run_status = ?", run.ID, fromStatus).
				Updates(map[string]interface{}{
					"run_status":       scheduledtaskmodel.RunStatusQueued,
					"queued_at":        now,
					"redis_message_id": "",
					"worker_id":        "",
					"started_at":       0,
					"heartbeat_at":     0,
					"next_retry_at":    0,
					"edit_time":        now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 {
				ids = append(ids, run.ID)
			}
		}
		return nil
	})
	return ids, err
}

func (store *GormStore) RecoverStaleRuns(ctx context.Context, staleBefore int64, limit int) ([]string, error) {
	err := store.InTransaction(ctx, func(tx *GormStore) error {
		db, cancel, err := tx.contextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()
		var runs []scheduledtaskmodel.Run
		if err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("run_status = ? AND heartbeat_at > 0 AND heartbeat_at <= ?", scheduledtaskmodel.RunStatusRunning, staleBefore).
			Order("heartbeat_at ASC").Limit(limit).Find(&runs).Error; err != nil {
			return err
		}
		for _, run := range runs {
			updates := map[string]interface{}{
				"finished_at":      staleBefore,
				"heartbeat_at":     0,
				"redis_message_id": "",
				"edit_time":        staleBefore,
				"error_code":       "worker_stale",
				"error_summary":    "worker heartbeat expired",
			}
			if run.CancelRequestedAt > 0 {
				updates["run_status"] = scheduledtaskmodel.RunStatusCanceled
				updates["error_code"] = "canceled"
				updates["error_summary"] = "cancellation recovered after worker heartbeat expired"
			} else {
				var task scheduledtaskmodel.Task
				_ = json.Unmarshal([]byte(run.TaskSnapshotJSON), &task)
				if run.Attempt < task.MaxRetries {
					updates["run_status"] = scheduledtaskmodel.RunStatusRetryWait
					updates["attempt"] = run.Attempt + 1
					updates["next_retry_at"] = staleBefore
					updates["finished_at"] = 0
				} else {
					updates["run_status"] = scheduledtaskmodel.RunStatusFailed
				}
			}
			if err := db.Model(&scheduledtaskmodel.Run{}).
				Where("id = ? AND run_status = ?", run.ID, scheduledtaskmodel.RunStatusRunning).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return nil, err
}

func (store *GormStore) CompleteRun(ctx context.Context, runID string, outcome scheduledtaskruntime.RunOutcome) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	updates := map[string]interface{}{
		"run_status":     outcome.Status,
		"attempt":        outcome.Attempt,
		"next_retry_at":  outcome.NextRetryAt,
		"finished_at":    outcome.FinishedAt,
		"result_summary": outcome.ResultSummary,
		"error_code":     outcome.ErrorCode,
		"error_summary":  outcome.ErrorSummary,
		"heartbeat_at":   0,
		"edit_time":      time.Now().UnixMilli(),
	}
	if outcome.Status == scheduledtaskmodel.RunStatusRetryWait {
		updates["redis_message_id"] = ""
	}
	result := db.Model(&scheduledtaskmodel.Run{}).
		Where("id = ? AND run_status = ?", runID, scheduledtaskmodel.RunStatusRunning).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrVersionConflict
	}
	return nil
}

func (store *GormStore) HeartbeatRun(ctx context.Context, runID, workerID string, now int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Model(&scheduledtaskmodel.Run{}).
		Where("id = ? AND run_status = ? AND worker_id = ?", runID, scheduledtaskmodel.RunStatusRunning, workerID).
		Updates(map[string]interface{}{"heartbeat_at": now, "edit_time": now}).Error
}

func (store *GormStore) Cleanup(ctx context.Context, runBefore, logBefore int64, limit int) (CleanupResult, error) {
	if limit <= 0 {
		limit = 1000
	}
	result := CleanupResult{}
	err := store.InTransaction(ctx, func(tx *GormStore) error {
		db, cancel, err := tx.contextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()
		var oldLogIDs []uint64
		if err := db.Model(&scheduledtaskmodel.RunLog{}).
			Where("log_time > 0 AND log_time < ?", logBefore).
			Order("id ASC").Limit(limit).Pluck("id", &oldLogIDs).Error; err != nil {
			return err
		}
		if len(oldLogIDs) > 0 {
			deleted := db.Where("id IN ?", oldLogIDs).Delete(&scheduledtaskmodel.RunLog{})
			if deleted.Error != nil {
				return deleted.Error
			}
			result.LogsDeleted += deleted.RowsAffected
		}
		var oldRunIDs []string
		if err := db.Model(&scheduledtaskmodel.Run{}).
			Where("run_status IN ? AND finished_at > 0 AND finished_at < ?", []string{
				scheduledtaskmodel.RunStatusSuccess,
				scheduledtaskmodel.RunStatusFailed,
				scheduledtaskmodel.RunStatusCanceled,
				scheduledtaskmodel.RunStatusSkipped,
			}, runBefore).
			Order("finished_at ASC").Limit(limit).Pluck("id", &oldRunIDs).Error; err != nil {
			return err
		}
		if len(oldRunIDs) == 0 {
			return nil
		}
		deletedLogs := db.Where("run_id IN ?", oldRunIDs).Delete(&scheduledtaskmodel.RunLog{})
		if deletedLogs.Error != nil {
			return deletedLogs.Error
		}
		result.LogsDeleted += deletedLogs.RowsAffected
		deletedRuns := db.Where("id IN ?", oldRunIDs).Delete(&scheduledtaskmodel.Run{})
		if deletedRuns.Error != nil {
			return deletedRuns.Error
		}
		result.RunsDeleted += deletedRuns.RowsAffected
		return nil
	})
	return result, err
}

func (store *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if store == nil || store.db == nil {
		return nil, func() {}, errors.New("scheduled task database is not initialized")
	}
	if store.txCtx != nil {
		return store.db.WithContext(store.txCtx), func() {}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return store.db.WithContext(queryCtx), cancel, nil
}

func translateStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return application.ErrDuplicateRunKey
	}
	var mysqlError *mysqlDriver.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return application.ErrDuplicateRunKey
	}
	return fmt.Errorf("scheduled task store: %w", err)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

var _ application.Store = (*GormStore)(nil)
var _ application.ManagementStore = (*GormStore)(nil)
var _ scheduledtaskruntime.SchedulerStore = (*GormStore)(nil)
var _ scheduledtaskruntime.WorkerStore = (*GormStore)(nil)
