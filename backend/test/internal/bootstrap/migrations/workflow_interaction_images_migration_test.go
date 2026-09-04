package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowInteractionImagesMigrationAddsStorageAndCommentUploadGrants(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_interaction_images.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("workflow interaction images migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read workflow interaction images migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"workflow_process_tasks", "task_images_json",
		"workflow_process_history", "event_images_json",
		"dingtalk_h5:api:workflow:comment", "dingtalk_h5:api:workflow:attachment",
		"INFORMATION_SCHEMA.COLUMNS", "ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workflow interaction images migration missing %q", snippet)
		}
	}
}
