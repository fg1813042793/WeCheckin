package adminmenuperm

import "testing"

func TestExamTemplateMenuDeclarations(t *testing.T) {
	want := map[string]struct{ parent, perms, typ, path string }{
		"admin:menu:exam:templates":      {parent: "admin:menu:exam", perms: "exam-template:list", typ: TypeMenu, path: "/exam/templates"},
		"admin:menu:exam:templates:edit": {parent: "admin:menu:exam:templates", perms: "exam-template:edit", typ: TypeButton},
	}
	found := make(map[string]Declaration)
	for _, item := range Declarations(true) {
		found[item.Key] = item
	}
	for key, expected := range want {
		item, ok := found[key]
		if !ok {
			t.Fatalf("exam template declaration %s is required", key)
		}
		if item.ParentKey != expected.parent || item.Perms != expected.perms || item.Type != expected.typ || item.Path != expected.path {
			t.Fatalf("exam template declaration %s = %#v", key, item)
		}
	}
}
