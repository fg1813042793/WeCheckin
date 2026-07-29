package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnifiedPermissionMigrationCreatesAndBackfillsTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_unified_permissions.sql"))
	if err != nil {
		t.Fatalf("glob unified permission migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("unified permission migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read unified permission migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `permissions`",
		"CREATE TABLE IF NOT EXISTS `permission_grants`",
		"'admin:login'",
		"'data:all'",
		"'data:dept'",
		"'data:self'",
		"'data:custom'",
		"'admin:api:user:list'",
		"'client:menu:home'",
		"'dingtalk_h5:menu:dashboard'",
		"`permission_icon`",
		"m.`menu_icon`",
		"@legacy_menus_exists",
		"@legacy_role_menus_exists",
		"@legacy_role_depts_exists",
		"`INFORMATION_SCHEMA`.`TABLES`",
		"JSON_OBJECT(''deptIds''",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("unified permission migration must include %s", snippet)
		}
	}
}

func TestAutoMigrateIncludesUnifiedPermissionModelsAndSeedStep(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"&model.Permission{}",
		"&model.PermissionGrant{}",
		`Name: "ensure_unified_permissions"`,
		"ensureUnifiedPermissions(db, enableExam)",
		`Name: "cleanup_legacy_role_authorization_tables"`,
		"cleanupLegacyRoleAuthorizationTables(db)",
		`Name: "cleanup_legacy_menu_table"`,
		"cleanupLegacyMenuTable(db)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("auto migrate must include unified permission snippet %s", snippet)
		}
	}
	ensureIndex := strings.Index(text, `Name: "ensure_unified_permissions"`)
	cleanupIndex := strings.Index(text, `Name: "cleanup_legacy_role_authorization_tables"`)
	if ensureIndex < 0 || cleanupIndex < 0 || cleanupIndex < ensureIndex {
		t.Fatalf("legacy role authorization cleanup must run after unified permissions are ensured")
	}
	for _, snippet := range []string{
		"&model.RoleMenu{}",
		"&model.RoleDept{}",
		"&model.Menu{}",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("auto migrate must not recreate old authorization table model %s", snippet)
		}
	}
}

func TestLegacyMenuCleanupMigrationDropsMenusTable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_drop_legacy_menu_table.sql"))
	if err != nil {
		t.Fatalf("glob legacy menu cleanup migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("legacy menus cleanup migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read legacy menu cleanup migration: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "DROP TABLE IF EXISTS `menus`") {
		t.Fatalf("legacy menu cleanup migration must drop menus")
	}
	for _, forbidden := range []string{
		"DROP TABLE IF EXISTS `permissions`",
		"DROP TABLE IF EXISTS `permission_grants`",
		"DROP TABLE IF EXISTS `roles`",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("legacy menu cleanup migration must not drop active table %s", forbidden)
		}
	}
}

func TestLegacyRoleAuthorizationCleanupMigrationDropsOldGrantTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_drop_legacy_role_authorization_tables.sql"))
	if err != nil {
		t.Fatalf("glob legacy role authorization cleanup migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("legacy role authorization cleanup migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read legacy role authorization cleanup migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"DROP TABLE IF EXISTS `role_menus`",
		"DROP TABLE IF EXISTS `role_depts`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("legacy role authorization cleanup migration must include %s", snippet)
		}
	}
	for _, snippet := range []string{
		"DROP TABLE IF EXISTS `roles`",
		"DROP TABLE IF EXISTS `menus`",
		"DROP TABLE IF EXISTS `permissions`",
		"DROP TABLE IF EXISTS `permission_grants`",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("legacy role authorization cleanup migration must not remove active table %s", snippet)
		}
	}
}

func TestRoleAPIPermissionBackfillMigrationUsesExistingMenuGrants(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_backfill_role_api_grants_from_menu_perms.sql"))
	if err != nil {
		t.Fatalf("glob role api permission backfill migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("role api permission backfill migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read role api permission backfill migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"FROM `permission_grants` menu_grant",
		"JOIN `permissions` menu_perm",
		"JOIN `permissions` api_perm",
		"FIND_IN_SET(api_perm.`permission_perms`, REPLACE(menu_perm.`permission_perms`, ' ', ''))",
		"api_perm.`permission_key`",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role api permission backfill migration must include %s", snippet)
		}
	}
}
