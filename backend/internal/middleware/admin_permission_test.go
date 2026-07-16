package middleware

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
	"wecheckin-backend/backend/internal/model"
)

func TestAdminPermRejectsUnmappedAdminRoute(t *testing.T) {
	h := server.New()
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Set("admin", &model.Admin{ID: 7, Type: 2, RoleID: 9})
		c.Next(ctx)
	})
	h.Use(AdminPerm())
	h.GET("/admin/unmapped_route", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "allowed")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/admin/unmapped_route", nil).Result()
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
	h.GET("/admin/user/menus", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "allowed")
	})

	resp := ut.PerformRequest(h.Engine, "GET", "/admin/user/menus", nil).Result()
	body := string(resp.Body())
	if body != "allowed" {
		t.Fatalf("explicitly public admin route should be allowed, got body %q", body)
	}
}

func TestRegisteredAdminRoutesHavePermissionDeclarations(t *testing.T) {
	routes := registeredAdminRoutes(t)
	var missing []string
	for _, route := range routes {
		if _, ok := routePerms[route]; !ok {
			missing = append(missing, route)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("registered admin routes missing permission declarations: %s", strings.Join(missing, ", "))
	}
}

func TestAdminPermUsesContextAwarePermissionLookup(t *testing.T) {
	src, err := os.ReadFile("admin_permission.go")
	if err != nil {
		t.Fatalf("read admin_permission.go: %v", err)
	}
	if !strings.Contains(string(src), "GetAdminPermsContext(ctx, admin)") {
		t.Fatalf("AdminPerm must use request context when loading permissions")
	}
}

func TestAdminPermAuditsDeniedRequests(t *testing.T) {
	src, err := os.ReadFile("admin_permission.go")
	if err != nil {
		t.Fatalf("read admin_permission.go: %v", err)
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

func registeredAdminRoutes(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "cmd", "routes_*.go"))
	if err != nil {
		t.Fatalf("glob routes: %v", err)
	}
	re := regexp.MustCompile(`adminGroup\.(?:GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
	seen := map[string]struct{}{}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, match := range re.FindAllStringSubmatch(string(src), -1) {
			seen["/admin"+match[1]] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for route := range seen {
		out = append(out, route)
	}
	sort.Strings(out)
	return out
}
