package menu

import (
	"os"
	"strings"
	"testing"
)

func TestMenuRuntimePrefersUnifiedPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissionsupport.AdminMenuPermissionsContext",
		"permissionsToMenuTree",
		"permissionRowsToMenus",
		"permission_perms",
		"type AdminMenu struct",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("menu runtime must use unified permissions with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"model.Menu",
		"`menu_status`",
		"`menu_sort`",
		"legacyAdminMenuTreeContext",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("menu runtime must not read legacy menus table snippet %s", forbidden)
		}
	}
}
