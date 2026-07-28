package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestAdminCompatibilityModelUsesUsersTable(t *testing.T) {
	if got := (Admin{}).TableName(); got != "users" {
		t.Fatalf("Admin compatibility model must use users table, got %q", got)
	}

	adminType := reflect.TypeOf(Admin{})
	requiredTags := map[string][]string{
		"Name":      {"column:user_name"},
		"Password":  {"column:user_password"},
		"Desc":      {"column:user_admin_desc"},
		"Pic":       {"column:user_pic"},
		"Phone":     {"column:user_mobile"},
		"Status":    {"column:user_status"},
		"Type":      {"column:user_admin_type"},
		"RoleID":    {"column:user_role_id"},
		"Token":     {"column:user_admin_token"},
		"TokenTime": {"column:user_admin_token_time"},
	}
	for fieldName, snippets := range requiredTags {
		field, ok := adminType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("Admin compatibility model must include %s", fieldName)
		}
		tag := string(field.Tag)
		for _, snippet := range snippets {
			if !strings.Contains(tag, snippet) {
				t.Fatalf("Admin.%s tag must include %s, got %s", fieldName, snippet, tag)
			}
		}
	}
	if _, ok := adminType.FieldByName("AdminEnabled"); ok {
		t.Fatalf("Admin compatibility model must not expose separate backend-login switch")
	}
}

func TestUserCarriesAdminPermissionFields(t *testing.T) {
	userType := reflect.TypeOf(User{})
	requiredTags := map[string][]string{
		"Account":        {`json:"account"`, "column:user_account"},
		"AdminEnabled":   {`json:"adminEnabled"`, "column:user_admin_enabled"},
		"AdminType":      {`json:"adminType"`, "column:user_admin_type"},
		"RoleID":         {`json:"roleId"`, "column:user_role_id"},
		"AdminDesc":      {`json:"adminDesc"`, "column:user_admin_desc"},
		"AdminToken":     {"column:user_admin_token"},
		"AdminTokenTime": {"column:user_admin_token_time"},
	}
	for fieldName, snippets := range requiredTags {
		field, ok := userType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("User must include %s for merged admin/user accounts", fieldName)
		}
		tag := string(field.Tag)
		for _, snippet := range snippets {
			if !strings.Contains(tag, snippet) {
				t.Fatalf("User.%s tag must include %s, got %s", fieldName, snippet, tag)
			}
		}
	}
}
