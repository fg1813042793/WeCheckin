package adminauth

import (
	"os"
	"strings"
	"testing"
)

func TestAdminLoginUsesUsersAndRolePermission(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"`user_mobile` = ?",
		"`user_name` = ?",
		"`user_account` = ?",
		"adminaccess.UserAllowsAdminAccessContext",
		"user_password",
		"RoleName  string",
		"RoleName:  role.Name",
		"user_login_cnt",
		"user_login_time",
		"user_admin_token",
		"user_admin_token_time",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin auth must use merged user account column %q", snippet)
		}
	}
	forbidden := []string{
		"`user_admin_enabled` = 1",
		"role_allow_admin_login",
		"`admin_name` = ?",
		"`admin_password`",
		"`admin_login_cnt`",
		"`admin_token`",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("admin auth must not use legacy admin column %q", snippet)
		}
	}
}
