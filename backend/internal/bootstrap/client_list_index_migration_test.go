package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientListQueryIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_client_list_query_indexes.sql"))
	if err != nil {
		t.Fatalf("glob client list query index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("client list query index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read client list query index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_enrolls_status_order_time",
		"idx_enrolls_title",
		"idx_events_status_type_order_time",
		"idx_surveys_status_order_id",
		"idx_exams_status_order_id",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("client list query index migration must include %s", snippet)
		}
	}
}
