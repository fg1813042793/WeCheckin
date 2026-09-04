package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPermissionIconMigrationIsIdempotentAndBackfillsLegacyIcons(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_permission_icon.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("permission icon migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read permission icon migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"COLUMN_NAME = 'permission_icon'",
		"ALTER TABLE `permissions` ADD COLUMN `permission_icon`",
		"m.`menu_icon`",
		"PREPARE stmt FROM @ddl",
		"DEALLOCATE PREPARE stmt",
		"PREPARE permission_icon_backfill_stmt",
		"DEALLOCATE PREPARE permission_icon_backfill_stmt",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission icon migration missing %q", snippet)
		}
	}
}
