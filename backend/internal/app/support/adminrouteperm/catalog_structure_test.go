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
