package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestRoleControlsAdminLoginPermission(t *testing.T) {
	roleType := reflect.TypeOf(Role{})
	field, ok := roleType.FieldByName("AllowAdminLogin")
	if !ok {
		t.Fatalf("Role must include AllowAdminLogin")
	}
	tag := string(field.Tag)
	for _, snippet := range []string{`json:"allowAdminLogin"`, "column:role_allow_admin_login", "default:1"} {
		if !strings.Contains(tag, snippet) {
			t.Fatalf("Role.AllowAdminLogin tag must include %s, got %s", snippet, tag)
		}
	}
}
