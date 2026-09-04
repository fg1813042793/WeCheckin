package config

import (
	"reflect"
	"strings"
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

func TestScheduledTaskConfigRejectsInvalidRuntimeRelationships(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantKey string
	}{
		{
			name:    "worker ttl must exceed heartbeat",
			yaml:    "scheduled_task:\n  worker_heartbeat_seconds: 10\n  worker_ttl_seconds: 10\n",
			wantKey: "scheduled_task.worker_ttl_seconds",
		},
		{
			name:    "recovery timeout must cover worker ttl",
			yaml:    "scheduled_task:\n  worker_ttl_seconds: 30\n  recovery_timeout_seconds: 29\n",
			wantKey: "scheduled_task.recovery_timeout_seconds",
		},
		{
			name:    "run log budget must cover a segment",
			yaml:    "scheduled_task:\n  max_log_segment_bytes: 2048\n  max_log_run_bytes: 1024\n",
			wantKey: "scheduled_task.max_log_run_bytes",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withTempConfigDir(t, test.yaml)
			_, err := LoadConfig("")
			if err == nil || !strings.Contains(err.Error(), test.wantKey) {
				t.Fatalf("LoadConfig() error = %v, want key %q", err, test.wantKey)
			}
		})
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
