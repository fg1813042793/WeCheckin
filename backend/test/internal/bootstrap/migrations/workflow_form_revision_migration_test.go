package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowFormRevisionMigrationAddsVersionAndPermissions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_form_revision.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow form revision migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow form revision migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"form_revision", "INFORMATION_SCHEMA.COLUMNS",
		"dingtalk_h5:button:workflow:form-revise",
		"dingtalk_h5:api:workflow:form-revise",
		"/api/v2/dingtalk/h5/workflows/instances/:id/form-data",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow form revision migration missing %q", snippet)
		}
	}
}
