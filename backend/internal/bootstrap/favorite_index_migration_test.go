package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFavoriteIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_favorite_indexes.sql"))
	if err != nil {
		t.Fatalf("glob favorite index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("favorite index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read favorite index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_favorites_user_time",
		"idx_favorites_user_oid",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("favorite index migration must include %s", snippet)
		}
	}
}
