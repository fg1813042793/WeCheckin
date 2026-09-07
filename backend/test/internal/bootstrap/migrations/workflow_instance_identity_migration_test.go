package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceIdentityMigrationAddsSnapshotsAndIndexes(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_identity.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance identity migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance identity migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS", "INFORMATION_SCHEMA.STATISTICS", "PREPARE stmt FROM @ddl",
		"instance_title", "business_period_type", "business_period_key", "business_period_label",
		"idx_workflow_instances_title", "idx_workflow_instances_definition_period_time",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance identity migration missing %q", snippet)
		}
	}
}
