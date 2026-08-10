package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleAdminLoginMigrationAddsControlAndBackfills(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_role_admin_login_control.sql"))
	if err != nil {
		t.Fatalf("glob role admin login migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("role admin login migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read role admin login migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"role_allow_admin_login",
		"DEFAULT 1",
		"UPDATE `roles` SET `role_allow_admin_login` = 1",
		"超级管理员",
		"`user_admin_type` = 1",
		"`user_role_id` = @super_admin_role_id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role admin login migration must include %s", snippet)
		}
	}
}
