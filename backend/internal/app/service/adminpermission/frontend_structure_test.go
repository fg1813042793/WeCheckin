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
		"api_category",
		"接口类别",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission frontend must use %s", snippet)
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
