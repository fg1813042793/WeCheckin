package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerTickPublishesGeneratedAndRecoverableRuns(t *testing.T) {
	store := &fakeSchedulerStore{
		generated: []string{"run-new"}, undelivered: []string{"run-old"},
		retryDue: []string{"run-retry"}, waitingReady: []string{"run-waiting"}, staleRecovered: []string{"run-stale"},
	}
	publisher := &fakeSchedulerPublisher{}
	scheduler := NewScheduler(store, publisher, SchedulerConfig{BatchSize: 20, DispatchRecoveryAge: time.Minute})

	if err := scheduler.Tick(context.Background(), time.UnixMilli(100000)); err != nil {
		t.Fatal(err)
	}
	want := "[run-new run-old run-retry run-waiting run-stale]"
	if fmt.Sprint(publisher.runIDs) != want || fmt.Sprint(store.dispatched) != want {
		t.Fatalf("published/marked = %v / %v, want %s", publisher.runIDs, store.dispatched, want)
	}
}

func TestSchedulerDoesNotMarkRunDispatchedWhenRedisPublishFails(t *testing.T) {
	store := &fakeSchedulerStore{generated: []string{"run-new"}}
	publisher := &fakeSchedulerPublisher{err: errors.New("redis unavailable")}
	scheduler := NewScheduler(store, publisher, SchedulerConfig{})

	if err := scheduler.Tick(context.Background(), time.Now()); err == nil {
		t.Fatal("Tick() must report publish failure")
	}
	if len(store.dispatched) != 0 {
		t.Fatalf("marked dispatched = %#v", store.dispatched)
	}
}

func TestSchedulerRunRecoversAfterTransientStoreFailure(t *testing.T) {
	store := &recoveringSchedulerStore{published: make(chan struct{}, 1)}
	publisher := &fakeSchedulerPublisher{}
	scheduler := NewScheduler(store, publisher, SchedulerConfig{PollInterval: time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- scheduler.Run(ctx) }()

	select {
	case <-store.published:
		cancel()
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal("scheduler did not retry after a transient store failure")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context cancellation", err)
	}
	if store.generateCalls.Load() < 2 {
		t.Fatalf("GenerateDueRuns() calls = %d, want at least 2", store.generateCalls.Load())
	}
}

func TestSchedulerTickContinuesOtherCollectionsAfterOneFails(t *testing.T) {
	store := &fakeSchedulerStore{
		generateErr: errors.New("database timeout"),
		undelivered: []string{"run-recoverable"},
	}
	publisher := &fakeSchedulerPublisher{}
	scheduler := NewScheduler(store, publisher, SchedulerConfig{})

	if err := scheduler.Tick(context.Background(), time.Now()); err == nil {
		t.Fatal("Tick() must report the failed collection")
	}
	if got := fmt.Sprint(publisher.runIDs); got != "[run-recoverable]" {
		t.Fatalf("published runs = %s, want [run-recoverable]", got)
	}
}

type fakeSchedulerStore struct {
	generated      []string
	generateErr    error
	undelivered    []string
	retryDue       []string
	waitingReady   []string
	staleRecovered []string
	dispatched     []string
}

func (store *fakeSchedulerStore) GenerateDueRuns(context.Context, int64, int) ([]string, error) {
	return store.generated, store.generateErr
}

type recoveringSchedulerStore struct {
	fakeSchedulerStore
	generateCalls atomic.Int32
	published     chan struct{}
}

func (store *recoveringSchedulerStore) GenerateDueRuns(context.Context, int64, int) ([]string, error) {
	if store.generateCalls.Add(1) == 1 {
		return nil, errors.New("database timeout")
	}
	select {
	case store.published <- struct{}{}:
	default:
	}
	return nil, nil
}
func (store *fakeSchedulerStore) ListUndeliveredRuns(context.Context, int64, int) ([]string, error) {
	return store.undelivered, nil
}
func (store *fakeSchedulerStore) WakeRetryRuns(context.Context, int64, int) ([]string, error) {
	return store.retryDue, nil
}
func (store *fakeSchedulerStore) WakeWaitingRuns(context.Context, int64, int) ([]string, error) {
	return store.waitingReady, nil
}
func (store *fakeSchedulerStore) RecoverStaleRuns(context.Context, int64, int) ([]string, error) {
	return store.staleRecovered, nil
}
func (store *fakeSchedulerStore) MarkRunDispatched(_ context.Context, runID, _ string, _ int64) error {
	store.dispatched = append(store.dispatched, runID)
	return nil
}

type fakeSchedulerPublisher struct {
	runIDs []string
	err    error
}

func (publisher *fakeSchedulerPublisher) PublishRun(_ context.Context, runID string) (string, error) {
	publisher.runIDs = append(publisher.runIDs, runID)
	if publisher.err != nil {
		return "", publisher.err
	}
	return "message-" + runID, nil
}
