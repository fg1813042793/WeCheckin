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
		"用户扩展应用权限",
		"额外授权",
		"额外授权 - 钉钉 H5 菜单/按钮",
		"禁止权限 - 钉钉 H5 菜单/按钮",
		"form.allowPermissionKeys",
		"payload.allowPermissionKeys = form.allowPermissionKeys.join(',')",
		"const dingtalkH5MenuButtonPrefixes = ['dingtalk_h5:menu:', 'dingtalk_h5:button:']",
		"const allowDingTalkH5MenuKeys = computed(() => applicationKeysByPrefixes(form.allowPermissionKeys, dingtalkH5MenuButtonPrefixes))",
		"const denyDingTalkH5MenuKeys = computed(() => applicationKeysByPrefixes(form.denyPermissionKeys, dingtalkH5MenuButtonPrefixes))",
		"...checkedKeys(allowDingTalkH5MenuTreeRef, { includeHalfChecked: true, prefixes: dingtalkH5MenuButtonPrefixes })",
		"...checkedKeys(denyDingTalkH5MenuTreeRef, { includeHalfChecked: true, prefixes: dingtalkH5MenuButtonPrefixes })",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin user page must configure application permissions with %s", snippet)
		}
	}
	for _, legacySnippet := range []string{
		"额外授权 - 钉钉 H5 菜单</div>",
		"禁止权限 - 钉钉 H5 菜单</div>",
	} {
		if strings.Contains(text, legacySnippet) {
			t.Fatalf("admin user page must not label h5 button permissions as menu-only: %s", legacySnippet)
		}
	}
}
