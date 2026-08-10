package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPositionMigrationCreatesTableAndUserPositionColumn(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_positions_table_and_user_position.sql"))
	if err != nil {
		t.Fatalf("glob position migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("position migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read position migration: %v", err)
	}
	text := string(src)
	required := []string{
		"CREATE TABLE IF NOT EXISTS `positions`",
		"`position_name`",
		"`position_status`",
		"`position_sort`",
		"`user_position_id`",
		"idx_users_position_id",
		"INFORMATION_SCHEMA.COLUMNS",
		"INFORMATION_SCHEMA.STATISTICS",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("position migration must include %s", snippet)
		}
	}
}
