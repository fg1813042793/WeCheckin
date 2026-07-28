package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleListAvoidsPerRowMenuAndDeptQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	forbidden := []string{
		"menuservice.GetRoleMenuIDs(r.ID)",
		"access.RoleDeptIDs(r.ID)",
		"access.AdminDeptIDs(admin.ID)",
		"access.RoleDeptIDs(admin.RoleID)",
		"model.RoleMenu",
		"model.RoleDept",
		"menuservice.SetRoleMenusTx",
		"access.SetRoleDeptsTx",
		"`role_depts`",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("role list must avoid uncached/per-row query snippet %q", snippet)
		}
	}

	required := []string{
		"loadRoleAdminPermissionKeyMapContext(ctx, db, list)",
		"loadRoleDeptIDMapContext(ctx, db, list)",
		"access.AdminDeptIDsContext(ctx, admin.ID)",
		"permissionsupport.RoleAdminPermissionKeyMapContext(ctx, db, roleIDs)",
		"permissionsupport.RoleCustomDeptIDMapContext(ctx, db, roleIDs)",
		"permissionsupport.RoleCustomDeptIDsContext(ctx, db, admin.RoleID)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role list must use batched/context-aware query with %q", snippet)
		}
	}
}
