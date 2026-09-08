package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowNotificationAdminDeleteMigrationAddsAuditColumnsAndPermission(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_notification_admin_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow notification admin delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow notification admin delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
		"workflow_notification_outbox",
		"admin_deleted_at",
		"admin_deleted_by",
		"idx_workflow_notification_admin_deleted_time",
		"admin:menu:workflow:notification:delete",
		"admin:api:workflow:notification:delete",
		"workflow:notification:delete",
		"/api/v2/admin/workflow-notifications/:id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow notification admin delete migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow notification delete permission must not be granted automatically")
	}
}
