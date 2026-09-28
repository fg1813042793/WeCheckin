package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowRuntimeStartConfigMigrationAddsLiveConfigColumns(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_runtime_start_config.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow runtime start config migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow runtime start config migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{"definition_start_config_json", "definition_start_config_revision", "MEDIUMTEXT NULL", "INFORMATION_SCHEMA.COLUMNS"} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow runtime start config migration missing %q", snippet)
		}
	}
}
