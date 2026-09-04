package dingtalk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkUserBindingAdminHandlerStructure(t *testing.T) {
	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob handler files: %v", err)
	}
	var builder strings.Builder
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		builder.Write(src)
	}
	text := builder.String()
	for _, want := range []string{
		"GetUserBindings",
		"SaveUserBinding",
		"StatusUserBinding",
		"DeleteUserBinding",
		"h.service.ListUserBindings",
		"h.service.SaveUserBinding",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk binding admin handler missing %q", want)
		}
	}
	if strings.Contains(text, `"appSecret":`) {
		t.Fatalf("dingtalk binding admin handler must not return app secrets")
	}
}

func TestDingTalkUserBindingAdminFrontendStructure(t *testing.T) {
	routeSrc, err := os.ReadFile("../../../../../admin/src/router/adminRoutes.ts")
	if err != nil {
		t.Fatalf("read admin routes: %v", err)
	}
	for _, want := range []string{
		`path: 'dingtalk/bindings'`,
		`name: 'DingTalkBindings'`,
		`../views/dingtalk-bindings/index.vue`,
	} {
		if !strings.Contains(string(routeSrc), want) {
			t.Fatalf("admin routes missing dingtalk binding route %q", want)
		}
	}

	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-bindings/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk binding page: %v", err)
	}
	for _, want := range []string{
		"钉钉用户绑定管理",
		"/api/v2/admin/dingtalk/user-bindings",
		"dingTalkUserId",
		"userId",
		"admin:menu:dingtalk:bindings:edit",
		"admin-toolbar",
		"admin-pagination",
	} {
		if !strings.Contains(string(viewSrc), want) {
			t.Fatalf("dingtalk binding page missing %q", want)
		}
	}
	if strings.Contains(string(viewSrc), "admin-card__title") {
		t.Fatalf("dingtalk binding page should use the standard admin toolbar layout instead of a custom card header")
	}
	for _, forbidden := range []string{
		"admin:menu:dingtalk:config:edit",
		"admin:api:dingtalk:config",
	} {
		if strings.Contains(string(viewSrc), forbidden) {
			t.Fatalf("dingtalk binding page should not inherit broad config permission %q", forbidden)
		}
	}
}

func TestDingTalkUserBindingLocalUserPickerUsesDeptTreeSingleSelect(t *testing.T) {
	serviceSrc, err := os.ReadFile("../../../service/admin/dingtalk/bindings.go")
	if err != nil {
		t.Fatalf("read dingtalk binding service: %v", err)
	}
	serviceText := string(serviceSrc)
	for _, want := range []string{
		"type UserBindingTreeNode struct",
		"buildBindingUserTree",
		"`user_depts`",
		"`departments`",
		`UserTreeOptions: userTreeOptions`,
	} {
		if !strings.Contains(serviceText, want) {
			t.Fatalf("dingtalk binding service should expose local users as dept tree, missing %q", want)
		}
	}

	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-bindings/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk binding page: %v", err)
	}
	viewText := string(viewSrc)
	for _, want := range []string{
		"<el-tree-select",
		`:data="userTreeOptions"`,
		`:props="userTreeProps"`,
		"check-strictly",
		`:filter-node-method="filterUserTreeNode"`,
		"function handleLocalUserChange",
	} {
		if !strings.Contains(viewText, want) {
			t.Fatalf("dingtalk binding page should use dept tree single-select user picker, missing %q", want)
		}
	}
	if strings.Contains(viewText, `<el-option
              v-for="user in userOptions"`) {
		t.Fatalf("dingtalk binding page must not render local users as a flat select")
	}
}
