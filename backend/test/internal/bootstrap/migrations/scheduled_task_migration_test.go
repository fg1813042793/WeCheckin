package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduledTaskMigrationCreatesRuntimeTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_create_scheduled_task_tables.sql"))
	if err != nil {
		t.Fatalf("glob scheduled task migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one scheduled task migration, got %d", len(matches))
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read scheduled task migration: %v", err)
	}
	text := string(source)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS `scheduled_tasks`",
		"CREATE TABLE IF NOT EXISTS `scheduled_task_runs`",
		"CREATE TABLE IF NOT EXISTS `scheduled_task_run_logs`",
		"UNIQUE KEY `uk_scheduled_task_code` (`code`)",
		"UNIQUE KEY `uk_scheduled_task_run_key` (`run_key`)",
		"KEY `idx_scheduled_tasks_due` (`enabled`,`next_run_at`)",
		"KEY `idx_scheduled_task_runs_recovery` (`run_status`,`heartbeat_at`)",
		"KEY `idx_scheduled_task_runs_task_time` (`task_id`,`scheduled_at`)",
		"KEY `idx_scheduled_task_run_logs_run_seq` (`run_id`,`log_sequence`)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("scheduled task migration missing %q", want)
		}
	}
}

func TestScheduledTaskPermissionMigrationRegistersHistoricalDefinitions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_scheduled_task_permissions.sql"))
	if err != nil {
		t.Fatalf("glob scheduled task permission migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one scheduled task permission migration, got %d", len(matches))
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read scheduled task permission migration: %v", err)
	}
	text := string(source)
	for _, want := range []string{
		"'admin:menu:scheduled-task'",
		"'admin:menu:scheduled-task:tasks'",
		"'admin:menu:scheduled-task:runs'",
		"'admin:menu:scheduled-task:workers'",
		"'admin:api-category:scheduled-task'",
		"'admin:api:scheduled-task:list'",
		"'admin:api:scheduled-task:run:retry'",
		"'admin:api:scheduled-task:sql:write'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("scheduled task permission migration missing %q", want)
		}
	}
}
