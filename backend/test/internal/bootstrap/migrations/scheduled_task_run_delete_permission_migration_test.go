package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScheduledTaskRunDeletePermissionMigration(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261009123000_add_scheduled_task_run_delete_permission.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"admin:menu:scheduled-task:run:delete",
		"admin:api:scheduled-task:run:delete",
		"scheduled-task:run:delete",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("scheduled task run delete migration missing %q", required)
		}
	}
}
