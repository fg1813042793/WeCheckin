package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowDefinitionVersionHistoryMigrationAddsImmutableSnapshotFields(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_definition_version_history.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow definition version history migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow definition version history migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"definition_metadata_json",
		"definition_change_base_version",
		"definition_change_summary_json",
		"definition_publish_note",
		"definition_content_hash",
		"definition_rollback_from_version",
		"INFORMATION_SCHEMA.COLUMNS",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow definition version history migration missing %q", snippet)
		}
	}
}
