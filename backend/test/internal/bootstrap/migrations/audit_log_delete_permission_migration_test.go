package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLogDeletePermissionsRequireExplicitNewGrants(t *testing.T) {
	// TestMain runs this package from backend/internal/bootstrap.
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "migrations", "20261106110000_add_audit_log_delete_permissions.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, key := range []string{"admin:menu:log:delete", "admin:api:log:delete", "admin:menu:log:security-delete", "admin:api:log:security-delete", "admin:api:log:del", "ON DUPLICATE KEY UPDATE"} {
		if !strings.Contains(sql, key) {
			t.Fatalf("migration missing %s", key)
		}
	}
	if strings.Contains(strings.ToLower(sql), "insert into `permission_grants`") {
		t.Fatal("new deletion permissions must not be auto-granted")
	}
}
