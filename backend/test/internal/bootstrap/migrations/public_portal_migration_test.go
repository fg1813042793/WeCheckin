package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestPublicPortalMigrationDefaultsPrivateAndPreservesExistingConfiguration(t *testing.T) {
	raw, err := os.ReadFile("../../migrations/20261103090000_create_public_activity_portal.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, fragment := range []string{"CREATE TABLE IF NOT EXISTS `public_portal_settings`", "CREATE TABLE IF NOT EXISTS `public_portal_items`", "`enabled` TINYINT NOT NULL DEFAULT 0", "`version` BIGINT UNSIGNED NOT NULL DEFAULT 1", "UNIQUE KEY `uk_public_portal_source` (`kind`,`source_id`)", "ON DUPLICATE KEY UPDATE `id` = `id`", "'admin:menu:public-portal'", "'admin:menu:public-portal:edit'", "'admin:api:public-portal:list'", "'admin:api:public-portal:edit'", "'超级管理员'", "NOT EXISTS"} {
		if !strings.Contains(s, fragment) {
			t.Errorf("migration missing %q", fragment)
		}
	}
	if strings.Contains(s, "source_grant") {
		t.Fatal("new public publishing permissions must not be copied to ordinary roles")
	}
}
