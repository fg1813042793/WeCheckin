package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationAdminDeleteMigrationAddsAuditFieldsAndPermissions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_notify_admin_soft_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("notification admin delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read notification admin delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"notify_admin_deleted_at",
		"notify_admin_deleted_by",
		"idx_notify_admin_deleted_id",
		"admin:menu:notification:delete",
		"admin:api:notification:delete",
		"notification:delete",
		"/api/v2/admin/in-app-notifications/:id",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("notification admin delete migration missing %q", snippet)
		}
	}
	if strings.Contains(text, "permission_grants") {
		t.Fatal("notification admin delete migration must not auto-grant destructive permission")
	}
}
