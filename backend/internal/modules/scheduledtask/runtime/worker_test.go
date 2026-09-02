package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

func TestWorkerAcknowledgesMessageWhenRunCannotBeClaimed(t *testing.T) {
	store := &fakeWorkerStore{claimed: false}
	queue := &fakeWorkerQueue{}
	worker := NewWorker(store, queue, &fakeExecutor{}, WorkerConfig{WorkerID: "worker-a"})

	if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "1-0", RunID: "run-1"}); err != nil {
		t.Fatal(err)
	}
	if len(queue.acked) != 1 || queue.acked[0] != "1-0" {
		t.Fatalf("acked = %#v", queue.acked)
	}
}

func TestWorkerPersistsSuccessBeforeAcknowledging(t *testing.T) {
	store := &fakeWorkerStore{claimed: true, run: workerRun(0)}
	queue := &fakeWorkerQueue{store: store}
	worker := NewWorker(store, queue, &fakeExecutor{result: ExecutionResult{Summary: "done"}}, WorkerConfig{WorkerID: "worker-a"})

	if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "2-0", RunID: "run-2"}); err != nil {
		t.Fatal(err)
	}
	if store.outcome.Status != scheduledtaskmodel.RunStatusSuccess || store.outcome.ResultSummary != "done" {
		t.Fatalf("outcome = %#v", store.outcome)
	}
	if queue.ackedBeforePersist {
		t.Fatal("message acknowledged before terminal state persisted")
	}
}

func TestWorkerPersistsRetryWaitForTemporaryFailure(t *testing.T) {
	store := &fakeWorkerStore{claimed: true, run: workerRun(2)}
	queue := &fakeWorkerQueue{store: store}
	executor := &fakeExecutor{err: &ExecutionError{Code: "temporary", Summary: "try later", Temporary: true}}
	worker := NewWorker(store, queue, executor, WorkerConfig{WorkerID: "worker-a", DefaultRetryDelay: 30 * time.Second})

	if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "3-0", RunID: "run-3"}); err != nil {
		t.Fatal(err)
	}
	if store.outcome.Status != scheduledtaskmodel.RunStatusRetryWait || store.outcome.NextRetryAt == 0 || store.outcome.Attempt != 1 {
		t.Fatalf("retry outcome = %#v", store.outcome)
	}
}

func TestWorkerDoesNotAcknowledgeWhenDatabaseCompletionFails(t *testing.T) {
	store := &fakeWorkerStore{claimed: true, run: workerRun(0), completeErr: errors.New("database unavailable")}
	queue := &fakeWorkerQueue{store: store}
	worker := NewWorker(store, queue, &fakeExecutor{}, WorkerConfig{WorkerID: "worker-a"})

	if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "4-0", RunID: "run-4"}); err == nil {
		t.Fatal("ProcessMessage() must report database failure")
	}
	if len(queue.acked) != 0 {
		t.Fatalf("acked = %#v", queue.acked)
	}
}

