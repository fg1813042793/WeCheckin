package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowSummaryPermissionMigrationRegistersWithoutAutoGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_workflow_summary_permissions.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow summary permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow summary permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:button:workflow:summary",
		"dingtalk_h5:menu:workflow",
		"dingtalk_h5:api:workflow:summary",
		"dingtalk_h5:api:workflow:export",
		"/api/v2/dingtalk/h5/workflows/summary/instances",
		"/api/v2/dingtalk/h5/workflows/summary/export",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow summary migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow summary migration must not auto-grant permissions")
	}
}
