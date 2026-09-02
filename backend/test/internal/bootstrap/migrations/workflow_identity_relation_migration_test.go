package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowIdentityRelationMigrationBackfillsLegacyAssignments(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_refactor_workflow_identity_relations.sql"))
	if err != nil {
		t.Fatalf("glob workflow identity relation migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one workflow identity relation migration, got %d", len(matches))
	}
	content, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow identity relation migration: %v", err)
	}
	source := string(content)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS `user_reporting_relations`",
		"SELECT `id`, `manager_user_id`",
		"ALTER TABLE `users`",
		"DROP COLUMN `manager_user_id`",
		"ADD COLUMN `subject_type`",
		"ADD COLUMN `subject_id`",
		"SET `subject_type` = 'department'",
		"`subject_id` = `department_id`",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("workflow identity relation migration must include %q", want)
		}
	}
}
