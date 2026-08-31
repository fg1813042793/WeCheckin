package adminrouteperm

import "testing"

func TestWorkflowRuntimeRouteDeclarations(t *testing.T) {
	want := map[string]string{
		"workflow:instance:list":     "/api/v2/admin/workflow-instances",
		"workflow:instance:start":    "/api/v2/admin/workflow-instances",
		"workflow:instance:detail":   "/api/v2/admin/workflow-instances/:id",
		"workflow:instance:cancel":   "/api/v2/admin/workflow-instances/:id/cancel",
		"workflow:task:list":         "/api/v2/admin/workflow-tasks",
		"workflow:task:complete":     "/api/v2/admin/workflow-tasks/:id/complete",
		"workflow:org-approver:list": "/api/v2/admin/workflow-org-approver-identities",
		"workflow:org-approver:edit": "/api/v2/admin/workflow-org-approver-assignments",
	}
	found := make(map[string]Declaration)
	for _, item := range Declarations() {
		found[item.Perms] = item
	}
	for perms, path := range want {
		item, ok := found[perms]
		if !ok {
			t.Fatalf("workflow runtime permission %s is required", perms)
		}
		if item.Path != path {
			t.Fatalf("workflow runtime permission %s path = %q, want %q", perms, item.Path, path)
		}
	}
}
