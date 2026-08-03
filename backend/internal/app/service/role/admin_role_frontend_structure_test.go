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
		"clientApiTreeData",
		"dingtalkH5ApiTreeData",
		"客户端菜单",
		"钉钉 H5 菜单/按钮",
		"客户端接口",
		"钉钉 H5 接口",
		"payload.clientMenuKeys = form.clientMenuKeys.join(',')",
		"payload.dingtalkH5MenuKeys = form.dingtalkH5MenuKeys.join(',')",
		"payload.clientApiPermissionKeys = form.clientApiPermissionKeys.join(',')",
		"payload.dingtalkH5ApiPermissionKeys = form.dingtalkH5ApiPermissionKeys.join(',')",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin role page must expose app permission tree with %s", snippet)
		}
	}
}

func TestAdminRolePageSplitsMenuAndAPIPermissions(t *testing.T) {
	pageSrc, err := os.ReadFile("../../../../../admin/src/views/role/index.vue")
	if err != nil {
		t.Fatalf("read role page: %v", err)
	}
	text := string(pageSrc)
	for _, snippet := range []string{
		"permission-layout",
		"permission-column permission-column--menu",
		"permission-column permission-column--api",
		"菜单权限",
		"接口权限",
		"后台菜单",
		"后台接口",
		"客户端菜单",
		"客户端接口",
		"钉钉 H5 菜单/按钮",
		"钉钉 H5 接口",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin role page must split menu and api permissions with %s", snippet)
		}
	}
}
