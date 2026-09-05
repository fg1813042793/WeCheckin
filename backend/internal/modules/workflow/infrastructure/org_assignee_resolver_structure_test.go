package infrastructure

import (
	"os"
	"sort"
	"strings"
	"testing"
)

func TestOrgApproverIdentityResolverStructure(t *testing.T) {
	modelSource := readWorkflowStructureFile(t, "../../../model/workflow/org_approver.go")
	coreTypes := readWorkflowStructureFile(t, "../../../workflowcore/types.go")
	resolverSource := readWorkflowStructureFile(t, "assignee_resolver.go")

	for _, want := range []string{
		"OrgApproverIdentity",
		"workflow_org_approver_identities",
		"OrgApproverAssignment",
		"workflow_org_approver_assignments",
		"IdentityCode",
		"SubjectType",
		"SubjectID",
		"OrgApproverSubjectTypeDepartment",
		"OrgApproverSubjectTypeUser",
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
		"user_reporting_relations",
		"employee_user_id",
		"subject_type",
		"subject_id",
		"OrgApproverSubjectTypeUser",
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

func readWorkflowPackageSource(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read workflow infrastructure package: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	var source strings.Builder
	for _, name := range names {
		source.WriteString(readWorkflowStructureFile(t, name))
		source.WriteString("\n")
	}
	return source.String()
}
