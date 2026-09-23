package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowHistoryImportPermissionMigrationBackfillsActiveStartAllows(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_workflow_history_import_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow history import permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow history import permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:button:workflow:history-import",
		"导入历史表单",
		"dingtalk_h5:api:workflow:start",
		"source_grant.`grant_effect` = 'allow'",
		"source_grant.`grant_status` = 1",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow history import permission migration missing %q", snippet)
		}
	}
	if strings.Contains(text, "dingtalk_h5:api:workflow:history-import") {
		t.Fatal("history import must reuse existing workflow APIs instead of adding an import API permission")
	}
}
