package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowNotificationMigrationCreatesRuntimeTablesAndPermissions(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_notifications.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow notification migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow notification migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `workflow_instance_participants`",
		"UNIQUE KEY `uk_workflow_participant_source` (`instance_id`,`user_id`,`participant_role`,`node_id`)",
		"CREATE TABLE IF NOT EXISTS `workflow_notification_outbox`",
		"UNIQUE KEY `uk_workflow_notification_dedupe` (`dedupe_key`)",
		"KEY `idx_workflow_notification_due` (`notification_status`,`next_retry_at`)",
		"workflow:notification:list",
		"workflow:notification:retry",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow notification migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow notification migration must not auto-grant permissions")
	}
}
