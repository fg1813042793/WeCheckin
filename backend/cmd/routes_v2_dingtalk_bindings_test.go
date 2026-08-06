package main

import (
	"os"
	"strings"
	"testing"
)

func assertStaticAdminRoutePermission(t *testing.T, permSrc, route, perm string) {
	t.Helper()
	if !containsStaticAdminRoutePermission(permSrc, route, perm) {
		t.Fatalf("admin route permissions missing mapping %s -> %s", route, perm)
	}
}

func containsStaticAdminRoutePermission(permSrc, route, perm string) bool {
	routeText := `"` + route + `"`
	permText := `"` + perm + `"`
	for _, line := range strings.Split(permSrc, "\n") {
		if strings.Contains(line, routeText) && strings.Contains(line, permText) {
			return true
		}
	}
	return false
}

func TestV2AdminDingTalkUserBindingRoutesAndPermissions(t *testing.T) {
	routesSrc, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	for _, want := range []string{
		`admin.GET("/dingtalk/user-bindings", aDingTalk.GetUserBindings)`,
		`admin.POST("/dingtalk/user-bindings", aDingTalk.SaveUserBinding)`,
		`admin.PATCH("/dingtalk/user-bindings/:id/status", withFormID(aDingTalk.StatusUserBinding))`,
		`admin.DELETE("/dingtalk/user-bindings/:id", withFormID(aDingTalk.DeleteUserBinding))`,
	} {
		if !strings.Contains(string(routesSrc), want) {
			t.Fatalf("routes_v2.go missing dingtalk binding route %q", want)
		}
	}

	permSrc, err := os.ReadFile("../internal/middleware/admin/route_permissions.go")
	if err != nil {
		t.Fatalf("read admin/route_permissions.go: %v", err)
	}
	permText := string(permSrc)
	for _, item := range []struct {
		route string
		perm  string
	}{
		{route: "GET /api/v2/admin/dingtalk/user-bindings", perm: "dingtalk:bindings:list"},
		{route: "POST /api/v2/admin/dingtalk/user-bindings", perm: "dingtalk:bindings:edit"},
	} {
		assertStaticAdminRoutePermission(t, permText, item.route, item.perm)
	}
	for _, want := range []string{
		`{method: "PATCH", path: "/api/v2/admin/dingtalk/user-bindings/:id/status", perm: "dingtalk:bindings:edit"}`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/user-bindings/:id", perm: "dingtalk:bindings:edit"}`,
	} {
		if !strings.Contains(permText, want) {
			t.Fatalf("admin route permissions missing dingtalk binding mapping %q", want)
		}
	}
	if containsStaticAdminRoutePermission(permText, "GET /api/v2/admin/dingtalk/user-bindings", "dingtalk:config") {
		t.Fatalf("dingtalk binding list route must not reuse broad dingtalk:config permission")
	}
}

func TestAdminDingTalkUserBindingMenuDeclarationsAndMigration(t *testing.T) {
	menuSrc, err := os.ReadFile("../internal/support/adminmenuperm/declarations.go")
	if err != nil {
		t.Fatalf("read admin menu declarations: %v", err)
	}
	for _, want := range []string{
		`Key: "admin:menu:dingtalk:bindings"`,
		`Name: "用户绑定管理"`,
		`Path: "/dingtalk/bindings"`,
		`Perms: "dingtalk:bindings:list"`,
		`ParentKey: "admin:menu:dingtalk"`,
		`Key: "admin:menu:dingtalk:bindings:list"`,
		`Key: "admin:menu:dingtalk:bindings:edit"`,
		`Perms: "dingtalk:bindings:edit"`,
	} {
		if !strings.Contains(string(menuSrc), want) {
			t.Fatalf("admin menu declarations missing dingtalk binding declaration %q", want)
		}
	}

	migrationSrc, err := os.ReadFile("../migrations/20260801103000_split_admin_dingtalk_permissions.sql")
	if err != nil {
		t.Fatalf("read dingtalk split permission migration: %v", err)
	}
	for _, want := range []string{
		"admin:menu:dingtalk:bindings",
		"用户绑定管理",
		"/dingtalk/bindings",
		"admin:menu:dingtalk:bindings:list",
		"admin:menu:dingtalk:bindings:edit",
		"dingtalk:bindings:list",
		"dingtalk:bindings:edit",
		"setup-dingtalk-split-backfill",
	} {
		if !strings.Contains(string(migrationSrc), want) {
			t.Fatalf("dingtalk binding migration missing %q", want)
		}
	}
}

func TestV2AdminDingTalkSettingsRoutesUseSplitPermissions(t *testing.T) {
	permSrc, err := os.ReadFile("../internal/middleware/admin/route_permissions.go")
	if err != nil {
		t.Fatalf("read admin/route_permissions.go: %v", err)
	}
	permText := string(permSrc)
	for _, item := range []struct {
		route string
		perm  string
	}{
		{route: "GET /api/v2/admin/dingtalk/settings", perm: "dingtalk:settings:list"},
		{route: "PUT /api/v2/admin/dingtalk/settings", perm: "dingtalk:settings:edit"},
		{route: "POST /api/v2/admin/dingtalk/settings/notification-test", perm: "dingtalk:settings:edit"},
	} {
		assertStaticAdminRoutePermission(t, permText, item.route, item.perm)
	}
	if containsStaticAdminRoutePermission(permText, "GET /api/v2/admin/dingtalk/settings", "dingtalk:config") {
		t.Fatalf("dingtalk settings routes must not reuse broad dingtalk:config permission")
	}
}

func TestAdminDingTalkSettingsMenuDeclarationsExposeTableControls(t *testing.T) {
	menuSrc, err := os.ReadFile("../internal/support/adminmenuperm/declarations.go")
	if err != nil {
		t.Fatalf("read admin menu declarations: %v", err)
	}
	for _, want := range []string{
		`Key: "admin:menu:dingtalk:config"`,
		`Perms: "dingtalk:settings:list"`,
		`Key: "admin:menu:dingtalk:config:list"`,
		`Name: "钉钉配置查看"`,
		`Perms: "dingtalk:settings:list"`,
		`Key: "admin:menu:dingtalk:config:edit"`,
		`Name: "钉钉配置保存"`,
		`Perms: "dingtalk:settings:edit"`,
		`Key: "admin:menu:dingtalk:config:test"`,
		`Name: "钉钉通知测试"`,
	} {
		if !strings.Contains(string(menuSrc), want) {
			t.Fatalf("admin menu declarations missing dingtalk settings control %q", want)
		}
	}
}

func TestV2AdminDingTalkNotificationDiagnosisRoute(t *testing.T) {
	routesSrc, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	routeText := string(routesSrc)
	if !strings.Contains(routeText, `admin.POST("/dingtalk/settings/notification-test", aDingTalk.TestNotification)`) {
		t.Fatalf("routes_v2.go should register dingtalk notification diagnosis route")
	}

	swaggerSrc, err := os.ReadFile("routes_v2_swagger.go")
	if err != nil {
		t.Fatalf("read routes_v2_swagger.go: %v", err)
	}
	if !strings.Contains(string(swaggerSrc), `@Router /api/v2/admin/dingtalk/settings/notification-test [post]`) {
		t.Fatalf("swagger should document dingtalk notification diagnosis route")
	}
}
