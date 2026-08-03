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
	cacheSrc, err := os.ReadFile("cache.go")
	if err != nil {
		t.Fatalf("read cache.go: %v", err)
	}
	text := string(src) + string(cacheSrc)

	for _, snippet := range []string{
		"ReservedSuperAdminRoleName",
		"RoleAllowsAdminAccessContext",
		"UserAllowsAdminAccessContext",
		"UserAllowsAdminAccessWithRoleIDsContext",
		"HasReservedSuperAdminRoleWithRoleIDsContext",
		"activeRoleByIDContext",
		"activeRolesByIDsContext",
		"getRoleAccessCache",
		"setRoleAccessCache",
		"InvalidateAdminAccessCacheForRole",
		"Take(&role)",
		"ApplyUserAdminAccessRoleFilter",
		"permissionsupport.AdminLoginPermissionKey",
		"permissionsupport.SubjectHasPermissionWithRoleIDsContext",
		"permissionsupport.SubjectHasPermissionContext",
		"permissionsupport.RoleHasPermissionContext",
		"`user_roles`",
		"`role_status` = 1",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin access rule must include %q", snippet)
		}
	}
	if strings.Contains(text, "First(&role)") {
		t.Fatalf("admin access role lookup should use Take to avoid unnecessary ORDER BY id")
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
