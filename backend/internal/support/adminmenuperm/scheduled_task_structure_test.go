package adminmenuperm

import "testing"

func TestScheduledTaskMenuDeclarations(t *testing.T) {
	want := map[string]struct {
		parent string
		perms  string
	}{
		"admin:menu:scheduled-task":            {perms: "scheduled-task:list"},
		"admin:menu:scheduled-task:tasks":      {parent: "admin:menu:scheduled-task", perms: "scheduled-task:list"},
		"admin:menu:scheduled-task:runs":       {parent: "admin:menu:scheduled-task", perms: "scheduled-task:run:list"},
		"admin:menu:scheduled-task:workers":    {parent: "admin:menu:scheduled-task", perms: "scheduled-task:worker:list"},
		"admin:menu:scheduled-task:add":        {parent: "admin:menu:scheduled-task:tasks", perms: "scheduled-task:add"},
		"admin:menu:scheduled-task:run":        {parent: "admin:menu:scheduled-task:tasks", perms: "scheduled-task:run"},
		"admin:menu:scheduled-task:shell":      {parent: "admin:menu:scheduled-task:tasks", perms: "scheduled-task:shell"},
		"admin:menu:scheduled-task:sql:write":  {parent: "admin:menu:scheduled-task:tasks", perms: "scheduled-task:sql:write"},
		"admin:menu:scheduled-task:run:retry":  {parent: "admin:menu:scheduled-task:runs", perms: "scheduled-task:run:retry"},
		"admin:menu:scheduled-task:run:cancel": {parent: "admin:menu:scheduled-task:runs", perms: "scheduled-task:run:cancel"},
	}
	found := make(map[string]Declaration)
	for _, item := range Declarations(false) {
		found[item.Key] = item
	}
	for key, expected := range want {
		item, ok := found[key]
		if !ok {
			t.Fatalf("scheduled task menu declaration %s is required", key)
		}
		if item.ParentKey != expected.parent || item.Perms != expected.perms {
			t.Fatalf("%s = parent %q perms %q", key, item.ParentKey, item.Perms)
		}
	}
}
