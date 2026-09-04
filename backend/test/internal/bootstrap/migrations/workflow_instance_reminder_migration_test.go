package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceReminderMigrationRegistersPermissionWithoutAutoGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_reminder_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance reminder migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance reminder migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:api:workflow:remind",
		"/api/v2/dingtalk/h5/workflows/instances/:id/reminders",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance reminder migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow instance reminder migration must not auto-grant reminder permission")
	}
}
