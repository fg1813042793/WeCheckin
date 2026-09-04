package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackendQueryIndexesMigrationCoversGrowingTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_backend_query_indexes.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("backend query indexes migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read backend query indexes migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.STATISTICS",
		"idx_notify_user_read_id",
		"(`notify_user_id`,`notify_is_read`,`notify_id`)",
		"idx_notify_source_delivery",
		"(`notify_source_type`,`notify_source_id`,`notify_delivery_key`)",
		"idx_workflow_instances_business_lookup",
		"(`business_type`,`business_key`)",
		"idx_workflow_tasks_handled_status_instance",
		"(`handled_by`,`task_status`,`instance_id`)",
		"idx_workflow_tasks_assignee_status_created",
		"(`task_assignee_id`,`task_status`,`admin_deleted_at`,`created_at`,`id`)",
		"idx_workflow_notification_status_edit",
		"(`notification_status`,`edit_time`)",
		"idx_workflow_notification_time",
		"(`add_time`,`id`)",
		"idx_scheduled_runs_task_status_time",
		"(`task_id`,`run_status`,`add_time`,`id`)",
		"idx_scheduled_runs_delivery_due",
		"(`run_status`,`redis_message_id`,`queued_at`,`id`)",
		"idx_scheduled_runs_status_schedule",
		"(`run_status`,`scheduled_at`,`id`)",
		"idx_scheduled_runs_cleanup",
		"(`run_status`,`finished_at`,`id`)",
		"idx_scheduled_run_logs_cleanup",
		"(`log_time`,`id`)",
		"idx_workflow_definitions_published_sort",
		"(`definition_status`,`definition_category`,`definition_name`,`id`)",
		"idx_dt_h5_bindings_user_enabled",
		"(`user_id`,`enabled`,`corp_id`,`dingtalk_user_id`)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("backend query indexes migration missing %q", snippet)
		}
	}
}
