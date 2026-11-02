package adminmenuperm

import "testing"

func TestSurveyTemplateMenuDeclarations(t *testing.T) {
	want := map[string]struct {
		parent string
		perms  string
		typ    string
		path   string
	}{
		"admin:menu:survey:templates":      {parent: "admin:menu:survey", perms: "survey-template:list", typ: TypeMenu, path: "/survey/templates"},
		"admin:menu:survey:templates:edit": {parent: "admin:menu:survey:templates", perms: "survey-template:edit", typ: TypeButton},
	}
	found := make(map[string]Declaration)
	for _, item := range Declarations(false) {
		found[item.Key] = item
	}
	for key, expected := range want {
		item, ok := found[key]
		if !ok {
			t.Fatalf("survey template declaration %s is required", key)
		}
		if item.ParentKey != expected.parent || item.Perms != expected.perms || item.Type != expected.typ || item.Path != expected.path {
			t.Fatalf("survey template declaration %s = %#v", key, item)
		}
	}
}
