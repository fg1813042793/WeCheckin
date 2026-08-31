package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowRuntimeMigrationCreatesIndependentEngineTables(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_create_go_workflow_runtime_tables.sql"))
	if err != nil {
		t.Fatalf("glob workflow runtime migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("workflow runtime migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read workflow runtime migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `workflow_process_instances`",
		"UNIQUE KEY `idx_workflow_instance_business` (`definition_id`,`business_type`,`business_key`)",
		"CREATE TABLE IF NOT EXISTS `workflow_process_tokens`",
		"CREATE TABLE IF NOT EXISTS `workflow_process_tasks`",
		"KEY `idx_workflow_tasks_assignee_status` (`task_assignee_id`,`task_status`)",
		"CREATE TABLE IF NOT EXISTS `workflow_process_variables`",
		"UNIQUE KEY `idx_workflow_variable_instance_key` (`instance_id`,`variable_key`)",
		"CREATE TABLE IF NOT EXISTS `workflow_process_history`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow runtime migration must include %q", snippet)
		}
	}
}
