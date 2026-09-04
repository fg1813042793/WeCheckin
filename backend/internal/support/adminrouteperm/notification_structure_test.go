package adminrouteperm

import "testing"

func TestNotificationRoutePermissionCatalog(t *testing.T) {
	want := map[string]struct {
		method string
		path   string
	}{
		"notification:list":          {method: "GET", path: "/api/v2/admin/in-app-notifications"},
		"notification:read":          {method: "PATCH", path: "/api/v2/admin/in-app-notifications/:id/read"},
		"notification:send":          {method: "POST", path: "/api/v2/admin/in-app-notifications"},
		"notification:dingtalk:send": {method: "POST", path: "/api/v2/admin/dingtalk-notifications"},
	}

	found := make(map[string]Declaration)
	for _, item := range Declarations() {
		found[item.Perms] = item
	}
	for permission, expected := range want {
		item, ok := found[permission]
		if !ok {
			t.Fatalf("notification route permission %s is required", permission)
		}
		if item.Method != expected.method || item.Path != expected.path {
			t.Fatalf("notification route permission %s = %s %s, want %s %s", permission, item.Method, item.Path, expected.method, expected.path)
		}
	}
}
