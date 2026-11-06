package admin

import (
	"testing"
	"wecheckin/backend/internal/support/adminrouteperm"
)

func TestPublicPortalRoutesHaveDedicatedPermissions(t *testing.T) {
	for _, row := range []struct{ method, path, permission string }{
		{"GET", "/api/v2/admin/public-portal", "public-portal:list"},
		{"GET", "/api/v2/admin/public-portal/options", "public-portal:list"},
		{"GET", "/api/v2/admin/public-portal/preview", "public-portal:list"},
		{"PUT", "/api/v2/admin/public-portal/settings", "public-portal:edit"},
		{"POST", "/api/v2/admin/public-portal/items", "public-portal:edit"},
		{"PUT", "/api/v2/admin/public-portal/items/7", "public-portal:edit"},
		{"DELETE", "/api/v2/admin/public-portal/items/7", "public-portal:edit"},
	} {
		got, ok := adminRoutePermission(row.method, row.path)
		if !ok || got != row.permission {
			t.Fatalf("%s %s permission=%q found=%t", row.method, row.path, got, ok)
		}
	}
	for _, key := range []string{"admin:api:public-portal:list", "admin:api:public-portal:edit"} {
		found := false
		for _, d := range adminrouteperm.Declarations() {
			if d.Key == key {
				found = true
				if d.CategoryKey != "admin:api-category:public-portal" {
					t.Fatal(d)
				}
			}
		}
		if !found {
			t.Fatalf("missing declaration %s", key)
		}
	}
}
