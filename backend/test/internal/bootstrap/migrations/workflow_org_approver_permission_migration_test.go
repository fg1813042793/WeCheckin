package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowOrgApproverPermissionMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_org_approver_admin_permissions.sql"))
	if err != nil {
		t.Fatalf("glob workflow org approver permission migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("workflow org approver admin permission migration is required for initialized databases")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read workflow org approver permission migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"INSERT INTO `permissions`",
		"admin:menu:workflow:org-approvers",
		"admin:menu:workflow:org-approver:list",
		"admin:menu:workflow:org-approver:edit",
		"admin:api:workflow:org-approver:list",
		"admin:api:workflow:org-approver:edit",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow org approver permission migration must include %s", snippet)
		}
	}
	if strings.Contains(text, "ensureMissing") || strings.Contains(text, "SyncAdmin") {
		t.Fatalf("workflow org approver permissions must be SQL migration data, not runtime backfill references")
	}
}
