package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleServicePersistsApplicationAPIPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"ClientAPIPermissionKeys",
		"`json:\"clientApiPermissionKeys\"`",
		"DingTalkH5APIPermissionKeys",
		"`json:\"dingtalkH5ApiPermissionKeys\"`",
		"ClientAPI",
		"DingTalkH5API",
		"loadRoleApplicationAPIKeyMapContext(ctx, db, list)",
		"permissionsupport.RoleApplicationAPIKeyMapContext(ctx, db, roleIDs)",
		"permissionsupport.SetRoleApplicationAPIPermissionsTx(tx, role.ID, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)",
		"permissionsupport.SetRoleApplicationAPIPermissionsTx(tx, id, clientAPIPermissionKeys, dingtalkH5APIPermissionKeys)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role service must persist app API permissions with %s", snippet)
		}
	}
}
