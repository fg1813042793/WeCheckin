package adminmenuperm

import "testing"

func TestNotificationMenuDeclarations(t *testing.T) {
	want := map[string]struct {
		parent string
		perms  string
		typ    string
	}{
		"admin:menu:notification":               {perms: "notification:list", typ: TypeMenu},
		"admin:menu:notification:list":          {parent: "admin:menu:notification", perms: "notification:list", typ: TypeButton},
		"admin:menu:notification:read":          {parent: "admin:menu:notification", perms: "notification:read", typ: TypeButton},
		"admin:menu:notification:send":          {parent: "admin:menu:notification", perms: "notification:send", typ: TypeButton},
		"admin:menu:notification:dingtalk-send": {parent: "admin:menu:notification", perms: "notification:dingtalk:send", typ: TypeButton},
	}

	found := make(map[string]Declaration)
	for _, item := range Declarations(false) {
		found[item.Key] = item
	}
	for key, expected := range want {
		item, ok := found[key]
		if !ok {
			t.Fatalf("notification menu declaration %s is required", key)
		}
		if item.ParentKey != expected.parent || item.Perms != expected.perms || item.Type != expected.typ {
			t.Fatalf("notification menu %s = parent %q perms %q type %q, want parent %q perms %q type %q", key, item.ParentKey, item.Perms, item.Type, expected.parent, expected.perms, expected.typ)
		}
	}
}
