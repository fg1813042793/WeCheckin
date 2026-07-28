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
		"loadRoleAdminPermissionKeyMapContext",
		"loadRoleAdminAPIPermissionKeyMapContext",
		"permissionsupport.RoleAdminPermissionKeyMapContext",
		"permissionsupport.RoleAdminAPIPermissionKeyMapContext",
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
