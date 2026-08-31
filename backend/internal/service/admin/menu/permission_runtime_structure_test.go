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
		"permissionsupport.AdminPermissionKeysContext",
		"permissionsToMenuTree",
		"permissionRowsToMenus",
		"type AdminMenu struct",
		"permissionKeys(rows)",
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
		"func GetTree()",
		"func GetTreeContext",
		"func GetList()",
		"func GetListContext",
		"flattenAdminMenus",
		"permissionsupport.AdminPermCodesContext",
		"allAdminPermissionsWithPermCodesContext",
		"permissionPermCodes",
		"`permission_perms` <> ''",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("menu runtime must not read legacy menus table snippet %s", forbidden)
		}
	}
}

func TestMenuRuntimeDoesNotBackfillPermissionCatalogOnRead(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	if strings.Contains(text, "ensureAdminRuntimePermissionCatalogContext") {
		t.Fatalf("menu runtime reads must not create permission catalog entries")
	}
	if strings.Contains(text, "EnsureApplicationPermissionCatalogContext") {
		t.Fatalf("menu runtime reads must rely on migration-created permissions, not catalog backfill")
	}
}
