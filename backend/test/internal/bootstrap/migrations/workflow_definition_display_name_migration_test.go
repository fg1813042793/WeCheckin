package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowDefinitionDisplayNameMigrationAddsDefinitionAndInstanceSnapshots(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_definition_display_name.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow definition display-name migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow definition display-name migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS", "PREPARE stmt FROM @ddl",
		"definition_display_name", "definition_name_snapshot",
		"UPDATE `workflow_process_instances` AS i", "workflow_definitions",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow definition display-name migration missing %q", snippet)
		}
	}
}
