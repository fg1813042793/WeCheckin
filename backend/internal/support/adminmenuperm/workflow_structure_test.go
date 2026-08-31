package adminmenuperm

import "testing"

func TestWorkflowRuntimeMenuDeclarations(t *testing.T) {
	want := map[string]struct {
		parent string
		perms  string
	}{
		"admin:menu:workflow:instances":         {parent: "admin:menu:workflow", perms: "workflow:instance:list"},
		"admin:menu:workflow:tasks":             {parent: "admin:menu:workflow", perms: "workflow:task:list"},
		"admin:menu:workflow:org-approvers":     {parent: "admin:menu:workflow", perms: "workflow:org-approver:list"},
		"admin:menu:workflow:instance:start":    {parent: "admin:menu:workflow:instances", perms: "workflow:instance:start"},
		"admin:menu:workflow:instance:cancel":   {parent: "admin:menu:workflow:instances", perms: "workflow:instance:cancel"},
		"admin:menu:workflow:task:complete":     {parent: "admin:menu:workflow:tasks", perms: "workflow:task:complete"},
		"admin:menu:workflow:org-approver:list": {parent: "admin:menu:workflow:org-approvers", perms: "workflow:org-approver:list"},
		"admin:menu:workflow:org-approver:edit": {parent: "admin:menu:workflow:org-approvers", perms: "workflow:org-approver:edit"},
	}

	found := make(map[string]Declaration)
	for _, item := range Declarations(false) {
		found[item.Key] = item
	}
	for key, expected := range want {
		item, ok := found[key]
		if !ok {
			t.Fatalf("workflow runtime menu declaration %s is required", key)
		}
		if item.ParentKey != expected.parent || item.Perms != expected.perms {
			t.Fatalf("workflow runtime menu %s = parent %q perms %q, want parent %q perms %q", key, item.ParentKey, item.Perms, expected.parent, expected.perms)
		}
	}
}
