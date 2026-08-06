package adminmgr

import (
	"os"
	"strings"
	"testing"
)

func TestManagerListAvoidsPerRowRoleAndDeptQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	forbidden := []string{
		"db.First(&role, a.RoleID)",
		"access.AdminDeptIDs(a.ID)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("manager list must avoid per-row query snippet %q", snippet)
		}
	}

	required := []string{
		"loadRoleNameMapContext(ctx, db, list)",
		"loadAdminDeptIDMapContext(ctx, db, list)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("manager list must batch load related data with %q", snippet)
		}
	}
}

func TestManagerListUsesUsersTableAndLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"var adminManagerListColumns = []string{",
		"db.Model(&model.User{})",
		"`user_role_id` > 0",
		"Select(adminManagerListColumns)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("manager list should query users with lightweight columns using %q", snippet)
		}
	}
}
