package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnrollClientLookupIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_enroll_client_lookup_indexes.sql"))
	if err != nil {
		t.Fatalf("glob enroll client lookup index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("enroll client lookup index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read enroll client lookup index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_enroll_joins_user_enroll",
		"idx_enroll_users_openid_enroll",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("enroll client lookup index migration must include %s", snippet)
		}
	}
}
