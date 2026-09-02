package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5WorkflowMenuMigrationCreatesPermissionWithoutGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_workflow_menu.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("DingTalk H5 workflow menu migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read DingTalk H5 workflow menu migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:menu:workflow",
		"'流程审批'",
		"'dingtalk_h5'",
		"'menu'",
		"'workflow'",
		"'workflow'",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 workflow menu migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("DingTalk H5 workflow menu migration must not auto-grant permissions")
	}
}
