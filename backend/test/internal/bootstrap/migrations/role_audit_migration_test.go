package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleAuditMigrationAddsUnifiedDataScopeFields(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_role_audit_fields.sql"))
	if err != nil {
		t.Fatalf("glob role audit migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("role audit migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read role audit migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"ALTER TABLE `roles` ADD COLUMN `create_by`",
		"ALTER TABLE `roles` ADD COLUMN `update_by`",
		"ALTER TABLE `roles` ADD COLUMN `create_dept_id`",
		"ALTER TABLE `roles` ADD COLUMN `update_dept_id`",
		"idx_roles_unified_audit_scope",
		"UPDATE `roles` r",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role audit migration must include %s", snippet)
		}
	}
}
