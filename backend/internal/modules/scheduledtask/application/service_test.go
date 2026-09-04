package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

func TestCreateTaskValidatesHandlerAndComputesNextRun(t *testing.T) {
	store := &fakeStore{}
	validator := &fakeValidator{}
	service := newTestService(store, validator, &fakePublisher{})

	task, err := service.CreateTask(context.Background(), 7, CreateTaskRequest{
		Code:              "workflow-monthly",
		Name:              "Monthly workflow",
		HandlerType:       scheduledtaskmodel.HandlerTypeWorkflow,
		HandlerConfigJSON: json.RawMessage(`{"definitionId":9}`),
		CronPrecision:     scheduledtaskmodel.CronPrecisionMinute,
		CronExpression:    "0 9 1 * *",
		Timezone:          "Asia/Shanghai",
		Enabled:           true,
		MisfirePolicy:     scheduledtaskmodel.MisfirePolicyFireOnce,
		ConcurrencyPolicy: scheduledtaskmodel.ConcurrencyPolicySkip,
	})
	if err != nil {
		t.Fatal(err)
	}
	if task.NextRunAt == 0 || store.createdTask != task {
		t.Fatalf("created task = %#v", task)
	}
	if task.CreatedBy != 7 || task.UpdatedBy != 7 || task.Version != 1 {
		t.Fatalf("task audit/version = %#v", task)
	}
	if validator.handlerType != scheduledtaskmodel.HandlerTypeWorkflow || string(validator.raw) != `{"definitionId":9}` {
		t.Fatalf("handler validation = %q %s", validator.handlerType, validator.raw)
	}
}

func TestCreateTaskRejectsInvalidHandlerConfigBeforePersistence(t *testing.T) {
	store := &fakeStore{}
	validator := &fakeValidator{err: errors.New("definition is required")}
	service := newTestService(store, validator, &fakePublisher{})

	_, err := service.CreateTask(context.Background(), 7, CreateTaskRequest{
		Code: "bad-task", Name: "Bad task", HandlerType: "workflow", HandlerConfigJSON: json.RawMessage(`{}`),
		CronPrecision: "minute", CronExpression: "0 * * * *", Timezone: "UTC",
	})
	if err == nil || store.createdTask != nil {
		t.Fatalf("error = %v, created task = %#v", err, store.createdTask)
	}
}

func TestCreateTaskRejectsReservedSystemCode(t *testing.T) {
	store := &fakeStore{}
	service := newTestService(store, &fakeValidator{}, nil)
	_, err := service.CreateTask(context.Background(), 7, CreateTaskRequest{Code: "system.custom", Name: "System task"})
	if !errors.Is(err, ErrSystemTaskReadOnly) || store.createdTask != nil {
		t.Fatalf("error=%v task=%#v", err, store.createdTask)
	}
}

