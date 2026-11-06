package admin

import "testing"

func TestAuditDeleteRoutesRequireIndependentPermissions(t *testing.T) {
	if routeMethodPerms["DELETE /api/v2/admin/logs"] != "log:delete" || routeMethodPerms["DELETE /api/v2/admin/login-logs"] != "log:security-delete" {
		t.Fatal("batch deletion must have independent permissions")
	}
	for path, want := range map[string]string{"/api/v2/admin/logs/:id": "log:delete", "/api/v2/admin/login-logs/:id": "log:security-delete"} {
		found := false
		for _, rule := range routeMethodPermPatterns {
			if rule.method == "DELETE" && rule.path == path && rule.perm == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("single delete missing permission: %s", path)
		}
	}
}
