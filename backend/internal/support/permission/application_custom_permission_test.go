package permission

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/support/appapiperm"
	"wecheckin/backend/internal/support/appmenuperm"
)

func TestNormalizeRoleApplicationMenuKeysKeepsCustomPermissions(t *testing.T) {
	got := normalizeRoleApplicationMenuKeys(
		[]string{"client:menu:custom"},
		[]string{"dingtalk_h5:menu:custom", "dingtalk_h5:button:custom:submit"},
	)
	want := []string{"client:menu:custom", "dingtalk_h5:menu:custom", "dingtalk_h5:button:custom:submit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("custom application menu permissions must be preserved, got %#v", got)
	}
}

func TestNormalizeRoleApplicationAPIKeysKeepsCustomPermissions(t *testing.T) {
	got := normalizeRoleApplicationAPIKeys(
		[]string{"client:api:custom:list"},
		[]string{"dingtalk_h5:api:custom:list"},
	)
	want := []string{"client:api:custom:list", "dingtalk_h5:api:custom:list"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("custom application API permissions must be preserved, got %#v", got)
	}
}

func TestOrderedApplicationPermissionKeysAppendCustomPermissions(t *testing.T) {
	menuGot := orderedApplicationMenuKeys(
		map[string]bool{"client:menu:known": true, "client:menu:custom": true},
		[]appmenuperm.Declaration{{Key: "client:menu:known"}},
	)
	if want := []string{"client:menu:known", "client:menu:custom"}; !reflect.DeepEqual(menuGot, want) {
		t.Fatalf("role menu query must return custom grants, got %#v", menuGot)
	}

	apiGot := orderedApplicationAPIKeys(
		map[string]bool{"client:api:known": true, "client:api:custom": true},
		[]appapiperm.Declaration{{Key: "client:api:known"}},
	)
	if want := []string{"client:api:known", "client:api:custom"}; !reflect.DeepEqual(apiGot, want) {
		t.Fatalf("role API query must return custom grants, got %#v", apiGot)
	}
}
