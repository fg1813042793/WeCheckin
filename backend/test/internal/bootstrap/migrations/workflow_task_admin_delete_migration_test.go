package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowTaskAdminDeleteMigrationAddsAuditColumnsAndPermission(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_task_admin_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow task admin delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow task admin delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
		"PREPARE stmt FROM @ddl",
		"workflow_process_tasks",
		"admin_deleted_at",
		"admin_deleted_by",
		"idx_workflow_tasks_admin_deleted_time",
		"admin:menu:workflow:task:delete",
		"admin:api:workflow:task:delete",
		"workflow:task:delete",
		"/api/v2/admin/workflow-tasks/:id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow task admin delete migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow task admin delete migration must not auto-grant destructive permission")
	}
}
