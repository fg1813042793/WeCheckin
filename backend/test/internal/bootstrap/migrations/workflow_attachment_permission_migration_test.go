package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowAttachmentPermissionMigrationBackfillsWriters(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_attachment_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow attachment permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow attachment permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:api:workflow:attachment",
		"/api/v2/dingtalk/h5/workflows/attachments",
		"dingtalk_h5:api:workflow:start",
		"dingtalk_h5:api:workflow:handle",
		"source_grant.`grant_subject_type` IN ('role', 'user')",
		"source_grant.`grant_effect` = 'allow'",
		"source_grant.`grant_status` = 1",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow attachment permission migration missing %q", snippet)
		}
	}
}
