package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceAdminDeleteMigrationAddsAuditColumnsAndPermission(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_admin_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance admin delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance admin delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"workflow_process_instances",
		"admin_deleted_at",
		"admin_deleted_by",
		"idx_workflow_instances_admin_deleted_time",
		"admin:menu:workflow:instance:delete",
		"admin:api:workflow:instance:delete",
		"workflow:instance:delete",
		"/api/v2/admin/workflow-instances/:id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance admin delete migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow instance admin delete migration must not auto-grant destructive permission")
	}
}
