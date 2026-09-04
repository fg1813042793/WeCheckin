package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInAppNotificationDeliveryMigrationRegistersSchemaAndPermissions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_in_app_notification_delivery.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("in-app notification delivery migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read in-app notification delivery migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"notify_delivery_key",
		"uk_notify_delivery_key",
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
		"admin:menu:notification",
		"admin:menu:notification:list",
		"admin:menu:notification:read",
		"admin:menu:notification:send",
		"admin:api-category:notification",
		"admin:api:notification:list",
		"admin:api:notification:read",
		"admin:api:notification:send",
		"/api/v2/admin/in-app-notifications",
		"permission_grants",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("in-app notification delivery migration missing %q", snippet)
		}
	}
}
