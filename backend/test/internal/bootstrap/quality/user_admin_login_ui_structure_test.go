package quality_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserManagementUIUsesRoleInsteadOfSeparateAdminLoginSwitch(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "..", "admin", "src", "views", "user", "index.vue"),
		filepath.Join("..", "..", "..", "frontend", "pages", "admin", "user", "admin_user_edit.vue"),
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			"后台登录",
			"登录账号",
			"adminEnabled",
			"onAdminEnabledChange",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must not expose separate admin login switch/account: %s", file, snippet)
			}
		}
		for _, snippet := range []string{
			"绑定角色",
			"roleId",
		} {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must keep role assignment with %s", file, snippet)
			}
		}
	}
}
