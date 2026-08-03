package main

import (
	"os"
	"strings"
	"testing"
)

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

	permSrc, err := os.ReadFile("../internal/middleware/admin_route_permissions.go")
	if err != nil {
		t.Fatalf("read admin_route_permissions.go: %v", err)
	}
	for _, want := range []string{
		`"GET /api/v2/admin/dingtalk/user-bindings":             "dingtalk:bindings:list"`,
		`"POST /api/v2/admin/dingtalk/user-bindings":            "dingtalk:bindings:edit"`,
		`{method: "PATCH", path: "/api/v2/admin/dingtalk/user-bindings/:id/status", perm: "dingtalk:bindings:edit"}`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/user-bindings/:id", perm: "dingtalk:bindings:edit"}`,
	} {
		if !strings.Contains(string(permSrc), want) {
			t.Fatalf("admin route permissions missing dingtalk binding mapping %q", want)
		}
	}
	if strings.Contains(string(permSrc), `"GET /api/v2/admin/dingtalk/user-bindings":             "dingtalk:config"`) {
		t.Fatalf("dingtalk binding list route must not reuse broad dingtalk:config permission")
	}
}

func TestAdminDingTalkUserBindingMenuDeclarationsAndMigration(t *testing.T) {
	menuSrc, err := os.ReadFile("../internal/app/support/adminmenuperm/declarations.go")
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
	permSrc, err := os.ReadFile("../internal/middleware/admin_route_permissions.go")
	if err != nil {
		t.Fatalf("read admin_route_permissions.go: %v", err)
	}
	for _, want := range []string{
		`"GET /api/v2/admin/dingtalk/settings":                  "dingtalk:settings:list"`,
		`"PUT /api/v2/admin/dingtalk/settings":                  "dingtalk:settings:edit"`,
	} {
		if !strings.Contains(string(permSrc), want) {
			t.Fatalf("admin settings route permissions missing split mapping %q", want)
		}
	}
	if strings.Contains(string(permSrc), `"GET /api/v2/admin/dingtalk/settings":                  "dingtalk:config"`) {
		t.Fatalf("dingtalk settings routes must not reuse broad dingtalk:config permission")
	}
}

func TestAdminDingTalkSettingsMenuDeclarationsExposeTableControls(t *testing.T) {
	menuSrc, err := os.ReadFile("../internal/app/support/adminmenuperm/declarations.go")
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
	} {
		if !strings.Contains(string(menuSrc), want) {
			t.Fatalf("admin menu declarations missing dingtalk settings control %q", want)
		}
	}
}
