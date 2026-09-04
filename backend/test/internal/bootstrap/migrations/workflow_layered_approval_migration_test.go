package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowLayeredApprovalMigrationAddsTaskSnapshotAndSupervisorIdentity(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_layered_approval.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow layered approval migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow layered approval migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"approval_chain_key", "approval_layer", "approval_layer_total",
		"source_department_id", "source_department_name",
		"'supervisor', '主管'", "INFORMATION_SCHEMA.COLUMNS", "ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow layered approval migration missing %q", snippet)
		}
	}
}
