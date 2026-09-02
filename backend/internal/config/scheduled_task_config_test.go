package config

import (
	"reflect"
	"testing"
)

func TestScheduledTaskConfigUsesSecureDefaults(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ScheduledTask.EnableShell || cfg.ScheduledTask.EnableSQL {
		t.Fatalf("high-risk handlers must be disabled: %#v", cfg.ScheduledTask)
	}
	if cfg.ScheduledTask.WorkerCount != 4 || cfg.ScheduledTask.MinimumSecondInterval != 5 {
		t.Fatalf("scheduled task defaults = %#v", cfg.ScheduledTask)
	}
	if got := scheduledTaskIntField(t, cfg.ScheduledTask, "SchedulerRecoverySeconds"); got != 30 {
		t.Fatalf("scheduler recovery interval = %d, want 30", got)
	}
	if cfg.ScheduledTask.RedisKeyPrefix != "wecheckin" {
		t.Fatalf("redis key prefix = %q", cfg.ScheduledTask.RedisKeyPrefix)
	}
}

func TestScheduledTaskConfigAllowsEnvironmentOverrides(t *testing.T) {
	withTempWorkingDir(t)
	t.Setenv("WECHECKIN_SCHEDULED_TASK_WORKER_COUNT", "7")
	t.Setenv("WECHECKIN_SCHEDULED_TASK_SCHEDULER_POLL_SECONDS", "3")
	t.Setenv("WECHECKIN_SCHEDULED_TASK_SCHEDULER_RECOVERY_SECONDS", "45")
	t.Setenv("WECHECKIN_SCHEDULED_TASK_ENABLE_SHELL", "true")
	t.Setenv("WECHECKIN_SCHEDULED_TASK_ENABLE_SQL", "true")
	t.Setenv("WECHECKIN_SCHEDULED_TASK_REDIS_KEY_PREFIX", "test-env")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ScheduledTask.WorkerCount != 7 || cfg.ScheduledTask.SchedulerPollSeconds != 3 {
		t.Fatalf("scheduled task runtime config = %#v", cfg.ScheduledTask)
	}
	if got := scheduledTaskIntField(t, cfg.ScheduledTask, "SchedulerRecoverySeconds"); got != 45 {
		t.Fatalf("scheduler recovery interval = %d, want 45", got)
	}
	if !cfg.ScheduledTask.EnableShell || !cfg.ScheduledTask.EnableSQL {
		t.Fatalf("scheduled task handler switches = %#v", cfg.ScheduledTask)
	}
	if cfg.ScheduledTask.RedisKeyPrefix != "test-env" {
		t.Fatalf("redis key prefix = %q", cfg.ScheduledTask.RedisKeyPrefix)
	}
}

func scheduledTaskIntField(t *testing.T, cfg ScheduledTaskConfig, name string) int {
	t.Helper()
	field := reflect.ValueOf(cfg).FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("ScheduledTaskConfig missing %s", name)
	}
	return int(field.Int())
}
