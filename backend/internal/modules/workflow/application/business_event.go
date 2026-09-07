package application

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	BusinessEventStatusPending = "pending"
	BusinessEventStatusSending = "sending"
	BusinessEventStatusSent    = "sent"
	BusinessEventStatusFailed  = "failed"
	BusinessEventStatusDead    = "dead"

	businessEventSendingLease = 10 * time.Minute
	businessEventMaxAttempts  = 5
)

type BusinessEventRecord struct {
	ID       string
	Event    LifecycleEvent
	Attempts int
}

type BusinessEventRepository interface {
	ClaimDue(context.Context, int64, int64, int) ([]BusinessEventRecord, error)
	MarkSent(context.Context, string, int64) error
	MarkFailed(context.Context, string, int, string, int64, string, int64) error
}

type BusinessEventDispatcher interface {
	DispatchDueBusinessEvents(context.Context, int) (int, error)
}

type businessEventDispatcher struct {
	repository BusinessEventRepository
	events     LifecycleEventDispatcher
	now        func() time.Time
}

func NewBusinessEventDispatcher(repository BusinessEventRepository, events LifecycleEventDispatcher) BusinessEventDispatcher {
	return newBusinessEventDispatcherWithClock(repository, events, time.Now)
}

func newBusinessEventDispatcherWithClock(
	repository BusinessEventRepository,
	events LifecycleEventDispatcher,
	now func() time.Time,
) *businessEventDispatcher {
	if now == nil {
		now = time.Now
	}
	return &businessEventDispatcher{repository: repository, events: events, now: now}
}

func (dispatcher *businessEventDispatcher) DispatchDueBusinessEvents(ctx context.Context, limit int) (int, error) {
	if dispatcher == nil || dispatcher.repository == nil || dispatcher.events == nil {
		return 0, errors.New("工作流业务事件派发器未初始化")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	now := dispatcher.now()
	records, err := dispatcher.repository.ClaimDue(
		ctx, now.UnixMilli(), now.Add(-businessEventSendingLease).UnixMilli(), limit,
	)
	if err != nil {
		return 0, err
	}
	var firstErr error
	for _, record := range records {
		if err := dispatcher.events.Dispatch(ctx, record.Event); err != nil {
			attempts := record.Attempts + 1
			status := BusinessEventStatusFailed
			nextRetryAt := now.Add(businessEventRetryDelay(attempts)).UnixMilli()
			if attempts >= businessEventMaxAttempts {
				status = BusinessEventStatusDead
				nextRetryAt = 0
			}
			if markErr := dispatcher.repository.MarkFailed(
				ctx, record.ID, attempts, status, nextRetryAt, truncateBusinessEventError(err), now.UnixMilli(),
			); markErr != nil && firstErr == nil {
				firstErr = markErr
			}
			continue
		}
		if err := dispatcher.repository.MarkSent(ctx, record.ID, now.UnixMilli()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return len(records), firstErr
}

func businessEventRetryDelay(attempts int) time.Duration {
	delays := [...]time.Duration{30 * time.Second, 60 * time.Second, 120 * time.Second, 300 * time.Second, 600 * time.Second}
	if attempts <= 1 {
		return delays[0]
	}
	if attempts >= len(delays) {
		return delays[len(delays)-1]
	}
	return delays[attempts-1]
}

func truncateBusinessEventError(err error) string {
	if err == nil {
		return ""
	}
	value := strings.TrimSpace(err.Error())
	runes := []rune(value)
	if len(runes) > 1000 {
		value = string(runes[:1000])
	}
	return value
}
