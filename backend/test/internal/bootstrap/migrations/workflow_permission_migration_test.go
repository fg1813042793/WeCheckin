package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowPermissionMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_admin_workflow_permissions.sql"))
	if err != nil {
		t.Fatalf("glob workflow permission migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("workflow permission migration is required for initialized databases")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read workflow permission migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"admin:menu:workflow",
		"admin:menu:workflow:definitions",
		"admin:menu:workflow:publish",
		"admin:api-category:workflow",
		"admin:api:workflow:list",
		"admin:api:workflow:publish",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow permission migration must include %s", snippet)
		}
	}
}
