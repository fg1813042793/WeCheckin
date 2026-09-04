package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceStarterDeleteGrantBackfillUsesWorkflowViewGrants(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_backfill_workflow_starter_delete_grants.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow starter delete grant backfill migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow starter delete grant backfill migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"FROM `permission_grants` source_grant",
		"'dingtalk_h5:api:workflow:view'",
		"'dingtalk_h5:api:workflow:delete'",
		"source_grant.`grant_subject_type` IN ('role', 'user')",
		"source_grant.`grant_effect` = 'allow'",
		"source_grant.`grant_status` = 1",
		"'workflow-view-delete-backfill'",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow starter delete grant backfill migration missing %q", snippet)
		}
	}
}
