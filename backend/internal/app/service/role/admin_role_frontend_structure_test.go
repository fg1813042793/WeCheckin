package role

import (
	"os"
	"strings"
	"testing"
)

func TestAdminRolePageShowsApplicationPermissionTrees(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../admin/src/views/role/index.vue")
	if err != nil {
		t.Fatalf("read role page: %v", err)
	}
	apiSrc, err := os.ReadFile("../../../../../admin/src/api/index.ts")
	if err != nil {
		t.Fatalf("read api: %v", err)
	}
	text := string(pageSrc) + string(apiSrc)
	for _, snippet := range []string{
		"appPermissionTree()",
		"clientMenuTreeData",
		"dingtalkH5MenuTreeData",
		"客户端权限",
		"钉钉 H5 权限",
		"payload.clientMenuKeys = form.clientMenuKeys.join(',')",
		"payload.dingtalkH5MenuKeys = form.dingtalkH5MenuKeys.join(',')",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin role page must expose app permission tree with %s", snippet)
		}
	}
}
