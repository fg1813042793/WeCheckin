package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowRuntimePermissionMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_admin_workflow_runtime_permissions.sql"))
	if err != nil {
		t.Fatalf("glob workflow runtime permission migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("workflow runtime permission migration is required for initialized databases")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read workflow runtime permission migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"admin:menu:workflow:instances",
		"admin:menu:workflow:tasks",
		"admin:menu:workflow:instance:start",
		"admin:menu:workflow:task:complete",
		"admin:api:workflow:instance:list",
		"admin:api:workflow:instance:start",
		"admin:api:workflow:instance:detail",
		"admin:api:workflow:task:list",
		"admin:api:workflow:task:complete",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow runtime permission migration must include %s", snippet)
		}
	}
}
