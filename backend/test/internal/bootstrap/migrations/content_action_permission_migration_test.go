package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentActionPermissionMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_split_admin_content_action_permissions.sql"))
	if err != nil {
		t.Fatalf("glob content action permission migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("content action permission migration is required for initialized databases")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read content action permission migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"admin:api:enroll:status",
		"admin:api:enroll:vouch",
		"admin:api:enroll:export",
		"admin:api:enroll:users",
		"admin:api:news:status",
		"admin:api:news:vouch",
		"admin:api:event:status",
		"admin:api:event:vouch",
		"admin:api:event:top",
		"admin:api:event:users",
		"admin:api:upload:create",
		"permission_grants",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("content action permission migration must include %s", snippet)
		}
	}
}
