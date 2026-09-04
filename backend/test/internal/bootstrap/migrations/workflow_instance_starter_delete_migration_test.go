package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceStarterDeleteMigrationAddsSoftDeleteAndPermission(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_starter_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance starter delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance starter delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"workflow_process_instances",
		"starter_deleted_at",
		"idx_workflow_instances_starter_deleted_time",
		"dingtalk_h5:api:workflow:delete",
		"/api/v2/dingtalk/h5/workflows/instances/:id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance starter delete migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow instance starter delete migration must not auto-grant delete permission")
	}
}
