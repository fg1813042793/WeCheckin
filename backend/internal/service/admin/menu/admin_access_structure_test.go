package menu

import (
	"os"
	"strings"
	"testing"
)

func TestMenuServiceUsesSharedAdminAccessRule(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	for _, snippet := range []string{
		"adminaccess.UserAllowsAdminAccessContext",
		"adminaccess.IsReservedSuperAdminRoleContext",
		"permissionsupport.AdminMenuPermissionsContext",
		"permissionsupport.AdminPermissionKeysContext",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("menu service must use shared admin access rule %q", snippet)
		}
	}
	for _, snippet := range []string{
		"role_allow_admin_login",
		"legacyRoleMenuIDsContext",
		"model.RoleMenu",
		"`role_menus`",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("menu service must not use old role menu access path %q", snippet)
		}
	}
}
