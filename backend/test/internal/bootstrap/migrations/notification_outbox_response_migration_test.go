package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationOutboxResponseMigrationAddsDeliveryReceiptColumns(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261027190000_add_notification_outbox_response_receipt.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"`response_status_code` INT NOT NULL DEFAULT 0",
		"`response_body` TEXT NOT NULL",
		"PREPARE",
		"DEALLOCATE PREPARE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("notification outbox response migration missing %q", required)
		}
	}
}
