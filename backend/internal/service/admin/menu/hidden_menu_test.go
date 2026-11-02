package menu

import (
	"testing"

	"wecheckin/backend/internal/model"
)

func TestPermissionsToMenuTreeKeepsHiddenSubtreeForAuthorization(t *testing.T) {
	visible := 0
	hidden := 1
	rows := []model.Permission{
		{Key: "admin:menu:root", Name: "Root", Platform: "admin", Type: "directory", Hidden: &visible},
		{Key: "admin:menu:hidden", Name: "Hidden", Platform: "admin", Type: "menu", ParentKey: "admin:menu:root", Hidden: &hidden},
		{Key: "admin:button:hidden:read", Name: "Hidden child", Platform: "admin", Type: "button", ParentKey: "admin:menu:hidden", Hidden: &visible},
		{Key: "admin:menu:visible", Name: "Visible", Platform: "admin", Type: "menu", ParentKey: "admin:menu:root", Hidden: &visible},
	}

	tree := permissionsToMenuTree(rows)
	if len(tree) != 1 || tree[0].Key != "admin:menu:root" {
		t.Fatalf("unexpected root tree: %#v", tree)
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("authorization tree must retain hidden and visible children: %#v", tree[0].Children)
	}
	if tree[0].Children[0].Key != "admin:menu:hidden" || tree[0].Children[0].Hidden != 1 {
		t.Fatalf("hidden menu marker must remain in authorization tree: %#v", tree[0].Children[0])
	}
	if len(tree[0].Children[0].Children) != 1 || tree[0].Children[0].Children[0].Key != "admin:button:hidden:read" {
		t.Fatalf("hidden menu descendants must remain available for authorization: %#v", tree[0].Children[0].Children)
	}
}

func TestHiddenMenuPermissionKeyRemainsAuthorized(t *testing.T) {
	hidden := 1
	keys := permissionKeys([]model.Permission{{Key: "admin:menu:hidden", Hidden: &hidden}})
	if len(keys) != 1 || keys[0] != "admin:menu:hidden" {
		t.Fatalf("hidden menus must keep permission keys: %#v", keys)
	}
}
