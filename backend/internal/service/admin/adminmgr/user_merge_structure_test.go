package adminmgr

import (
	"os"
	"strings"
	"testing"
)

func TestManagerServiceUsesMergedUserAccountColumns(t *testing.T) {
	serviceSrc, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	serviceText := string(serviceSrc)
	for _, snippet := range []string{
		"user_role_id",
		"adminaccess.ApplyUserAdminAccessRoleFilter",
		"adminaccess.RoleAllowsAdminAccessContext",
		"user_name",
		"RoleName string `json:\"roleName\"`",
		"Type:",
	} {
		if !strings.Contains(serviceText, snippet) {
			t.Fatalf("service.go must use users table role-login rule %q", snippet)
		}
	}

	passwordSrc, err := os.ReadFile("password.go")
	if err != nil {
		t.Fatalf("read password.go: %v", err)
	}
	passwordText := string(passwordSrc)
	for _, snippet := range []string{"adminLoginRoleFilter", "user_password"} {
		if !strings.Contains(passwordText, snippet) {
			t.Fatalf("password.go must use users table role-login rule %q", snippet)
		}
	}

	for _, item := range []struct {
		file string
		text string
	}{
		{file: "service.go", text: serviceText},
		{file: "password.go", text: passwordText},
	} {
		forbidden := []string{
			"user_admin_enabled",
			"user_account",
			"role_allow_admin_login",
			"AdminEnabled",
			"admin_name",
			"admin_phone",
			"admin_role_id",
			"admin_password",
			"admin_status",
			"admin_add_time",
		}
		for _, snippet := range forbidden {
			if strings.Contains(item.text, snippet) {
				t.Fatalf("%s must not use legacy admin column %q", item.file, snippet)
			}
		}
	}
}
