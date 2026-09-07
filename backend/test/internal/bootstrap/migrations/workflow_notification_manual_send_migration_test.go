package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowNotificationManualSendMigrationSnapshotsBusinessKey(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_notification_manual_send.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow notification manual send migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow notification manual send migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"ADD COLUMN `business_key` VARCHAR(160) NOT NULL DEFAULT ''",
		"UPDATE `workflow_notification_outbox` AS n",
		"JOIN `workflow_process_instances` AS i",
		"SET n.`business_key` = i.`business_key`",
		"idx_workflow_notification_business_key",
		"admin:api:workflow:notification:send",
		"/api/v2/admin/workflow-notifications/:id/send",
		"workflow:notification:retry",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow notification manual send migration missing %q", snippet)
		}
	}
}
