package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserRolesMigrationCreatesMultiRoleBindingTable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_user_roles.sql"))
	if err != nil {
		t.Fatalf("glob user roles migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("user roles migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read user roles migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `user_roles`",
		"`user_role_user_id` bigint unsigned NOT NULL",
		"`user_role_role_id` bigint unsigned NOT NULL",
		"`user_role_is_primary` tinyint NOT NULL DEFAULT 0",
		"UNIQUE KEY `uk_user_roles_user_role` (`user_role_user_id`, `user_role_role_id`)",
		"INSERT INTO `user_roles`",
		"SELECT `id`, `user_role_id`, 1, 1, 'legacy'",
		"`user_role_id` > 0",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user roles migration must include %s", snippet)
		}
	}
}

func TestAutoMigrateIncludesUserRoleModel(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	if !strings.Contains(string(src), "&model.UserRole{}") {
		t.Fatalf("auto migrate must include user role model")
	}
}
