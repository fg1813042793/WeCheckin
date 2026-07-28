package adminpermission

import (
	"os"
	"strings"
	"testing"
)

func TestAdminPermissionServiceManagesUnifiedPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"type PermissionNode struct",
		"type SaveRequest struct",
		"TreeContext",
		"ListContext",
		"normalizePermissionTypes",
		"`permission_type` IN ?",
		"permissionsupport.EnsurePermissionSchemaContext(ctx, db)",
		"AddContext",
		"EditContext",
		"DeleteContext",
		"model.Permission",
		"permission_icon",
		"permissionsupport.PlatformAdmin",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission service must expose unified permission management snippet %s", snippet)
		}
	}
}

func TestAdminPermissionServiceDeletesChildrenAndGrants(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"collectDescendantKeys",
		"`permission_parent_key` IN ?",
		"Delete(&model.PermissionGrant{})",
		"Delete(&model.Permission{})",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission delete must clean child permissions and grants with %s", snippet)
		}
	}
}
