package adminmenuperm

import "testing"

func TestAdminMenuDeclarationsIncludePositionManagement(t *testing.T) {
	want := map[string]bool{
		"position:list": false,
		"position:add":  false,
		"position:edit": false,
		"position:del":  false,
	}
	var foundMenu bool
	for _, item := range Declarations(true) {
		if item.Key == "admin:menu:position" {
			foundMenu = true
			if item.Path != "/position" {
				t.Fatalf("position menu path = %q, want /position", item.Path)
			}
			if item.Perms != "position:list" {
				t.Fatalf("position menu perms = %q, want position:list", item.Perms)
			}
		}
		if _, ok := want[item.Perms]; ok {
			want[item.Perms] = true
			if item.ParentKey != "admin:menu:position" && item.Key != "admin:menu:position" {
				t.Fatalf("position permission %s should belong to admin:menu:position", item.Key)
			}
		}
	}
	if !foundMenu {
		t.Fatalf("admin menu declarations missing position management menu")
	}
	for perms, ok := range want {
		if !ok {
			t.Fatalf("admin menu declarations missing position permission %s", perms)
		}
	}
}
