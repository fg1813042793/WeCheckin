package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5WorkflowAPIMigrationCreatesPermissionsWithoutGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_workflow_api_permissions.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("DingTalk H5 workflow API migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read DingTalk H5 workflow API migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:api-category:workflow",
		"dingtalk_h5:api:workflow:view",
		"dingtalk_h5:api:workflow:start",
		"dingtalk_h5:api:workflow:handle",
		"dingtalk_h5:api:workflow:withdraw",
		"/api/v2/dingtalk/h5/workflows/instances",
		"/api/v2/dingtalk/h5/workflows/tasks/:id/complete",
		"'dingtalk_h5'",
		"'api_category'",
		"'api'",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk H5 workflow API migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("DingTalk H5 workflow API migration must not auto-grant permissions")
	}
}
