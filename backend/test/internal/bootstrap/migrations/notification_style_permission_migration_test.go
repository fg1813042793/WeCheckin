package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationStylePermissionMigrationRegistersAndBackfillsPermissions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_notification_style_permissions.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("notification style permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read notification style permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"admin:menu:notification:style:list",
		"admin:menu:notification:style:edit",
		"admin:api:notification:style:list",
		"admin:api:notification:style:edit",
		"notification:style:list",
		"notification:style:edit",
		"/api/v2/admin/notification-styles",
		"notification-style-permission-backfill",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("notification style permission migration missing %q", snippet)
		}
	}
}
