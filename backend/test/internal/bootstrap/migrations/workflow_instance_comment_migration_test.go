package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInstanceCommentMigrationRegistersPermissionWithoutAutoGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_instance_comment_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow instance comment migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow instance comment migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"dingtalk_h5:api:workflow:comment",
		"/api/v2/dingtalk/h5/workflows/instances/:id/comments",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow instance comment migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("workflow instance comment migration must not auto-grant comment permission")
	}
}
