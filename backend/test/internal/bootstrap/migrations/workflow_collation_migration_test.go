package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowRuntimeCollationMigrationNormalizesRelatedTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_normalize_workflow_runtime_collations.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow runtime collation migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow runtime collation migration: %v", err)
	}
	text := string(source)
	for _, table := range []string{
		"workflow_instance_participants",
		"workflow_notification_outbox",
	} {
		statement := "ALTER TABLE `" + table + "` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
		if !strings.Contains(text, statement) {
			t.Fatalf("workflow runtime collation migration must include %q", statement)
		}
	}
}
