package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserListIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_user_list_indexes.sql"))
	if err != nil {
		t.Fatalf("glob user list index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("user list index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read user list index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_users_add_time_id",
		"idx_users_mobile",
		"idx_users_name",
		"idx_user_depts_dept_user",
		"idx_user_depts_user_dept",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list index migration must include %s", snippet)
		}
	}
}
