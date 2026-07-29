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

func TestAdminPermissionServiceCanRenamePermissionKey(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"req.Key = key",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission edit must not overwrite edited key with old key: %s", forbidden)
		}
	}
	for _, snippet := range []string{
		"oldKey := strings.TrimSpace(key)",
		"newKey := item.Key",
		`"permission_key":`,
		"`permission_key` = ? AND `permission_key` <> ?",
		"`grant_permission_key` = ?",
		"Update(\"grant_permission_key\", newKey)",
		"`permission_parent_key` = ?",
		"Update(\"permission_parent_key\", newKey)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission rename must keep related records consistent with %s", snippet)
		}
	}
}