func TestWorkerRunRecoversAfterTransientQueueFailure(t *testing.T) {
	store := &fakeWorkerStore{}
	queue := &recoveringWorkerQueue{readStarted: make(chan struct{}, 1)}
	worker := NewWorker(store, queue, &fakeExecutor{}, WorkerConfig{
		WorkerID: "worker-a", PollBlock: time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	select {
	case <-queue.readStarted:
		cancel()
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal("worker did not continue after a transient queue failure")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context cancellation", err)
	}
	if queue.autoClaimCalls.Load() < 2 {
		t.Fatalf("AutoClaim() calls = %d, want at least 2", queue.autoClaimCalls.Load())
	}
}

func TestWorkerRunContinuesAfterMessageDatabaseFailure(t *testing.T) {
	store := &claimFailingWorkerStore{}
	queue := &messageFailureWorkerQueue{loopContinued: make(chan struct{}, 1)}
	worker := NewWorker(store, queue, &fakeExecutor{}, WorkerConfig{
		WorkerID: "worker-a", PollBlock: time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	select {
	case <-queue.loopContinued:
		cancel()
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal("worker stopped after a transient database failure while processing a message")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context cancellation", err)
	}
	if len(queue.acked) != 0 {
		t.Fatalf("acked = %#v, failed message must remain pending", queue.acked)
	}
}

func TestWorkerRunRecreatesConsumerGroupAfterQueueFailure(t *testing.T) {
	store := &fakeWorkerStore{}
	queue := &groupRecoveringWorkerQueue{groupRechecked: make(chan struct{}, 1)}
	worker := NewWorker(store, queue, &fakeExecutor{}, WorkerConfig{
		WorkerID: "worker-a", PollBlock: time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	select {
	case <-queue.groupRechecked:
		cancel()
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal("worker did not recheck the consumer group after a queue failure")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context cancellation", err)
	}
}

func TestWorkerHonorsCancellationAndTimeout(t *testing.T) {
	t.Run("cancel requested", func(t *testing.T) {
		run := workerRun(0)
		run.CancelRequestedAt = 1
		store := &fakeWorkerStore{claimed: true, run: run}
		queue := &fakeWorkerQueue{store: store}
		executor := &fakeExecutor{}
		worker := NewWorker(store, queue, executor, WorkerConfig{WorkerID: "worker-a"})
		if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "5-0", RunID: run.ID}); err != nil {
			t.Fatal(err)
		}
		if store.outcome.Status != scheduledtaskmodel.RunStatusCanceled || executor.calls != 0 {
			t.Fatalf("outcome/calls = %#v / %d", store.outcome, executor.calls)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		run := workerRun(0)
		run.TaskSnapshotJSON = `{"timeoutSeconds":1,"maxRetries":0}`
		store := &fakeWorkerStore{claimed: true, run: run}
		queue := &fakeWorkerQueue{store: store}
		executor := &fakeExecutor{waitForContext: true}
		worker := NewWorker(store, queue, executor, WorkerConfig{WorkerID: "worker-a"})
		if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "6-0", RunID: run.ID}); err != nil {
			t.Fatal(err)
		}
		if store.outcome.Status != scheduledtaskmodel.RunStatusFailed || store.outcome.ErrorCode != "timeout" {
			t.Fatalf("timeout outcome = %#v", store.outcome)
		}
	})

	t.Run("cancel requested during execution", func(t *testing.T) {
		run := workerRun(0)
		run.TaskSnapshotJSON = `{"timeoutSeconds":1,"maxRetries":0}`
		store := &fakeWorkerStore{claimed: true, run: run, cancelOnSubsequentRead: true}
		queue := &fakeWorkerQueue{store: store}
		executor := &fakeExecutor{waitForContext: true}
		worker := NewWorker(store, queue, executor, WorkerConfig{
			WorkerID: "worker-a", HeartbeatInterval: time.Millisecond,
		})
		if err := worker.ProcessMessage(context.Background(), QueueMessage{MessageID: "7-0", RunID: run.ID}); err != nil {
			t.Fatal(err)
		}
		if store.outcome.Status != scheduledtaskmodel.RunStatusCanceled {
			t.Fatalf("cancel outcome = %#v", store.outcome)
		}
	})
}

func workerRun(maxRetries int) *scheduledtaskmodel.Run {
	return &scheduledtaskmodel.Run{
		ID: "run-2", Status: scheduledtaskmodel.RunStatusRunning,
		TaskSnapshotJSON: `{"timeoutSeconds":300,"maxRetries":` + string(rune('0'+maxRetries)) + `}`,
	}
}

type fakeWorkerStore struct {
	claimed                bool
	run                    *scheduledtaskmodel.Run
	outcome                RunOutcome
	persisted              bool
	completeErr            error
	cancelOnSubsequentRead bool
	reads                  atomic.Int32
}

func (store *fakeWorkerStore) TryClaimRun(context.Context, string, string, int64) (bool, error) {
	return store.claimed, nil
}
func (store *fakeWorkerStore) GetRun(context.Context, string) (*scheduledtaskmodel.Run, error) {
	if store.cancelOnSubsequentRead && store.reads.Add(1) > 1 {
		copy := *store.run
		copy.CancelRequestedAt = 1
		return &copy, nil
	}
	return store.run, nil
}
func (store *fakeWorkerStore) CompleteRun(_ context.Context, _ string, outcome RunOutcome) error {
	if store.completeErr != nil {
		return store.completeErr
	}
	store.outcome = outcome
	store.persisted = true
	return nil
}
func (store *fakeWorkerStore) HeartbeatRun(context.Context, string, string, int64) error { return nil }

type fakeWorkerQueue struct {
	store              *fakeWorkerStore
	acked              []string
	ackedBeforePersist bool
}

func (queue *fakeWorkerQueue) EnsureGroup(context.Context) error { return nil }

func (queue *fakeWorkerQueue) Ack(_ context.Context, ids ...string) error {
	if queue.store != nil && !queue.store.persisted {
		queue.ackedBeforePersist = true
	}
	queue.acked = append(queue.acked, ids...)
	return nil
}
func (queue *fakeWorkerQueue) Read(context.Context, string, int64, time.Duration) ([]QueueMessage, error) {
	return nil, nil
}
func (queue *fakeWorkerQueue) AutoClaim(context.Context, string, time.Duration, string, int64) ([]QueueMessage, string, error) {
	return nil, "0-0", nil
}
func (queue *fakeWorkerQueue) HeartbeatWorker(context.Context, WorkerHeartbeat, time.Duration) error {
	return nil
}

type groupRecoveringWorkerQueue struct {
	fakeWorkerQueue
	ensureCalls    atomic.Int32
	autoClaimCalls atomic.Int32
	groupRechecked chan struct{}
}

func (queue *groupRecoveringWorkerQueue) EnsureGroup(context.Context) error {
	if queue.ensureCalls.Add(1) >= 2 {
		select {
		case queue.groupRechecked <- struct{}{}:
		default:
		}
	}
	return nil
}

func (queue *groupRecoveringWorkerQueue) AutoClaim(context.Context, string, time.Duration, string, int64) ([]QueueMessage, string, error) {
	if queue.autoClaimCalls.Add(1) == 1 {
		return nil, "0-0", errors.New("NOGROUP consumer group is missing")
	}
	return nil, "0-0", nil
}

type recoveringWorkerQueue struct {
	fakeWorkerQueue
	autoClaimCalls atomic.Int32
	readStarted    chan struct{}
}

func (queue *recoveringWorkerQueue) AutoClaim(context.Context, string, time.Duration, string, int64) ([]QueueMessage, string, error) {
	if queue.autoClaimCalls.Add(1) == 1 {
		return nil, "0-0", errors.New("redis unavailable")
	}
	return nil, "0-0", nil
}

func (queue *recoveringWorkerQueue) Read(ctx context.Context, _ string, _ int64, _ time.Duration) ([]QueueMessage, error) {
	select {
	case queue.readStarted <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

type claimFailingWorkerStore struct{ fakeWorkerStore }

func (*claimFailingWorkerStore) TryClaimRun(context.Context, string, string, int64) (bool, error) {
	return false, errors.New("database timeout")
}

type messageFailureWorkerQueue struct {
	fakeWorkerQueue
	autoClaimCalls atomic.Int32
	readCalls      atomic.Int32
	loopContinued  chan struct{}
}

func (queue *messageFailureWorkerQueue) AutoClaim(context.Context, string, time.Duration, string, int64) ([]QueueMessage, string, error) {
	if queue.autoClaimCalls.Add(1) >= 2 {
		select {
		case queue.loopContinued <- struct{}{}:
		default:
		}
	}
	return nil, "0-0", nil
}

func (queue *messageFailureWorkerQueue) Read(ctx context.Context, _ string, _ int64, _ time.Duration) ([]QueueMessage, error) {
	if queue.readCalls.Add(1) == 1 {
		return []QueueMessage{{MessageID: "failed-message", RunID: "failed-run"}}, nil
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

type fakeExecutor struct {
	result         ExecutionResult
	err            error
	calls          int
	waitForContext bool
}

func (executor *fakeExecutor) Execute(ctx context.Context, _ *scheduledtaskmodel.Run) (ExecutionResult, error) {
	executor.calls++
	if executor.waitForContext {
		<-ctx.Done()
		return ExecutionResult{}, ctx.Err()
	}
	return executor.result, executor.err
}
