package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPermissionHiddenMigrationAddsMenuVisibilityColumn(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261029140000_add_permission_hidden.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read permission hidden migration: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"permission_hidden",
		"INFORMATION_SCHEMA",
		"ALTER TABLE `permissions`",
		"DEFAULT 0",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("permission hidden migration missing %q", required)
		}
	}
}
