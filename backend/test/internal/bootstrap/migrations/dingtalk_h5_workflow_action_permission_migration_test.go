package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5WorkflowActionPermissionMigration(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_split_dingtalk_h5_workflow_action_permissions.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow action permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow action permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:button:workflow:approve",
		"dingtalk_h5:button:workflow:reject",
		"dingtalk_h5:button:workflow:return",
		"dingtalk_h5:button:workflow:submit",
		"dingtalk_h5:api:workflow:approve",
		"dingtalk_h5:api:workflow:reject",
		"dingtalk_h5:api:workflow:return",
		"dingtalk_h5:api:workflow:submit",
		"dingtalk_h5:api:workflow:handle",
		"source_grant.`grant_effect`",
		"source_grant.`grant_scope_value`",
		"ON DUPLICATE KEY UPDATE",
		"`permission_status` = 0",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("migration missing %q", snippet)
		}
	}
}
