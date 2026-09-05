package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationSoftDeleteMigrationAddsColumnAndLookupIndex(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_notify_soft_delete.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("notification soft-delete migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read notification soft-delete migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"notify_deleted_at",
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
		"idx_notify_user_deleted_read_id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("notification soft-delete migration missing %q", snippet)
		}
	}
}
