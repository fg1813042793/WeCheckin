package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationOutboxMigrationCreatesTableAndSystemTask(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20260904113000_create_notification_outbox.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS `notification_outbox`",
		"UNIQUE KEY `uk_notification_outbox_idempotency` (`idempotency_key`)",
		"KEY `idx_notification_outbox_due` (`notification_status`,`next_retry_at`)",
		"KEY `idx_notification_outbox_status_edit` (`notification_status`,`edit_time`)",
		"'system.notification-outbox-dispatch'",
		`'{"handlerKey":"notification.outbox.dispatch_due","params":{"limit":100}}'`,
		"'* * * * *'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("notification outbox migration missing %q", required)
		}
	}
}
