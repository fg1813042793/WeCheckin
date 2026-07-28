package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentLogIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_content_log_indexes.sql"))
	if err != nil {
		t.Fatalf("glob content log index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("content log index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read content log index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_news_status_order_time",
		"idx_news_status_vouch_order_time",
		"idx_news_add_time_id",
		"idx_news_title",
		"idx_news_dept_create_time",
		"idx_logs_add_time_id",
		"idx_logs_admin_time",
		"idx_logs_admin_name",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("content log index migration must include %s", snippet)
		}
	}
}
