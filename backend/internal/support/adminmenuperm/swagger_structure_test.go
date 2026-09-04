package adminmenuperm

import "testing"

func TestAdminSwaggerMenuDeclaration(t *testing.T) {
	for _, item := range Declarations(false) {
		if item.Key != "admin:menu:swagger" {
			continue
		}
		if item.Name != "接口文档" || item.Path != "/swagger-docs" {
			t.Fatalf("swagger menu name/path = %q/%q", item.Name, item.Path)
		}
		if item.Perms != "swagger:view" || item.Icon != "Document" {
			t.Fatalf("swagger menu perms/icon = %q/%q", item.Perms, item.Icon)
		}
		if item.ParentKey != "" || item.Type != TypeMenu || item.Sort != 19 {
			t.Fatalf("swagger menu parent/type/sort = %q/%q/%d", item.ParentKey, item.Type, item.Sort)
		}
		return
	}
	t.Fatal("admin menu declarations missing swagger documentation menu")
}
