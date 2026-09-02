package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type SchedulerStore interface {
	GenerateDueRuns(context.Context, int64, int) ([]string, error)
	ListUndeliveredRuns(context.Context, int64, int) ([]string, error)
	WakeRetryRuns(context.Context, int64, int) ([]string, error)
	WakeWaitingRuns(context.Context, int64, int) ([]string, error)
	RecoverStaleRuns(context.Context, int64, int) ([]string, error)
	MarkRunDispatched(context.Context, string, string, int64) error
}

type RunPublisher interface {
	PublishRun(context.Context, string) (string, error)
}

type SchedulerConfig struct {
	PollInterval        time.Duration
	RecoveryInterval    time.Duration
	BatchSize           int
	DispatchRecoveryAge time.Duration
	StaleRunAge         time.Duration
	RetryInitialDelay   time.Duration
	RetryMaxDelay       time.Duration
	Logf                logFunc
	Now                 func() time.Time
}

type Scheduler struct {
	store     SchedulerStore
	publisher RunPublisher
	config    SchedulerConfig
}

func NewScheduler(store SchedulerStore, publisher RunPublisher, cfg SchedulerConfig) *Scheduler {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.RecoveryInterval <= 0 {
		cfg.RecoveryInterval = 30 * time.Second
	}
	if cfg.DispatchRecoveryAge <= 0 {
		cfg.DispatchRecoveryAge = 30 * time.Second
	}
	if cfg.StaleRunAge <= 0 {
		cfg.StaleRunAge = 90 * time.Second
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Scheduler{store: store, publisher: publisher, config: cfg}
}

func (scheduler *Scheduler) Run(ctx context.Context) error {
	if scheduler == nil || scheduler.store == nil || scheduler.publisher == nil {
		return errors.New("scheduled task scheduler is not initialized")
	}
	loops := []struct {
		name     string
		interval time.Duration
		tick     func(context.Context, time.Time) error
	}{
		{name: "scheduler", interval: scheduler.config.PollInterval, tick: scheduler.tickDue},
		{name: "scheduler-recovery", interval: scheduler.config.RecoveryInterval, tick: scheduler.tickRecovery},
	}
	var group sync.WaitGroup
	errorsCh := make(chan error, len(loops))
	for _, loop := range loops {
		loop := loop
		group.Add(1)
		go func() {
			defer group.Done()
			errorsCh <- scheduler.runLoop(ctx, loop.name, loop.interval, loop.tick)
		}()
	}
	group.Wait()
	close(errorsCh)
	var errs []error
	for err := range errorsCh {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (scheduler *Scheduler) Tick(ctx context.Context, now time.Time) error {
	if scheduler == nil || scheduler.store == nil || scheduler.publisher == nil {
		return errors.New("scheduled task scheduler is not initialized")
	}
	now = now.UTC()
	return scheduler.tickCollections(ctx, now, []runCollection{
		{name: "generate due runs", collect: func() ([]string, error) {
			return scheduler.store.GenerateDueRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "list undelivered runs", collect: func() ([]string, error) {
			return scheduler.store.ListUndeliveredRuns(ctx, now.Add(-scheduler.config.DispatchRecoveryAge).UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "wake retry runs", collect: func() ([]string, error) {
			return scheduler.store.WakeRetryRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "wake waiting runs", collect: func() ([]string, error) {
			return scheduler.store.WakeWaitingRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "recover stale runs", collect: func() ([]string, error) {
			return scheduler.store.RecoverStaleRuns(ctx, now.Add(-scheduler.config.StaleRunAge).UnixMilli(), scheduler.config.BatchSize)
		}},
	})
}

type runCollection struct {
	name    string
	collect func() ([]string, error)
}

func (scheduler *Scheduler) tickDue(ctx context.Context, now time.Time) error {
	now = now.UTC()
	return scheduler.tickCollections(ctx, now, []runCollection{
		{name: "generate due runs", collect: func() ([]string, error) {
			return scheduler.store.GenerateDueRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "wake retry runs", collect: func() ([]string, error) {
			return scheduler.store.WakeRetryRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "wake waiting runs", collect: func() ([]string, error) {
			return scheduler.store.WakeWaitingRuns(ctx, now.UnixMilli(), scheduler.config.BatchSize)
		}},
	})
}

func (scheduler *Scheduler) tickRecovery(ctx context.Context, now time.Time) error {
	now = now.UTC()
	return scheduler.tickCollections(ctx, now, []runCollection{
		{name: "list undelivered runs", collect: func() ([]string, error) {
			return scheduler.store.ListUndeliveredRuns(ctx, now.Add(-scheduler.config.DispatchRecoveryAge).UnixMilli(), scheduler.config.BatchSize)
		}},
		{name: "recover stale runs", collect: func() ([]string, error) {
			return scheduler.store.RecoverStaleRuns(ctx, now.Add(-scheduler.config.StaleRunAge).UnixMilli(), scheduler.config.BatchSize)
		}},
	})
}

func (scheduler *Scheduler) tickCollections(ctx context.Context, now time.Time, collections []runCollection) error {
	runIDs := make([]string, 0)
	seen := make(map[string]struct{})
	var collectionErrors []error
	for _, collection := range collections {
		ids, err := collection.collect()
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("%s: %w", collection.name, err))
			continue
		}
		for _, id := range ids {
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			runIDs = append(runIDs, id)
		}
	}

	dispatchErrors := collectionErrors
	for _, runID := range runIDs {
		messageID, err := scheduler.publisher.PublishRun(ctx, runID)
		if err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Errorf("publish run %s: %w", runID, err))
			continue
		}
		if err := scheduler.store.MarkRunDispatched(ctx, runID, messageID, now.UnixMilli()); err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Errorf("mark run %s dispatched: %w", runID, err))
		}
	}
	return errors.Join(dispatchErrors...)
}

func (scheduler *Scheduler) runLoop(
	ctx context.Context,
	name string,
	interval time.Duration,
	tick func(context.Context, time.Time) error,
) error {
	retry := newRetryController(
		configuredRetryInitial(scheduler.config.RetryInitialDelay, interval),
		scheduler.config.RetryMaxDelay,
		scheduler.config.Logf,
	)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := tick(ctx, scheduler.config.Now()); err != nil {
			if waitErr := retry.WaitAfterFailure(ctx, name, "tick", err); waitErr != nil {
				return waitErr
			}
			continue
		}
		retry.MarkSuccess(name)
		if err := waitContext(ctx, interval); err != nil {
			return err
		}
	}
}

func configuredRetryInitial(configured, interval time.Duration) time.Duration {
	if configured > 0 {
		return configured
	}
	return retryInitialDelay(interval)
}
