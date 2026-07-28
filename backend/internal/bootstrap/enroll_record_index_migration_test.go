package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnrollRecordIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_enroll_record_indexes.sql"))
	if err != nil {
		t.Fatalf("glob enroll record index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("enroll record index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read enroll record index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_enroll_users_enroll_time",
		"idx_enroll_users_enroll_openid",
		"idx_enroll_users_openid_time",
		"idx_enroll_users_enroll_rank",
		"idx_enroll_joins_enroll_time",
		"idx_enroll_joins_enroll_user_day",
		"idx_enroll_joins_enroll_day_user",
		"idx_enroll_joins_user_day_time",
		"idx_enroll_joins_user_time",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("enroll record index migration must include %s", snippet)
		}
	}
}
