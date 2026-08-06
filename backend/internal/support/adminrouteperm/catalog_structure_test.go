package adminrouteperm

import "testing"

func TestAdminRoutePermissionCatalogProvidesCategories(t *testing.T) {
	list := Declarations()
	if len(list) == 0 {
		t.Fatalf("admin route permission catalog must not be empty")
	}
	for _, item := range list {
		if item.CategoryKey == "" {
			t.Fatalf("api permission %s must include category key", item.Key)
		}
		if item.CategoryName == "" {
			t.Fatalf("api permission %s must include category name", item.Key)
		}
	}
	categories := Categories()
	if len(categories) == 0 {
		t.Fatalf("admin route permission categories must not be empty")
	}
	if categories[0].Key == "" || categories[0].Name == "" {
		t.Fatalf("admin route permission categories must expose key and name")
	}
}

func TestAdminRoutePermissionCatalogIncludesPositionPermissions(t *testing.T) {
	want := map[string]bool{
		"position:list": false,
		"position:add":  false,
		"position:edit": false,
		"position:del":  false,
	}
	for _, item := range Declarations() {
		if _, ok := want[item.Perms]; ok {
			want[item.Perms] = true
			if item.CategoryKey != "admin:api-category:user" {
				t.Fatalf("position api permission %s should be grouped under user category, got %s", item.Perms, item.CategoryKey)
			}
		}
	}
	for perms, ok := range want {
		if !ok {
			t.Fatalf("admin route permission catalog missing %s", perms)
		}
	}
}

func TestAdminRoutePermissionCatalogSplitsDingTalkPermissions(t *testing.T) {
	want := map[string]bool{
		"dingtalk:settings:list":  false,
		"dingtalk:settings:edit":  false,
		"dingtalk:bindings:list":  false,
		"dingtalk:bindings:edit":  false,
	}
	for _, item := range Declarations() {
		if _, ok := want[item.Perms]; ok {
			want[item.Perms] = true
			if item.CategoryKey != "admin:api-category:dingtalk" {
				t.Fatalf("dingtalk api permission %s should be grouped under dingtalk category, got %s", item.Perms, item.CategoryKey)
			}
		}
		if item.Perms == "dingtalk:config" {
			t.Fatalf("dingtalk admin api permissions should use split settings/bindings permissions")
		}
	}
	for perms, ok := range want {
		if !ok {
			t.Fatalf("admin route permission catalog missing %s", perms)
		}
	}
}
