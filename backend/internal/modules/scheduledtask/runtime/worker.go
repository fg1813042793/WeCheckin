package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type QueueMessage = application.QueueMessage
type WorkerHeartbeat = application.WorkerHeartbeat

type WorkerStore interface {
	TryClaimRun(context.Context, string, string, int64) (bool, error)
	GetRun(context.Context, string) (*scheduledtaskmodel.Run, error)
	CompleteRun(context.Context, string, RunOutcome) error
	HeartbeatRun(context.Context, string, string, int64) error
}

type WorkerQueue interface {
	EnsureGroup(context.Context) error
	Read(context.Context, string, int64, time.Duration) ([]QueueMessage, error)
	Ack(context.Context, ...string) error
	AutoClaim(context.Context, string, time.Duration, string, int64) ([]QueueMessage, string, error)
	HeartbeatWorker(context.Context, WorkerHeartbeat, time.Duration) error
}

type RunExecutor interface {
	Execute(context.Context, *scheduledtaskmodel.Run) (application.HandlerResult, error)
}

type ExecutionResult = application.HandlerResult
type ExecutionError = application.HandlerError

type RunOutcome struct {
	Status        string
	Attempt       int
	NextRetryAt   int64
	FinishedAt    int64
	ResultSummary string
	ErrorCode     string
	ErrorSummary  string
}

type WorkerConfig struct {
	WorkerID          string
	Role              string
	Version           string
	WorkerCount       int
	BatchSize         int64
	PollBlock         time.Duration
	ClaimIdle         time.Duration
	HeartbeatInterval time.Duration
	WorkerTTL         time.Duration
	DefaultRetryDelay time.Duration
	RetryInitialDelay time.Duration
	RetryMaxDelay     time.Duration
	Logf              logFunc
	Now               func() time.Time
}

type Worker struct {
	store       WorkerStore
	queue       WorkerQueue
	executor    RunExecutor
	config      WorkerConfig
	startedAt   int64
	currentRuns atomic.Int64
}

func NewWorker(store WorkerStore, queue WorkerQueue, executor RunExecutor, cfg WorkerConfig) *Worker {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = int64(cfg.WorkerCount)
	}
	if cfg.PollBlock <= 0 {
		cfg.PollBlock = 5 * time.Second
	}
	if cfg.ClaimIdle <= 0 {
		cfg.ClaimIdle = time.Minute
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 10 * time.Second
	}
	if cfg.WorkerTTL <= 0 {
		cfg.WorkerTTL = 30 * time.Second
	}
	if cfg.DefaultRetryDelay <= 0 {
		cfg.DefaultRetryDelay = 30 * time.Second
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Role == "" {
		cfg.Role = "worker"
	}
	return &Worker{store: store, queue: queue, executor: executor, config: cfg, startedAt: cfg.Now().UnixMilli()}
}

func (worker *Worker) Run(ctx context.Context) error {
	if worker == nil || worker.store == nil || worker.queue == nil || worker.executor == nil {
		return errors.New("scheduled task worker is not initialized")
	}
	if worker.config.WorkerID == "" {
		return errors.New("scheduled task worker ID is required")
	}
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go worker.runWorkerHeartbeat(heartbeatCtx)

	claimCursor := "0-0"
	queueReady := false
	retry := newRetryController(
		configuredRetryInitial(worker.config.RetryInitialDelay, worker.config.PollBlock),
		worker.config.RetryMaxDelay,
		worker.config.Logf,
	)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !queueReady {
			if err := worker.queue.EnsureGroup(ctx); err != nil {
				if waitErr := retry.WaitAfterFailure(ctx, "worker", "ensure-group", err); waitErr != nil {
					return waitErr
				}
				continue
			}
			queueReady = true
		}
		claimed, next, err := worker.queue.AutoClaim(ctx, worker.config.WorkerID, worker.config.ClaimIdle, claimCursor, worker.config.BatchSize)
		if err != nil {
			queueReady = false
			if waitErr := retry.WaitAfterFailure(ctx, "worker", "auto-claim", err); waitErr != nil {
				return waitErr
			}
			continue
		}
		claimCursor = next
		if err := worker.processBatch(ctx, claimed); err != nil {
			queueReady = false
			if waitErr := retry.WaitAfterFailure(ctx, "worker", "process-claimed", err); waitErr != nil {
				return waitErr
			}
			continue
		}
		messages, err := worker.queue.Read(ctx, worker.config.WorkerID, worker.config.BatchSize, worker.config.PollBlock)
		if err != nil {
			queueReady = false
			if waitErr := retry.WaitAfterFailure(ctx, "worker", "read", err); waitErr != nil {
				return waitErr
			}
			continue
		}
		if err := worker.processBatch(ctx, messages); err != nil {
			queueReady = false
			if waitErr := retry.WaitAfterFailure(ctx, "worker", "process-message", err); waitErr != nil {
				return waitErr
			}
			continue
		}
		retry.MarkSuccess("worker")
	}
}

