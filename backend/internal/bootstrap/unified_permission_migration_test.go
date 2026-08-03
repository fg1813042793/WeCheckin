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

func TestMissingClientFavoriteMenuPermissionPatchMigration(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_missing_client_favorite_permission.sql"))
	if err != nil {
		t.Fatalf("glob client favorite permission patch migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("client favorite permission patch migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read client favorite permission patch migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"'client:menu:favorite'",
		"'我的收藏'",
		"'client'",
		"'menu'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("client favorite permission patch migration must include %s", snippet)
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

func TestPermissionGrantLookupIndexMigrationIsRepeatable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_permission_grant_lookup_indexes.sql"))
	if err != nil {
		t.Fatalf("glob permission grant lookup index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("permission grant lookup index migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read permission grant lookup index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_permission_grants_subject_effect_status_key",
		"INFORMATION_SCHEMA.STATISTICS",
		"ALTER TABLE `permission_grants` ADD INDEX",
		"`grant_subject_type`, `grant_effect`, `grant_status`, `grant_subject_id`, `grant_permission_key`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission grant lookup index migration must include %s", snippet)
		}
	}
}

func TestPermissionGrantSubjectStatusIndexMigrationIsRepeatable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_permission_grant_subject_status_index.sql"))
	if err != nil {
		t.Fatalf("glob permission grant subject status index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("permission grant subject status index migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read permission grant subject status index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_permission_grants_subject_status_key",
		"INFORMATION_SCHEMA.STATISTICS",
		"ALTER TABLE `permission_grants` ADD INDEX",
		"`grant_subject_type`, `grant_subject_id`, `grant_status`, `grant_permission_key`, `grant_effect`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission grant subject status index migration must include %s", snippet)
		}
	}
}

func TestSetupKeyLookupIndexMigrationIsRepeatable(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_setup_lookup_indexes.sql"))
	if err != nil {
		t.Fatalf("glob setup lookup index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("setup lookup index migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read setup lookup index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_setups_setup_key",
		"INFORMATION_SCHEMA.STATISTICS",
		"TABLE_NAME = 'setups'",
		"COLUMN_NAME = 'setup_key'",
		"ALTER TABLE `setups` ADD INDEX",
		"`setup_key`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("setup lookup index migration must include %s", snippet)
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

func TestDingTalkH5OrgAPIPermissionBackfillMigrationUsesExistingGrants(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_backfill_dingtalk_h5_org_api_grants.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 org api permission backfill migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 org api permission backfill migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 org api permission backfill migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"FROM `permission_grants` source_grant",
		"'dingtalk_h5:menu:performance:org'",
		"'dingtalk_h5:button:user:config'",
		"'dingtalk_h5:api:user:list'",
		"'dingtalk_h5:api:user:edit'",
		"source_grant.`grant_subject_type` IN ('role', 'user')",
		"`grant_source`, `grant_status`",
		"'h5-menu-api-backfill'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 org api permission backfill migration must include %s", snippet)
		}
	}
}
