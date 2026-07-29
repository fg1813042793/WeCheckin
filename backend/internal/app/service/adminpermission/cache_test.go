package adminpermission

import (
	"testing"
	"time"
)

func TestPermissionTreeCacheReturnsDeepCopiesAndExpires(t *testing.T) {
	now := time.Now()
	invalidatePermissionTreeCache()

	tree := []*PermissionNode{{
		Key:      "admin:menu:root",
		Name:     "根菜单",
		Children: []*PermissionNode{{Key: "admin:api:list", Name: "列表接口"}},
	}}
	setPermissionTreeCache("admin", []string{"menu", "api"}, tree, now)

	got, ok := getPermissionTreeCache("admin", []string{"menu", "api"}, now.Add(permissionTreeCacheTTL/2))
	if !ok || len(got) != 1 || len(got[0].Children) != 1 {
		t.Fatalf("expected cached permission tree, got %#v ok=%v", got, ok)
	}
	got[0].Name = "mutated"
	got[0].Children[0].Name = "mutated"

	gotAgain, ok := getPermissionTreeCache("admin", []string{"api", "menu"}, now.Add(permissionTreeCacheTTL/2))
	if !ok || gotAgain[0].Name != "根菜单" || gotAgain[0].Children[0].Name != "列表接口" {
		t.Fatalf("permission tree cache should return deep copies, got %#v ok=%v", gotAgain, ok)
	}
	if _, ok := getPermissionTreeCache("admin", []string{"menu", "api"}, now.Add(permissionTreeCacheTTL+time.Second)); ok {
		t.Fatalf("expired permission tree cache should miss")
	}

	invalidatePermissionTreeCache()
	if _, ok := getPermissionTreeCache("admin", []string{"menu", "api"}, now.Add(permissionTreeCacheTTL/2)); ok {
		t.Fatalf("invalidated permission tree cache should miss")
	}
}
