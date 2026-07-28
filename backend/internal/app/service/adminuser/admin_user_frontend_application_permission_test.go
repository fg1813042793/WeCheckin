package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestAdminUserPageCanConfigureApplicationPermissions(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../admin/src/views/user/index.vue")
	if err != nil {
		t.Fatalf("read user page: %v", err)
	}
	text := string(pageSrc)
	for _, snippet := range []string{
		"loadApplicationPermissionTree",
		"clientMenuTreeData",
		"dingtalkH5MenuTreeData",
		"用户应用权限",
		"额外授权",
		"form.allowPermissionKeys",
		"payload.allowPermissionKeys = form.allowPermissionKeys.join(',')",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin user page must configure application permissions with %s", snippet)
		}
	}
}
