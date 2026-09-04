package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminSwaggerMenuMigration(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_admin_swagger_menu.sql"))
	if err != nil {
		t.Fatalf("glob admin swagger menu migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one admin swagger menu migration, got %d", len(matches))
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read admin swagger menu migration: %v", err)
	}
	text := string(source)
	for _, want := range []string{
		"INSERT INTO `permissions`",
		"'admin:menu:swagger'",
		"'接口文档'",
		"'/swagger-docs'",
		"'Document'",
		"'swagger:view'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("admin swagger menu migration missing %q", want)
		}
	}
	if strings.Contains(text, "INSERT INTO `permission_grants`") {
		t.Fatal("admin swagger menu migration must not expand ordinary role grants")
	}
}
