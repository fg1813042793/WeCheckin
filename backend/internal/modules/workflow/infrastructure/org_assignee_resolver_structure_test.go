package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestOrgApproverIdentityResolverStructure(t *testing.T) {
	modelSource := readWorkflowStructureFile(t, "../../../model/workflow/org_approver.go")
	coreTypes := readWorkflowStructureFile(t, "../../../workflow/types.go")
	resolverSource := readWorkflowStructureFile(t, "assignee_resolver.go")

	for _, want := range []string{
		"OrgApproverIdentity",
		"workflow_org_approver_identities",
		"OrgApproverAssignment",
		"workflow_org_approver_assignments",
		"IdentityCode",
		"DepartmentID",
		"UserID",
		"Sort",
	} {
		if !strings.Contains(modelSource, want) {
			t.Fatalf("workflow org approver model must include snippet %q", want)
		}
	}
	if !strings.Contains(coreTypes, "AssigneeTypeOrgIdentity") {
		t.Fatalf("workflow core types must define AssigneeTypeOrgIdentity")
	}
	for _, want := range []string{
		"resolveOrgIdentity",
		"starter_department:",
		"department:",
		"manager_user_id",
		"department_leader",
	} {
		if !strings.Contains(resolverSource, want) {
			t.Fatalf("assignee resolver must include organization identity snippet %q", want)
		}
	}
}

func readWorkflowStructureFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
