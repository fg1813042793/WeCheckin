package model

import (
	"reflect"
	"testing"
)

func TestUserRoleModelMapsMultiRoleBindingTable(t *testing.T) {
	if got := (UserRole{}).TableName(); got != "user_roles" {
		t.Fatalf("UserRole table name = %q, want user_roles", got)
	}
	userRoleType := reflect.TypeOf(UserRole{})
	for _, name := range []string{"UserID", "RoleID", "IsPrimary", "Status", "Source", "AddTime", "EditTime"} {
		if _, ok := userRoleType.FieldByName(name); !ok {
			t.Fatalf("UserRole must expose %s", name)
		}
	}
}

func TestUserModelsExposeIgnoredRoleIDsForRuntimeAggregation(t *testing.T) {
	for _, item := range []struct {
		name string
		typ  reflect.Type
	}{
		{name: "User", typ: reflect.TypeOf(User{})},
		{name: "Admin", typ: reflect.TypeOf(Admin{})},
		{name: "DingTalkH5PerfUser", typ: reflect.TypeOf(DingTalkH5PerfUser{})},
	} {
		field, ok := item.typ.FieldByName("RoleIDs")
		if !ok {
			t.Fatalf("%s must expose RoleIDs for multi-role runtime aggregation", item.name)
		}
		if got := field.Tag.Get("gorm"); got != "-" {
			t.Fatalf("%s.RoleIDs gorm tag = %q, want -", item.name, got)
		}
	}
}
