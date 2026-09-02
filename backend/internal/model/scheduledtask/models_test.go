package scheduledtaskmodel

import "testing"

func TestScheduledTaskModelsExposeRuntimeTablesAndStatuses(t *testing.T) {
	if got := (Task{}).TableName(); got != "scheduled_tasks" {
		t.Fatalf("Task.TableName() = %q", got)
	}
	if got := (Run{}).TableName(); got != "scheduled_task_runs" {
		t.Fatalf("Run.TableName() = %q", got)
	}
	if got := (RunLog{}).TableName(); got != "scheduled_task_run_logs" {
		t.Fatalf("RunLog.TableName() = %q", got)
	}
	for name, value := range map[string]string{
		"waiting":    RunStatusWaiting,
		"queued":     RunStatusQueued,
		"running":    RunStatusRunning,
		"retry_wait": RunStatusRetryWait,
		"success":    RunStatusSuccess,
		"failed":     RunStatusFailed,
		"canceled":   RunStatusCanceled,
		"skipped":    RunStatusSkipped,
	} {
		if value != name {
			t.Fatalf("run status %s = %q", name, value)
		}
	}
}
