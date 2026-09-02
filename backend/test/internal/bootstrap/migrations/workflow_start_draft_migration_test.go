package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowStartDraftMigrationCreatesOwnedDraftTable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_create_workflow_start_drafts.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow start draft migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow start draft migration: %v", err)
	}
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `workflow_start_drafts`",
		"`form_data_json` mediumtext",
		"UNIQUE KEY `uk_workflow_start_draft_owner` (`definition_id`,`starter_id`)",
	} {
		if !strings.Contains(string(source), snippet) {
			t.Fatalf("workflow start draft migration must include %q", snippet)
		}
	}
}
