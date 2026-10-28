package adminpermission

import "testing"

func TestFilterPermissionTreeKeepsAllowedLeavesAndAncestors(t *testing.T) {
	tree := []*PermissionNode{{
		Key: "admin:api-category:system", Name: "系统权限", Children: []*PermissionNode{
			{Key: "admin:api:menu:list", Name: "菜单查看"},
			{Key: "admin:api:setup:edit", Name: "系统设置"},
		},
	}}
	filtered := filterPermissionTree(tree, map[string]struct{}{"admin:api:menu:list": {}})
	if len(filtered) != 1 || len(filtered[0].Children) != 1 || filtered[0].Children[0].Key != "admin:api:menu:list" {
		t.Fatalf("filtered tree = %#v", filtered)
	}
}

