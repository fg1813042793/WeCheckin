package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleServicePersistsApplicationMenuPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"ClientMenuKeys",
		"`json:\"clientMenuKeys\"`",
		"DingTalkH5MenuKeys",
		"`json:\"dingtalkH5MenuKeys\"`",
		"loadRoleAssignmentMapsContext(ctx, db, list)",
		"permissionsupport.RoleAssignmentMapsContext(ctx, db, roleIDs)",
		"permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, role.ID, clientMenuKeys, dingtalkH5MenuKeys)",
		"permissionsupport.SetRoleApplicationMenuPermissionsTx(tx, id, clientMenuKeys, dingtalkH5MenuKeys)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role service must persist app menu permissions with %s", snippet)
		}
	}
}
