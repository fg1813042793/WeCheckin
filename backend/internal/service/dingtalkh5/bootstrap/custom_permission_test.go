package bootstrap

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/support/appmenuperm"
)

func TestOrderedDingTalkH5PermissionKeysAppendCustomPermissions(t *testing.T) {
	selected := map[string]bool{
		"dingtalk_h5:menu:dashboard":       true,
		"dingtalk_h5:menu:custom":          true,
		"dingtalk_h5:button:custom:export": true,
		"dingtalk_h5:api:custom:view":      true,
		"client:api:custom:view":           true,
	}

	if got, want := orderedDingTalkH5MenuKeys(selected), []string{
		"dingtalk_h5:menu:dashboard",
		"dingtalk_h5:menu:custom",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("menu keys = %#v, want %#v", got, want)
	}
	if got, want := orderedDingTalkH5ButtonKeys(selected), []string{
		"dingtalk_h5:button:custom:export",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("button keys = %#v, want %#v", got, want)
	}
	if got, want := orderedDingTalkH5APIKeys(selected), []string{
		"dingtalk_h5:api:custom:view",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api keys = %#v, want %#v", got, want)
	}
}

func TestDingTalkH5MenuDeclarationsIncludeCustomCatalogRows(t *testing.T) {
	declarations := dingTalkH5MenuDeclarationsFromCatalog([]dingTalkH5PermissionCatalogRow{
		{
			Key:          "dingtalk_h5:menu:custom",
			Name:         "自定义模块",
			Type:         appmenuperm.TypeMenu,
			ResourcePath: "custom",
			Icon:         "grid",
			Sort:         110,
		},
	})

	var custom *appmenuperm.Declaration
	for index := range declarations {
		if declarations[index].Key == "dingtalk_h5:menu:custom" {
			custom = &declarations[index]
			break
		}
	}
	if custom == nil {
		t.Fatal("custom menu declaration not found")
	}
	if custom.Path != "custom" || custom.Name != "自定义模块" || custom.Icon != "grid" {
		t.Fatalf("custom menu declaration = %#v", *custom)
	}
}

func TestDingTalkH5MenusBuildCustomCatalogNodesAndAncestors(t *testing.T) {
	declarations := append(appmenuperm.DingTalkH5MenuDeclarations(),
		appmenuperm.Declaration{
			Key:      "dingtalk_h5:menu:custom",
			Name:     "自定义模块",
			Platform: "dingtalk_h5",
			Type:     appmenuperm.TypeDirectory,
			Path:     "custom",
			Icon:     "grid",
			Sort:     110,
		},
		appmenuperm.Declaration{
			Key:       "dingtalk_h5:menu:custom:report",
			Name:      "自定义报表",
			Platform:  "dingtalk_h5",
			Type:      appmenuperm.TypeMenu,
			Path:      "custom:report",
			Icon:      "summary",
			ParentKey: "dingtalk_h5:menu:custom",
			Sort:      120,
		},
	)

	menus := dingTalkH5MenusByDeclarations([]string{
		"dingtalk_h5:menu:custom:report",
	}, declarations, nil, nil)

	if len(menus) != 1 {
		t.Fatalf("top-level menus = %d, want 1: %#v", len(menus), menus)
	}
	if got := menus[0].Key; got != "custom" {
		t.Fatalf("root menu key = %q, want custom", got)
	}
	if len(menus[0].Children) != 1 || menus[0].Children[0].Key != "custom:report" {
		t.Fatalf("custom children = %#v, want custom:report", menus[0].Children)
	}
}
