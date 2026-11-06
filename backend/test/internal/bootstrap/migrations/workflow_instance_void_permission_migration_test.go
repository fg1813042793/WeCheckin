package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceVoidPermissionMigrationRegistersMenuAndAPI(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_void_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance void permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance void permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"admin:menu:workflow:instance:void",
		"admin:api:workflow:instance:void",
		"workflow:instance:void",
		"/api/v2/admin/workflow-instances/:id/void",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance void permission migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow instance void permission migration must not auto-grant destructive permission")
	}
}
