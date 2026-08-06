package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2AdminDingTalkPerfDataRoutesAndPermissions(t *testing.T) {
	routesSrc, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	for _, want := range []string{
		`admin.GET("/dingtalk/perf-reviews", aDingTalk.GetPerfReviews)`,
		`admin.DELETE("/dingtalk/perf-reviews", aDingTalk.DeletePerfReviews)`,
		`admin.GET("/dingtalk/perf-reviews/:id", withQueryID(aDingTalk.GetPerfReviewDetail))`,
		`admin.DELETE("/dingtalk/perf-reviews/:id", withFormID(aDingTalk.DeletePerfReview))`,
		`admin.GET("/dingtalk/perf-histories", aDingTalk.GetPerfHistories)`,
		`admin.DELETE("/dingtalk/perf-histories", aDingTalk.DeletePerfHistories)`,
		`admin.DELETE("/dingtalk/perf-histories/:id", withFormID(aDingTalk.DeletePerfHistory))`,
	} {
		if !strings.Contains(string(routesSrc), want) {
			t.Fatalf("routes_v2.go missing dingtalk performance data route %q", want)
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
		{route: "GET /api/v2/admin/dingtalk/perf-reviews", perm: "dingtalk:perf-reviews:list"},
		{route: "DELETE /api/v2/admin/dingtalk/perf-reviews", perm: "dingtalk:perf-reviews:del"},
		{route: "GET /api/v2/admin/dingtalk/perf-histories", perm: "dingtalk:perf-histories:list"},
		{route: "DELETE /api/v2/admin/dingtalk/perf-histories", perm: "dingtalk:perf-histories:del"},
	} {
		assertStaticAdminRoutePermission(t, permText, item.route, item.perm)
	}
	for _, want := range []string{
		`{method: "GET", path: "/api/v2/admin/dingtalk/perf-reviews/:id", perm: "dingtalk:perf-reviews:detail"}`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/perf-reviews/:id", perm: "dingtalk:perf-reviews:del"}`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/perf-histories/:id", perm: "dingtalk:perf-histories:del"}`,
	} {
		if !strings.Contains(permText, want) {
			t.Fatalf("admin route permissions missing dingtalk performance mapping %q", want)
		}
	}
}

func TestAdminDingTalkPerfDataMenuDeclarationsAndCatalog(t *testing.T) {
	menuSrc, err := os.ReadFile("../internal/support/adminmenuperm/declarations.go")
	if err != nil {
		t.Fatalf("read admin menu declarations: %v", err)
	}
	for _, want := range []string{
		`Key: "admin:menu:dingtalk:perf-reviews"`,
		`Name: "绩效考评单"`,
		`Path: "/dingtalk/perf-reviews"`,
		`Perms: "dingtalk:perf-reviews:list"`,
		`ParentKey: "admin:menu:dingtalk"`,
		`Key: "admin:menu:dingtalk:perf-reviews:list"`,
		`Key: "admin:menu:dingtalk:perf-reviews:detail"`,
		`Key: "admin:menu:dingtalk:perf-reviews:del"`,
		`Key: "admin:menu:dingtalk:perf-histories"`,
		`Name: "绩效流转记录"`,
		`Path: "/dingtalk/perf-histories"`,
		`Perms: "dingtalk:perf-histories:list"`,
		`Key: "admin:menu:dingtalk:perf-histories:list"`,
		`Key: "admin:menu:dingtalk:perf-histories:del"`,
	} {
		if !strings.Contains(string(menuSrc), want) {
			t.Fatalf("admin menu declarations missing dingtalk performance declaration %q", want)
		}
	}

	catalogSrc, err := os.ReadFile("../internal/support/adminrouteperm/catalog.go")
	if err != nil {
		t.Fatalf("read admin route permission catalog: %v", err)
	}
	for _, want := range []string{
		`{"dingtalk:perf-reviews:list", "绩效考评单查看接口"}`,
		`{"dingtalk:perf-reviews:detail", "绩效考评单详情接口"}`,
		`{"dingtalk:perf-reviews:del", "绩效考评单删除接口"}`,
		`{"dingtalk:perf-histories:list", "绩效流转记录查看接口"}`,
		`{"dingtalk:perf-histories:del", "绩效流转记录删除接口"}`,
		`"dingtalk:perf-reviews:list":`,
		`{method: "GET", path: "/api/v2/admin/dingtalk/perf-reviews"}`,
		`"dingtalk:perf-reviews:detail":`,
		`{method: "GET", path: "/api/v2/admin/dingtalk/perf-reviews/:id"}`,
		`"dingtalk:perf-reviews:del":`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/perf-reviews/:id"}`,
		`"dingtalk:perf-histories:list":`,
		`{method: "GET", path: "/api/v2/admin/dingtalk/perf-histories"}`,
		`"dingtalk:perf-histories:del":`,
		`{method: "DELETE", path: "/api/v2/admin/dingtalk/perf-histories/:id"}`,
	} {
		if !strings.Contains(string(catalogSrc), want) {
			t.Fatalf("admin route permission catalog missing dingtalk performance item %q", want)
		}
	}
}
