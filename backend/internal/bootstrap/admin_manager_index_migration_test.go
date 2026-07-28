package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminManagerIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_admin_manager_indexes.sql"))
	if err != nil {
		t.Fatalf("glob admin manager index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("admin manager index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read admin manager index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_admins_add_time_id",
		"idx_admins_phone",
		"idx_admin_depts_dept_admin",
		"idx_admin_depts_admin_dept",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin manager index migration must include %s", snippet)
		}
	}
}
