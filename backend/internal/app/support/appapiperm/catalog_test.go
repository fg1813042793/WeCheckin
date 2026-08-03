package appapiperm

import (
	"strings"
	"testing"
)

func TestDingTalkH5APIDeclarationsAreCategorized(t *testing.T) {
	categories := DingTalkH5APICategories()
	if len(categories) == 0 {
		t.Fatalf("dingtalk h5 API categories must not be empty")
	}
	categoryKeys := map[string]bool{}
	for _, category := range categories {
		if !strings.HasPrefix(category.Key, "dingtalk_h5:api-category:") {
			t.Fatalf("unexpected dingtalk h5 API category key %q", category.Key)
		}
		categoryKeys[category.Key] = true
	}

	declarations := DingTalkH5APIDeclarations()
	if len(declarations) == 0 {
		t.Fatalf("dingtalk h5 API declarations must not be empty")
	}
	required := map[string]bool{
		"dingtalk_h5:api:workbench:view":     false,
		"dingtalk_h5:api:review:self_submit": false,
		"dingtalk_h5:api:review:hrbp_submit": false,
		"dingtalk_h5:api:review:finalize":    false,
		"dingtalk_h5:api:user:edit":          false,
		"dingtalk_h5:api:template:view":      false,
		"dingtalk_h5:api:template:save":      false,
	}
	for _, declaration := range declarations {
		if !strings.HasPrefix(declaration.Key, "dingtalk_h5:api:") {
			t.Fatalf("unexpected dingtalk h5 API permission key %q", declaration.Key)
		}
		if !categoryKeys[declaration.CategoryKey] {
			t.Fatalf("declaration %q points to unknown category %q", declaration.Key, declaration.CategoryKey)
		}
		if _, ok := required[declaration.Key]; ok {
			required[declaration.Key] = true
		}
	}
	for key, ok := range required {
		if !ok {
			t.Fatalf("missing dingtalk h5 API permission %s", key)
		}
	}
}

func TestClientAPIDeclarationsAreCategorized(t *testing.T) {
	categories := ClientAPICategories()
	if len(categories) == 0 {
		t.Fatalf("client API categories must not be empty")
	}
	categoryKeys := map[string]bool{}
	for _, category := range categories {
		if !strings.HasPrefix(category.Key, "client:api-category:") {
			t.Fatalf("unexpected client API category key %q", category.Key)
		}
		categoryKeys[category.Key] = true
	}

	declarations := ClientAPIDeclarations()
	if len(declarations) == 0 {
		t.Fatalf("client API declarations must not be empty")
	}
	required := map[string]bool{
		"client:api:bootstrap:view":  false,
		"client:api:user:view":       false,
		"client:api:news:view":       false,
		"client:api:enroll:submit":   false,
		"client:api:event:score":     false,
		"client:api:survey:response": false,
		"client:api:exam:answer":     false,
	}
	for _, declaration := range declarations {
		if !strings.HasPrefix(declaration.Key, "client:api:") {
			t.Fatalf("unexpected client API permission key %q", declaration.Key)
		}
		if !categoryKeys[declaration.CategoryKey] {
			t.Fatalf("declaration %q points to unknown category %q", declaration.Key, declaration.CategoryKey)
		}
		if _, ok := required[declaration.Key]; ok {
			required[declaration.Key] = true
		}
	}
	for key, ok := range required {
		if !ok {
			t.Fatalf("missing client API permission %s", key)
		}
	}

	foundBootstrapRoute := false
	for _, route := range ClientRouteDeclarations() {
		if route.Method == "GET" && route.Path == "/api/v2/me/bootstrap" && route.PermissionKey == "client:api:bootstrap:view" {
			foundBootstrapRoute = true
			break
		}
	}
	if !foundBootstrapRoute {
		t.Fatalf("client bootstrap route must be protected by client:api:bootstrap:view")
	}
}
