package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowCompletedFormRevisionMigration(t *testing.T) {
	source := readWorkflowCompletedRevisionMigration(t, "*_create_workflow_completed_form_revisions.sql")
	for _, token := range []string{
		"workflow_form_revision_requests",
		"workflow_form_revision_tasks",
		"workflow_business_event_outbox",
		"uk_workflow_form_revision_active",
		"workflow.business-event.dispatch_due",
	} {
		if !strings.Contains(source, token) {
			t.Fatalf("migration missing %s", token)
		}
	}
}

func TestWorkflowCompletedRevisionPermissionsAreRegisteredWithoutRoleGrant(t *testing.T) {
	source := readWorkflowCompletedRevisionMigration(t, "*_add_workflow_completed_form_revision_permissions.sql")
	for _, key := range []string{
		"dingtalk_h5:button:workflow:form-revision-create",
		"dingtalk_h5:api:workflow:form-revision-create",
		"dingtalk_h5:button:workflow:form-revision-handle",
		"dingtalk_h5:api:workflow:form-revision-handle",
	} {
		if !strings.Contains(source, key) {
			t.Fatalf("permission migration missing %s", key)
		}
	}
	if strings.Contains(strings.ToLower(source), "permission_grants") {
		t.Fatal("completed revision permissions must not be granted automatically")
	}
}

func readWorkflowCompletedRevisionMigration(t *testing.T, pattern string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", pattern))
	if err != nil || len(matches) != 1 {
		t.Fatalf("migration matches=%v err=%v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}
