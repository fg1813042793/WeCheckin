package adminpermission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminPermissionFrontendUsesPermissionAPIs(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..", "admin", "src")
	page, err := os.ReadFile(filepath.Join(root, "views", "menu", "index.vue"))
	if err != nil {
		t.Fatalf("read menu page: %v", err)
	}
	text := string(page)
	for _, snippet := range []string{
		"权限管理",
		"permissionTree",
		"permissionAdd",
		"permissionEdit",
		"permissionDel",
		"permissionKey",
		"parentKey",
		"permissionScope",
		"activePermissionScope",
		"types: activePermissionScope.value.types",
		"value: 'admin_api'",
		"platform: ''",
		"types: 'api_category,api'",
		"defaultPlatform: 'admin'",
		"api_category",
		"接口类别",
		"platform: 'client', types: 'menu'",
		"platform: 'dingtalk_h5', types: 'directory,menu,button'",
		"permissionPlatformFromKey(parentKey)",
		"scopeValueForPermission(form.platform, form.type)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission frontend must use %s", snippet)
		}
	}
	for _, forbidden := range []string{
		"{ label: '客户端', value: 'client', platform: 'client', types: '', defaultType: 'menu' }",
		"{ label: '钉钉 H5', value: 'dingtalk_h5', platform: 'dingtalk_h5', types: '', defaultType: 'menu' }",
		"platform: 'dingtalk_h5', types: 'menu'",
		"if (platform === 'admin' && (type === 'api_category' || type === 'api')) return 'admin_api'",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission frontend must not keep API permissions mixed into app menu scope with %s", forbidden)
		}
	}
	api, err := os.ReadFile(filepath.Join(root, "api", "index.ts"))
	if err != nil {
		t.Fatalf("read admin api: %v", err)
	}
	for _, snippet := range []string{
		"permissionTree",
		"types?: string",
		"permissionList",
		"permissionAdd",
		"permissionEdit",
		"permissionDel",
		"/permissions/tree",
		"/permissions",
	} {
		if !strings.Contains(string(api), snippet) {
			t.Fatalf("admin api must expose permission method %s", snippet)
		}
	}
}

func TestAdminPermissionFrontendUsesStandardAdminToolbarLayout(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..", "admin", "src")
	page, err := os.ReadFile(filepath.Join(root, "views", "menu", "index.vue"))
	if err != nil {
		t.Fatalf("read menu page: %v", err)
	}
	text := string(page)
	for _, snippet := range []string{
		`class="admin-toolbar"`,
		`class="admin-toolbar__left"`,
		`class="admin-toolbar__right"`,
		`class="admin-table-actions"`,
		`style="width:100%"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission frontend should use standard admin layout snippet %s", snippet)
		}
	}
	if strings.Count(text, `class="admin-toolbar"`) < 2 {
		t.Fatalf("permission frontend should split search and actions into separate standard toolbars")
	}
	if strings.Contains(text, "admin-page__title") {
		t.Fatalf("permission frontend should not use standalone admin-page__title inside the card")
	}
	if !strings.Contains(text, `<el-button v-if="hasPerm('admin:menu:permission:add')" type="success" @click="showAdd('')">新增权限</el-button>`) {
		t.Fatalf("permission frontend root add button should use the standard green success style")
	}
}

func TestAdminPermissionFrontendAllowsEditingPermissionKey(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..", "admin", "src")
	page, err := os.ReadFile(filepath.Join(root, "views", "menu", "index.vue"))
	if err != nil {
		t.Fatalf("read menu page: %v", err)
	}
	text := string(page)
	for _, forbidden := range []string{
		`:disabled="!dialog.isCreate"`,
		"req.Key = key",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission edit form must not lock permission key with %s", forbidden)
		}
	}
	for _, snippet := range []string{
		"originalPermissionKey",
		"originalKey: originalPermissionKey.value",
		"await adminApi.permissionEdit({ ...payload, originalKey: originalPermissionKey.value })",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission edit form must preserve old key while submitting edited key with %s", snippet)
		}
	}

	api, err := os.ReadFile(filepath.Join(root, "api", "index.ts"))
	if err != nil {
		t.Fatalf("read admin api: %v", err)
	}
	for _, snippet := range []string{
		"originalKey?: ID",
		"const key = data.originalKey ?? data.key ?? data.permissionKey",
	} {
		if !strings.Contains(string(api), snippet) {
			t.Fatalf("permission edit api must address old key while submitting new key with %s", snippet)
		}
	}
}

func TestAdminPermissionFrontendDoesNotExposeLegacyPerms(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..", "admin", "src")
	page, err := os.ReadFile(filepath.Join(root, "views", "menu", "index.vue"))
	if err != nil {
		t.Fatalf("read menu page: %v", err)
	}
	text := string(page)
	for _, forbidden := range []string{
		"兼容标识",
		`prop="perms"`,
		`v-model="form.perms"`,
		"perms: form.perms",
		"item.perms",
		"form.perms",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("permission frontend must not expose legacy perms field with %s", forbidden)
		}
	}
}