func TestUpdateSystemTaskIsReadOnly(t *testing.T) {
	task := validTask()
	task.Code = "system.notification-outbox-dispatch"
	task.Version = 1
	store := &fakeStore{task: task}
	service := newTestService(store, &fakeValidator{}, nil)
	_, err := service.UpdateTask(context.Background(), task.ID, 7, UpdateTaskRequest{Version: 1})
	if !errors.Is(err, ErrSystemTaskReadOnly) {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateScheduledRunAppliesConcurrencyPolicies(t *testing.T) {
	now := time.Date(2026, time.September, 1, 1, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		policy      string
		active      *scheduledtaskmodel.Run
		waiting     *scheduledtaskmodel.Run
		wantStatus  string
		wantCreated bool
		wantMerged  bool
	}{
		{name: "allow creates queued", policy: "allow", wantStatus: "queued", wantCreated: true},
		{name: "skip records skipped", policy: "skip", active: &scheduledtaskmodel.Run{ID: "active"}, wantStatus: "skipped", wantCreated: true},
		{name: "queue once creates waiting", policy: "queue_once", active: &scheduledtaskmodel.Run{ID: "active"}, wantStatus: "waiting", wantCreated: true},
		{name: "queue once merges waiting", policy: "queue_once", active: &scheduledtaskmodel.Run{ID: "active"}, waiting: &scheduledtaskmodel.Run{ID: "waiting", CoalescedCount: 2}, wantMerged: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{activeRun: test.active, waitingRun: test.waiting}
			service := newTestService(store, &fakeValidator{}, &fakePublisher{})
			task := &scheduledtaskmodel.Task{ID: 9, Code: "task", HandlerType: "go", HandlerConfigJSON: `{}`, ConcurrencyPolicy: test.policy}

			result, err := service.CreateScheduledRun(context.Background(), task, now, 1, scheduledtaskmodel.TriggerTypeScheduled)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantMerged {
				if result.Run.ID != "waiting" || !result.Merged || store.coalescedRunID != "waiting" {
					t.Fatalf("merged result = %#v, store = %#v", result, store)
				}
				return
			}
			if (store.createdRun != nil) != test.wantCreated || result.Run.Status != test.wantStatus {
				t.Fatalf("created/result = %#v / %#v", store.createdRun, result)
			}
		})
	}
}

func TestRunNowKeepsRunWhenQueuePublishFails(t *testing.T) {
	store := &fakeStore{task: validTask()}
	publisher := &fakePublisher{err: errors.New("redis unavailable")}
	service := newTestService(store, &fakeValidator{}, publisher)

	result, err := service.RunNow(context.Background(), 9, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Run == nil || store.createdRun != result.Run || !result.DispatchPending {
		t.Fatalf("run result = %#v", result)
	}
	if result.Run.TriggerType != scheduledtaskmodel.TriggerTypeManual || result.Run.Status != scheduledtaskmodel.RunStatusQueued {
		t.Fatalf("run = %#v", result.Run)
	}
}

func TestRetryFailedRunCreatesLinkedManualRetry(t *testing.T) {
	task := validTask()
	failed := &scheduledtaskmodel.Run{ID: "run-old", TaskID: task.ID, Status: scheduledtaskmodel.RunStatusFailed}
	store := &fakeStore{task: task, run: failed}
	service := newTestService(store, &fakeValidator{}, &fakePublisher{messageID: "1-0"})

	result, err := service.RetryRun(context.Background(), failed.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Run.ParentRunID != failed.ID || result.Run.TriggerType != scheduledtaskmodel.TriggerTypeManualRetry {
		t.Fatalf("retry run = %#v", result.Run)
	}
	if result.DispatchPending || store.dispatchedMessageID != "1-0" {
		t.Fatalf("dispatch result = %#v, message = %q", result, store.dispatchedMessageID)
	}
}

func newTestService(store Store, validator HandlerConfigValidator, publisher QueuePublisher) *Service {
	return NewService(store, validator, publisher, ServiceConfig{
		MinimumSecondInterval: 5 * time.Second,
		Now: func() time.Time {
			return time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
		},
		NewRunID: func() string { return "run-new" },
	})
}

func validTask() *scheduledtaskmodel.Task {
	return &scheduledtaskmodel.Task{
		ID: 9, Code: "task", Name: "Task", HandlerType: "go", HandlerConfigJSON: `{}`,
		CronPrecision: "minute", CronExpression: "0 * * * *", Timezone: "UTC",
		MisfirePolicy: "skip", ConcurrencyPolicy: "allow", TimeoutSeconds: 300,
	}
}

type fakeStore struct {
	createdTask         *scheduledtaskmodel.Task
	createdRun          *scheduledtaskmodel.Run
	task                *scheduledtaskmodel.Task
	run                 *scheduledtaskmodel.Run
	activeRun           *scheduledtaskmodel.Run
	waitingRun          *scheduledtaskmodel.Run
	coalescedRunID      string
	dispatchedMessageID string
}

func (store *fakeStore) CreateTask(_ context.Context, task *scheduledtaskmodel.Task) error {
	store.createdTask = task
	if task.ID == 0 {
		task.ID = 9
	}
	return nil
}

func (store *fakeStore) UpdateTask(context.Context, *scheduledtaskmodel.Task, int64) error {
	return nil
}
func (store *fakeStore) GetTask(_ context.Context, _ uint64) (*scheduledtaskmodel.Task, error) {
	if store.task == nil {
		return nil, ErrTaskNotFound
	}
	return store.task, nil
}
func (store *fakeStore) CreateRun(_ context.Context, run *scheduledtaskmodel.Run) error {
	store.createdRun = run
	return nil
}
func (store *fakeStore) GetRun(_ context.Context, _ string) (*scheduledtaskmodel.Run, error) {
	if store.run == nil {
		return nil, ErrRunNotFound
	}
	return store.run, nil
}
func (store *fakeStore) FindActiveRun(context.Context, uint64) (*scheduledtaskmodel.Run, error) {
	return store.activeRun, nil
}
func (store *fakeStore) FindWaitingRun(context.Context, uint64) (*scheduledtaskmodel.Run, error) {
	return store.waitingRun, nil
}
func (store *fakeStore) IncrementRunCoalesced(_ context.Context, runID string, count int, _ int64) error {
	store.coalescedRunID = runID
	store.waitingRun.CoalescedCount += count
	return nil
}
func (store *fakeStore) MarkRunDispatched(_ context.Context, _ string, messageID string, _ int64) error {
	store.dispatchedMessageID = messageID
	return nil
}

type fakeValidator struct {
	handlerType string
	raw         json.RawMessage
	err         error
}

func (validator *fakeValidator) ValidateConfig(_ context.Context, handlerType string, raw json.RawMessage) error {
	validator.handlerType = handlerType
	validator.raw = append(json.RawMessage(nil), raw...)
	return validator.err
}

type fakePublisher struct {
	messageID string
	err       error
}

func (publisher *fakePublisher) PublishRun(context.Context, string) (string, error) {
	return publisher.messageID, publisher.err
}