func (worker *Worker) ProcessMessage(ctx context.Context, message QueueMessage) error {
	now := worker.config.Now().UnixMilli()
	claimed, err := worker.store.TryClaimRun(ctx, message.RunID, worker.config.WorkerID, now)
	if err != nil {
		return err
	}
	if !claimed {
		return worker.queue.Ack(ctx, message.MessageID)
	}
	run, err := worker.store.GetRun(ctx, message.RunID)
	if err != nil {
		return err
	}
	if run.CancelRequestedAt > 0 {
		return worker.persistThenAck(ctx, message.MessageID, run.ID, RunOutcome{
			Status: scheduledtaskmodel.RunStatusCanceled, Attempt: run.Attempt, FinishedAt: now,
			ErrorCode: "canceled", ErrorSummary: "cancellation requested before execution",
		})
	}

	var task scheduledtaskmodel.Task
	if err := json.Unmarshal([]byte(run.TaskSnapshotJSON), &task); err != nil {
		return worker.persistThenAck(ctx, message.MessageID, run.ID, RunOutcome{
			Status: scheduledtaskmodel.RunStatusFailed, Attempt: run.Attempt, FinishedAt: now,
			ErrorCode: "invalid_snapshot", ErrorSummary: "task snapshot is invalid",
		})
	}
	if task.TimeoutSeconds <= 0 {
		task.TimeoutSeconds = 300
	}
	executionCtx, cancel := context.WithTimeout(ctx, time.Duration(task.TimeoutSeconds)*time.Second)
	stopHeartbeat := make(chan struct{})
	go worker.runRunHeartbeat(executionCtx, run.ID, stopHeartbeat, cancel)
	worker.currentRuns.Add(1)
	result, executeErr := worker.executor.Execute(executionCtx, run)
	worker.currentRuns.Add(-1)
	close(stopHeartbeat)
	timedOut := errors.Is(executionCtx.Err(), context.DeadlineExceeded)
	cancel()

	if ctx.Err() != nil {
		return ctx.Err()
	}
	now = worker.config.Now().UnixMilli()
	latest, err := worker.store.GetRun(ctx, run.ID)
	if err != nil {
		return err
	}
	if latest.CancelRequestedAt > 0 {
		return worker.persistThenAck(ctx, message.MessageID, run.ID, RunOutcome{
			Status: scheduledtaskmodel.RunStatusCanceled, Attempt: run.Attempt, FinishedAt: now,
			ErrorCode: "canceled", ErrorSummary: "cancellation requested during execution",
		})
	}
	if executeErr == nil {
		return worker.persistThenAck(ctx, message.MessageID, run.ID, RunOutcome{
			Status: scheduledtaskmodel.RunStatusSuccess, Attempt: run.Attempt,
			FinishedAt: now, ResultSummary: truncateSummary(result.Summary),
		})
	}

	executionError := classifyExecutionError(executeErr, timedOut)
	outcome := RunOutcome{
		Status: scheduledtaskmodel.RunStatusFailed, Attempt: run.Attempt,
		FinishedAt: now, ErrorCode: executionError.Code, ErrorSummary: truncateSummary(executionError.Summary),
	}
	if executionError.Temporary && run.Attempt < task.MaxRetries {
		outcome.Status = scheduledtaskmodel.RunStatusRetryWait
		outcome.Attempt = run.Attempt + 1
		outcome.FinishedAt = 0
		outcome.NextRetryAt = worker.config.Now().Add(worker.retryDelay(task)).UnixMilli()
	}
	return worker.persistThenAck(ctx, message.MessageID, run.ID, outcome)
}

func (worker *Worker) processBatch(ctx context.Context, messages []QueueMessage) error {
	if len(messages) == 0 {
		return nil
	}
	semaphore := make(chan struct{}, worker.config.WorkerCount)
	errCh := make(chan error, len(messages))
	var group sync.WaitGroup
	for _, message := range messages {
		message := message
		group.Add(1)
		go func() {
			defer group.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
			if err := worker.ProcessMessage(ctx, message); err != nil {
				errCh <- err
			}
		}()
	}
	group.Wait()
	close(errCh)
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (worker *Worker) persistThenAck(ctx context.Context, messageID, runID string, outcome RunOutcome) error {
	if err := worker.store.CompleteRun(ctx, runID, outcome); err != nil {
		return err
	}
	return worker.queue.Ack(ctx, messageID)
}

func (worker *Worker) runRunHeartbeat(ctx context.Context, runID string, stop <-chan struct{}, cancel context.CancelFunc) {
	ticker := time.NewTicker(worker.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case <-ticker.C:
			_ = worker.store.HeartbeatRun(ctx, runID, worker.config.WorkerID, worker.config.Now().UnixMilli())
			run, err := worker.store.GetRun(ctx, runID)
			if err == nil && run.CancelRequestedAt > 0 {
				cancel()
				return
			}
		}
	}
}

func (worker *Worker) runWorkerHeartbeat(ctx context.Context) {
	write := func() {
		_ = worker.queue.HeartbeatWorker(ctx, WorkerHeartbeat{
			WorkerID: worker.config.WorkerID, Role: worker.config.Role, Version: worker.config.Version,
			StartedAt: worker.startedAt, LastHeartbeat: worker.config.Now().UnixMilli(),
			CurrentRuns: int(worker.currentRuns.Load()), WorkerCount: worker.config.WorkerCount,
		}, worker.config.WorkerTTL)
	}
	write()
	ticker := time.NewTicker(worker.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			write()
		}
	}
}

func (worker *Worker) retryDelay(task scheduledtaskmodel.Task) time.Duration {
	var config struct {
		Type    string `json:"type"`
		Seconds int    `json:"seconds"`
	}
	if json.Unmarshal([]byte(task.RetryBackoffJSON), &config) == nil && config.Seconds > 0 {
		return time.Duration(config.Seconds) * time.Second
	}
	return worker.config.DefaultRetryDelay
}

func classifyExecutionError(err error, timedOut bool) *ExecutionError {
	if timedOut || errors.Is(err, context.DeadlineExceeded) {
		return &ExecutionError{Code: "timeout", Summary: "scheduled task execution timed out", Temporary: true}
	}
	var executionError *ExecutionError
	if errors.As(err, &executionError) {
		return executionError
	}
	return &ExecutionError{Code: "execution_failed", Summary: err.Error()}
}

func truncateSummary(value string) string {
	const maxBytes = 2000
	if len(value) <= maxBytes {
		return value
	}
	return value[:maxBytes]
}
