package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEventRecordIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_event_record_indexes.sql"))
	if err != nil {
		t.Fatalf("glob event record index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("event record index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read event record index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_events_status_type_order_time",
		"idx_events_add_time_id",
		"idx_events_title",
		"idx_events_dept_create_time",
		"idx_event_parts_event_time",
		"idx_event_parts_openid_event",
		"idx_event_parts_event_openid",
		"idx_event_roles_user_event",
		"idx_event_roles_event_user",
		"idx_event_dynamics_event_time",
		"idx_event_scores_event_time",
		"idx_event_scores_event_participant",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("event record index migration must include %s", snippet)
		}
	}
}
