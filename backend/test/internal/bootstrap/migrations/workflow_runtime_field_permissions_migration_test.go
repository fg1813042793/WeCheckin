package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowRuntimeFieldPermissionsMigration(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_runtime_field_permissions.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{"definition_field_permissions_json", "definition_field_permissions_revision", "MEDIUMTEXT NULL", "INFORMATION_SCHEMA.COLUMNS"} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("migration missing %q", snippet)
		}
	}
}
