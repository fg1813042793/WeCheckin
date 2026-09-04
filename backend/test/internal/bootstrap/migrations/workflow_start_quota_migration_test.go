package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowStartQuotaMigrationCreatesUsageTableAndLookupIndex(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_create_workflow_start_quota_usage.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow start quota migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow start quota migration: %v", err)
	}
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `workflow_start_quota_usage`",
		"UNIQUE KEY `uk_workflow_start_quota_period` (`definition_id`,`starter_id`,`period_key`)",
		"idx_workflow_instances_definition_starter_time",
	} {
		if !strings.Contains(string(source), snippet) {
			t.Fatalf("workflow start quota migration must include %q", snippet)
		}
	}
}

func TestWorkflowStartQuotaTableIsNotAutoMigrated(t *testing.T) {
	source, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	if strings.Contains(string(source), "WorkflowStartQuotaUsage") {
		t.Fatal("workflow start quota table must be created by versioned SQL, not AutoMigrate")
	}
}
