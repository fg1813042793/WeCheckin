package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleHandlerAcceptsAdminPermissionKeys(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`c.PostForm("adminPermissionKeys")`,
		`c.PostForm("adminApiPermissionKeys")`,
		"adminPermissionKeys := parsePermissionKeys",
		"adminAPIPermissionKeys := parsePermissionKeys",
		"AddWithAssignmentsContext(ctx, name, remark, c.ClientIP(), sort, dataScope, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys",
		"EditWithAssignmentsContext(ctx, uint(id), name, remark, c.ClientIP(), sort, status, dataScope, allowAdminLogin, adminPermissionKeys, adminAPIPermissionKeys",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role handler must accept admin permission keys with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		`c.PostForm("menuIds")`,
		"menuIDs",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("role handler must not parse legacy menu IDs snippet %s", forbidden)
		}
	}
}
