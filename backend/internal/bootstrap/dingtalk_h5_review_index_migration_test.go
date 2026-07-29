package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5ReviewIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_review_indexes.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 review index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 review index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 review index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_dt_h5_review_employee_period",
		"idx_dt_h5_review_manager_status",
		"idx_dt_h5_review_hrbp_status",
		"idx_dt_h5_review_hrbp_reviewer_status",
		"idx_dt_h5_review_status_period",
		"idx_dt_h5_history_review_time",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 review index migration must include %s", snippet)
		}
	}
}
