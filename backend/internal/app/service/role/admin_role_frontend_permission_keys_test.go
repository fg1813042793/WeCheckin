package role

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleFrontendUsesPermissionKeysForAdminPermissions(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "..", "admin", "src", "views", "role", "index.vue")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read role page: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"adminPermissionKeys",
		"adminApiPermissionKeys",
		"permissionTree",
		"types: 'directory,menu,button'",
		"types: 'api_category,api'",
		"apiTreeRef",
		"form.adminApiPermissionKeys.join(',')",
		`node-key="key"`,
		":default-checked-keys=\"form.adminPermissionKeys\"",
		":default-checked-keys=\"form.adminApiPermissionKeys\"",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role frontend must use permission keys with %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"menuIds",
		"menuTree()",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("role frontend must not use menu IDs snippet %s", forbidden)
		}
	}
}
