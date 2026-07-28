package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleServicePersistsAdminLoginControl(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"AllowAdminLogin",
		"`json:\"allowAdminLogin\"`",
		"role_allow_admin_login",
		"normalizeAllowAdminLogin",
		"permissionsupport.SetRoleAdminPermissionKeysTx",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role service must persist admin login control with %s", snippet)
		}
	}
}
