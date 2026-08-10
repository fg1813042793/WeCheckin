package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyAdminCleanupMigrationDropsOnlyObsoleteAdminTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_drop_legacy_admin_tables.sql"))
	if err != nil {
		t.Fatalf("glob legacy admin cleanup migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("legacy admin cleanup migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read legacy admin cleanup migration: %v", err)
	}
	text := string(src)
	required := []string{
		"INSERT INTO `user_depts`",
		"FROM `admin_depts`",
		"DROP TABLE IF EXISTS `admin_depts`",
		"DROP TABLE IF EXISTS `admins`",
		"DROP TABLE IF EXISTS `admin_user_merge_maps`",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("legacy admin cleanup migration must include %s", snippet)
		}
	}

	forbidden := []string{
		"DROP TABLE IF EXISTS `roles`",
		"DROP TABLE IF EXISTS `menus`",
		"DROP TABLE IF EXISTS `logs`",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("legacy admin cleanup migration must not remove active table %s", snippet)
		}
	}

	insertIndex := strings.Index(text, "INSERT INTO `user_depts`")
	dropIndex := strings.Index(text, "DROP TABLE IF EXISTS `admin_depts`")
	if insertIndex < 0 || dropIndex < 0 || dropIndex < insertIndex {
		t.Fatalf("legacy admin cleanup migration must merge admin_depts into user_depts before dropping admin_depts")
	}
}

func TestAutoMigrateRunsLegacyAdminCleanupStep(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`Name: "cleanup_legacy_admin_tables"`,
		"cleanupLegacyAdminTables(db)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("auto migrate must run legacy admin table cleanup step %s", snippet)
		}
	}

	mergeIndex := strings.Index(text, `Name: "merge_admins_into_users"`)
	cleanupIndex := strings.Index(text, `Name: "cleanup_legacy_admin_tables"`)
	if mergeIndex < 0 || cleanupIndex < 0 || cleanupIndex < mergeIndex {
		t.Fatalf("legacy admin cleanup must run after admins have been merged into users")
	}
}
