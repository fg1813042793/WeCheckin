package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleServiceUsesAdminPermissionKeysInsteadOfMenuIDs(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"AdminPermissionKeys",
		"`json:\"adminPermissionKeys\"`",
		"AdminAPIPermissionKeys",
		"`json:\"adminApiPermissionKeys\"`",
		"loadRoleAssignmentMapsContext",
		"permissionsupport.RoleAssignmentMapsContext",
		"permissionsupport.SetRoleAdminPermissionKeysTx",
		"adminAPIPermissionKeys",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role service must use admin permission keys with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"MenuIDs",
		"menuIDs []uint",
		"RoleAdminMenuIDMapContext",
		"SetRoleAdminPermissionsTx",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("role service must not use legacy menu IDs snippet %s", forbidden)
		}
	}
}

func TestRoleServiceInvalidatesRuntimePermissionCacheWhenDeletingRole(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := text
	if start := strings.Index(body, "func DeleteContext"); start >= 0 {
		body = body[start:]
		if end := strings.Index(body, "\n}\n\nfunc "); end >= 0 {
			body = body[:end+3]
		}
	}
	for _, snippet := range []string{
		"Delete(&model.PermissionGrant{})",
		"permissionsupport.InvalidateRuntimePermissionCaches()",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("role delete must clear runtime permission cache with %s", snippet)
		}
	}
}
