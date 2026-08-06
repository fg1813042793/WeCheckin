package admin

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminrouteperm"
)

func TestAdminPermRejectsUnmappedAdminRoute(t *testing.T) {
	h := server.New()
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Set("admin", &model.Admin{ID: 7, Type: 2, RoleID: 9})
		c.Next(ctx)
	})
	h.Use(AdminPerm())
	h.GET("/api/v2/admin/unmapped-route", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "allowed")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/unmapped-route", nil).Result()
	body := string(resp.Body())
	if strings.Contains(body, "allowed") {
		t.Fatalf("unmapped admin route should be rejected, got body %q", body)
	}
	if !strings.Contains(body, "无权限") {
		t.Fatalf("unmapped admin route should return permission error, got body %q", body)
	}
}

func TestAdminPermAllowsExplicitlyPublicAdminRoute(t *testing.T) {
	h := server.New()
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Set("admin", &model.Admin{ID: 7, Type: 2, RoleID: 9})
		c.Next(ctx)
	})
	h.Use(AdminPerm())
	h.GET("/api/v2/admin/me/menus", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "allowed")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/me/menus", nil).Result()
	body := string(resp.Body())
	if body != "allowed" {
		t.Fatalf("explicitly public admin route should be allowed, got body %q", body)
	}
}

func TestRegisteredV2AdminRoutesHavePermissionDeclarations(t *testing.T) {
	routes := registeredV2AdminRoutes(t)
	var missing []string
	for _, route := range routes {
		if _, ok := adminRoutePermission(route.method, route.path); !ok {
			missing = append(missing, route.method+" "+route.path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("registered v2 admin routes missing permission declarations: %s", strings.Join(missing, ", "))
	}
}

func TestAdminRoutePermissionCodesHaveCatalogDeclarations(t *testing.T) {
	declared := map[string]bool{}
	for _, item := range adminrouteperm.Declarations() {
		declared[item.Perms] = true
	}

	required := map[string]bool{}
	for _, perms := range routeMethodPerms {
		for _, code := range permissionCodes(perms) {
			required[code] = true
		}
	}
	for _, route := range routeMethodPermPatterns {
		for _, code := range permissionCodes(route.perm) {
			required[code] = true
		}
	}

	var missing []string
	for code := range required {
		if !declared[code] {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("admin route permission codes missing catalog declarations: %s", strings.Join(missing, ", "))
	}
}

func TestAdminPermResolvesRESTfulV2RouteDeclarations(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   string
	}{
		{method: "GET", path: "/api/v2/admin/users/42", want: "user:list"},
		{method: "PUT", path: "/api/v2/admin/users/42", want: "user:edit"},
		{method: "DELETE", path: "/api/v2/admin/users/42", want: "user:del"},
		{method: "PATCH", path: "/api/v2/admin/users/42/status", want: "user:edit"},
		{method: "GET", path: "/api/v2/admin/surveys/7/responses", want: "response:list"},
		{method: "DELETE", path: "/api/v2/admin/survey-question-bank/12", want: "question-bank:del"},
		{method: "PUT", path: "/api/v2/admin/exams/8", want: "exam:edit"},
		{method: "DELETE", path: "/api/v2/admin/exams/8/records/99", want: "exam:del"},
	}
	for _, tc := range cases {
		got, ok := adminRoutePermission(tc.method, tc.path)
		if !ok {
			t.Fatalf("%s %s should be declared", tc.method, tc.path)
		}
		if got != tc.want {
			t.Fatalf("%s %s permission = %q, want %q", tc.method, tc.path, got, tc.want)
		}
	}
}

func TestAdminPermUsesContextAwarePermissionLookup(t *testing.T) {
	src, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "database.WithContext(ctx)") {
		t.Fatalf("AdminPerm must use request context when loading permissions")
	}
	if strings.Contains(text, "GetAdminPermsContext(ctx, admin)") {
		t.Fatalf("AdminPerm must not fall back to menu permission codes for API access")
	}
}

func TestAdminPermissionMiddlewareKeepsRouteDeclarationsSeparate(t *testing.T) {
	middlewareSrc, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	declarationSrc, err := os.ReadFile("route_permissions.go")
	if err != nil {
		t.Fatalf("read route_permissions.go: %v", err)
	}
	middlewareText := string(middlewareSrc)
	for _, snippet := range []string{
		"var routeMethodPerms",
		"var routeMethodPermPatterns",
	} {
		if strings.Contains(middlewareText, snippet) {
			t.Fatalf("admin permission middleware should not keep route declaration table %s", snippet)
		}
	}
	declarationText := string(declarationSrc)
	for _, snippet := range []string{
		"var routeMethodPerms",
		"var routeMethodPermPatterns",
	} {
		if !strings.Contains(declarationText, snippet) {
			t.Fatalf("admin route declaration file must contain %s", snippet)
		}
	}
}

func TestAdminPermissionRemovesLegacyAdminRouteMap(t *testing.T) {
	src, err := os.ReadFile("route_permissions.go")
	if err != nil {
		t.Fatalf("read route_permissions.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"var routePerms",
		`"/admin/`,
		"routePerms[path]",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("legacy /admin permission declaration must be removed: %s", forbidden)
		}
	}
	if _, ok := adminRoutePermission("GET", "/admin/user/menus"); ok {
		t.Fatalf("legacy /admin route must not be declared after v2 migration")
	}
}

func TestAdminPermissionMiddlewareRemovesUnusedPermissionMatcher(t *testing.T) {
	src, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	if strings.Contains(string(src), "func permissionMatches") {
		t.Fatalf("unused permissionMatches helper should be removed")
	}
}

func TestAdminPermAuditsDeniedRequests(t *testing.T) {
	src, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	if !strings.Contains(string(src), "auditAdminPermissionDenied(") {
		t.Fatalf("AdminPerm must audit denied permission checks")
	}
}

func TestAuditAdminPermissionDeniedToleratesNilLogger(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("auditAdminPermissionDenied should tolerate nil logger, panic=%v", r)
		}
	}()
	auditAdminPermissionDenied(&model.Admin{ID: 1, RoleID: 2}, "/admin/x", "x:list", "test")
}

type registeredRoute struct {
	method string
	path   string
}

func registeredV2AdminRoutes(t *testing.T) []registeredRoute {
	t.Helper()
	file := filepath.Join("..", "..", "..", "cmd", "routes_v2.go")
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	re := regexp.MustCompile(`admin\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
	var out []registeredRoute
	for _, match := range re.FindAllStringSubmatch(string(src), -1) {
		out = append(out, registeredRoute{method: match[1], path: "/api/v2/admin" + match[2]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].path == out[j].path {
			return out[i].method < out[j].method
		}
		return out[i].path < out[j].path
	})
	return out
}
