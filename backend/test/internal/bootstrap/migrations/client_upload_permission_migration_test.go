package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientUploadPermissionMigrationRegistersAndBackfillsPermission(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20260908100000_add_client_upload_permission.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read client upload permission migration: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"client:api:upload:create",
		"/api/v2/uploads",
		"client-upload-permission-backfill",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("client upload permission migration missing %q", required)
		}
	}
}
