package application

import (
	"context"
	"errors"
	"testing"
	"time"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

func TestSetTaskEnabledUsesOptimisticVersionAndRecomputesSchedule(t *testing.T) {
	task := validTask()
	task.Version = 4
	task.Enabled = 0
	store := &fakeManagementStore{fakeStore: fakeStore{task: task}}
	service := newTestService(store, &fakeValidator{}, nil)

	updated, err := service.SetTaskEnabled(context.Background(), task.ID, 7, true, 4)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Enabled != 1 || updated.Version != 5 || updated.NextRunAt == 0 {
		t.Fatalf("updated task = %#v", updated)
	}
	if store.updatedExpectedVersion != 4 || store.updatedTask != updated {
		t.Fatalf("store update = %#v, version = %d", store.updatedTask, store.updatedExpectedVersion)
	}
}

func TestSystemTaskCannotBeDisabledOrDeleted(t *testing.T) {
	for _, operation := range []string{"disable", "delete"} {
		t.Run(operation, func(t *testing.T) {
			task := validTask()
			task.Code = "system.notification-outbox-dispatch"
			task.Version = 2
			store := &fakeManagementStore{fakeStore: fakeStore{task: task}}
			service := newTestService(store, &fakeValidator{}, nil)
			var err error
			if operation == "disable" {
				_, err = service.SetTaskEnabled(context.Background(), task.ID, 7, false, task.Version)
			} else {
				err = service.DeleteTask(context.Background(), task.ID, 7)
			}
			if !errors.Is(err, ErrSystemTaskReadOnly) {
				t.Fatalf("error = %v", err)
			}
			if store.updatedTask != nil || store.deletedTaskID != 0 {
				t.Fatalf("system task was mutated: %#v", store)
			}
		})
	}
}

func TestSystemTaskCanRunNow(t *testing.T) {
	task := validTask()
	task.Code = "system.notification-outbox-dispatch"
	store := &fakeManagementStore{fakeStore: fakeStore{task: task}}
	service := newTestService(store, &fakeValidator{}, &fakePublisher{})
	if _, err := service.RunNow(context.Background(), task.ID, 7); err != nil {
		t.Fatal(err)
	}
	if store.createdRun == nil {
		t.Fatal("manual run was not created")
	}
}

func TestCancelRunDistinguishesQueuedAndRunning(t *testing.T) {
	for _, test := range []struct {
		status     string
		wantStatus string
		wantSignal bool
	}{
		{status: scheduledtaskmodel.RunStatusQueued, wantStatus: scheduledtaskmodel.RunStatusCanceled},
		{status: scheduledtaskmodel.RunStatusRunning, wantStatus: scheduledtaskmodel.RunStatusRunning, wantSignal: true},
	} {
		t.Run(test.status, func(t *testing.T) {
			run := &scheduledtaskmodel.Run{ID: "run-1", Status: test.status}
			store := &fakeManagementStore{fakeStore: fakeStore{run: run}}
			service := newTestService(store, &fakeValidator{}, nil)

			updated, err := service.CancelRun(context.Background(), run.ID, 7)
			if err != nil {
				t.Fatal(err)
			}
			if updated.Status != test.wantStatus || updated.CancelRequestedAt == 0 {
				t.Fatalf("canceled run = %#v", updated)
			}
			if store.cancelRunID != run.ID || store.cancelRunning != test.wantSignal {
				t.Fatalf("cancel store call = %q/%v", store.cancelRunID, store.cancelRunning)
			}
		})
	}
}

func TestPreviewCronReturnsLocalOccurrences(t *testing.T) {
	service := newTestService(&fakeManagementStore{}, &fakeValidator{}, nil)
	result, err := service.PreviewCron(CronPreviewRequest{
		Expression: "0 9 * * *", Precision: "minute", Timezone: "Asia/Shanghai", Count: 2,
		After: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Occurrences) != 2 || result.Occurrences[0].LocalTime != "2026-09-01 09:00:00 Asia/Shanghai" {
		t.Fatalf("preview = %#v", result)
	}
}

type fakeManagementStore struct {
	fakeStore
	updatedTask            *scheduledtaskmodel.Task
	updatedExpectedVersion int64
	cancelRunID            string
	cancelRunning          bool
	deletedTaskID          uint64
}

func (store *fakeManagementStore) UpdateTask(_ context.Context, task *scheduledtaskmodel.Task, version int64) error {
	store.updatedTask = task
	store.updatedExpectedVersion = version
	return nil
}

func (store *fakeManagementStore) ListTasks(context.Context, TaskQuery) ([]scheduledtaskmodel.Task, int64, error) {
	return nil, 0, nil
}

func (store *fakeManagementStore) DeleteTask(_ context.Context, taskID uint64, _ uint64, _ int64) error {
	store.deletedTaskID = taskID
	return nil
}

func (store *fakeManagementStore) ListRuns(context.Context, RunQuery) ([]scheduledtaskmodel.Run, int64, error) {
	return nil, 0, nil
}

func (store *fakeManagementStore) ListRunLogs(context.Context, string) ([]scheduledtaskmodel.RunLog, error) {
	return nil, nil
}

func (store *fakeManagementStore) CancelRun(_ context.Context, runID string, running bool, _ uint64, now int64) error {
	store.cancelRunID = runID
	store.cancelRunning = running
	store.run.CancelRequestedAt = now
	if !running {
		store.run.Status = scheduledtaskmodel.RunStatusCanceled
		store.run.FinishedAt = now
	}
	return nil
}
