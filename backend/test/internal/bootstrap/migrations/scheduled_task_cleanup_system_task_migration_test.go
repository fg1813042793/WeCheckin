package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduledTaskCleanupSystemTaskMigration(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261009124000_create_scheduled_task_cleanup_system_task.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"system.scheduled-task-cleanup",
		"scheduled-task.cleanup",
		"0 3 * * *",
		"Asia/Shanghai",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("scheduled task cleanup migration missing %q", required)
		}
	}
}
