package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowTableMigrationExistsForInitializedDatabases(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_create_workflow_definition_tables.sql"))
	if err != nil {
		t.Fatalf("glob workflow table migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("workflow table migration is required for initialized databases")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read workflow table migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `workflow_definitions`",
		"UNIQUE KEY `idx_workflow_definitions_definition_key` (`definition_key`)",
		"CREATE TABLE IF NOT EXISTS `workflow_definition_versions`",
		"UNIQUE KEY `idx_workflow_definition_version` (`definition_id`,`definition_version`)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow table migration must include %q", snippet)
		}
	}
}
