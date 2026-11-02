package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPermissionCacheMigrationAddsMenuCacheColumn(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261029130000_add_permission_cache_enabled.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read permission cache migration: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"permission_cache_enabled",
		"INFORMATION_SCHEMA",
		"ALTER TABLE `permissions`",
		"DEFAULT 1",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("permission cache migration missing %q", required)
		}
	}
}
