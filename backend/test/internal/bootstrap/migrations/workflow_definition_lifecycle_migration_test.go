package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowDefinitionLifecycleMigrationAddsSoftDeleteColumnsAndIndex(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_definition_lifecycle_controls.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow definition lifecycle migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow definition lifecycle migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
		"workflow_definitions",
		"definition_deleted_at",
		"definition_deleted_by",
		"idx_workflow_definitions_active_status_edit",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow definition lifecycle migration missing %q", snippet)
		}
	}
}
