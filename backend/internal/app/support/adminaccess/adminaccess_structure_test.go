package adminaccess

import (
	"os"
	"strings"
	"testing"
)

func TestAdminAccessRuleUsesRoleStatusAndAssignedMenus(t *testing.T) {
	src, err := os.ReadFile("adminaccess.go")
	if err != nil {
		t.Fatalf("read adminaccess.go: %v", err)
	}
	text := string(src)

	for _, snippet := range []string{
		"ReservedSuperAdminRoleName",
		"RoleAllowsAdminAccessContext",
		"UserAllowsAdminAccessContext",
		"ApplyUserAdminAccessRoleFilter",
		"permissionsupport.AdminLoginPermissionKey",
		"permissionsupport.SubjectHasPermissionContext",
		"permissionsupport.RoleHasPermissionContext",
		"`role_status` = 1",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin access rule must include %q", snippet)
		}
	}

	for _, snippet := range []string{
		"role_allow_admin_login",
		"user_admin_enabled",
		"user_admin_type",
		"`role_menus`",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("admin access rule must not use old login field %q", snippet)
		}
	}
}
