package migrations_test

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

func TestDingTalkH5ReviewScopeIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_review_scope_indexes.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 review scope index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 review scope index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 review scope index migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"idx_dt_h5_review_del_status_period_id",
		"idx_dt_h5_review_del_employee_period_id",
		"idx_dt_h5_review_del_manager_period_id",
		"idx_dt_h5_review_del_hrbp_period_id",
		"idx_dt_h5_review_del_hrbp_reviewer_period_id",
		"idx_dt_h5_review_del_hrbp_status_period",
		"idx_dt_h5_review_del_hrbp_reviewer_status",
		"idx_dt_h5_review_del_dept1_status_period",
		"idx_dt_h5_review_del_dept2_status_period",
		"idx_dt_h5_review_del_dept3_status_period",
		"idx_dt_h5_review_del_create_by_status",
		"idx_dt_h5_review_del_create_dept_status",
		"`deleted_at`, `status`, `period`, `id`",
		"`deleted_at`, `employee_account`, `period`, `id`",
		"`deleted_at`, `manager_account`, `period`, `id`",
		"`deleted_at`, `hrbp_account`, `period`, `id`",
		"`deleted_at`, `hrbp_reviewer_account`, `period`, `id`",
		"`deleted_at`, `hrbp_account`, `status`, `period`, `id`",
		"`deleted_at`, `hrbp_reviewer_account`, `status`, `period`, `id`",
		"`deleted_at`, `create_by`, `status`, `period`, `id`",
		"`deleted_at`, `create_dept_id`, `status`, `period`, `id`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 review scope index migration must include %s", snippet)
		}
	}
}
