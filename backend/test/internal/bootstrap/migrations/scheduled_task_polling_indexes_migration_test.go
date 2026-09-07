package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduledTaskPollingIndexesMigrationCoversRuntimeQueries(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_optimize_scheduled_task_polling_indexes.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("scheduled task polling indexes migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read scheduled task polling indexes migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.STATISTICS",
		"idx_scheduled_tasks_poll_due",
		"(`enabled`,`deleted_at`,`next_run_at`,`id`)",
		"idx_scheduled_runs_delivery_due",
		"(`run_status`,`redis_message_id`,`queued_at`,`id`)",
		"idx_scheduled_task_runs_recovery",
		"(`run_status`,`heartbeat_at`)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("scheduled task polling indexes migration missing %q", snippet)
		}
	}
}
