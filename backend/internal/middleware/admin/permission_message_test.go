package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
)

func TestAdminPermissionMessageIdentifiesFullKeyAndName(t *testing.T) {
	for _, test := range []struct{ required, want string }{
		{"menu:list", "缺少权限-admin:api:menu:list（菜单查看接口）"},
		{"role:edit", "缺少权限-admin:api:role:edit（角色编辑接口）"},
		{"menu:list, role:edit,menu:list", "缺少权限-admin:api:menu:list（菜单查看接口） 或 admin:api:role:edit（角色编辑接口）（满足任一项即可）"},
		{"future:action", "缺少权限-admin:api:future:action"},
	} {
		if got := missingAdminPermissionMessage(test.required); got != test.want {
			t.Fatalf("%q: got %q want %q", test.required, got, test.want)
		}
	}
}

func TestAdminPermissionDeniedResponseNamesActualRoutePermission(t *testing.T) {
	for _, test := range []struct {
		name, path string
		allow      bool
		lookupErr  error
		want       string
	}{
		{"missing", "/api/v2/admin/permissions/tree", false, nil, "缺少权限-admin:api:menu:list（菜单查看接口）"},
		{"allowed", "/api/v2/admin/permissions/tree", true, nil, "allowed"},
		{"lookup failed", "/api/v2/admin/permissions/tree", false, errors.New("private database connection failure"), "权限校验失败，请稍后重试"},
		{"undeclared", "/api/v2/admin/no-such-permission-route", false, nil, "接口未配置权限规则，请联系管理员"},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := server.New()
			h.Use(func(ctx context.Context, c *app.RequestContext) {
				c.Set("admin", &model.Admin{ID: 7, RoleIDs: []uint{9}})
				c.Next(ctx)
			})
			h.Use(adminPermWithLookup(func(context.Context, *gorm.DB, *model.Admin, []uint, string) (bool, error) {
				return test.allow, test.lookupErr
			}))
			h.GET(test.path, func(_ context.Context, c *app.RequestContext) { c.String(200, "allowed") })
			resp := ut.PerformRequest(h.Engine, "GET", test.path, nil).Result()
			if test.allow {
				if string(resp.Body()) != "allowed" {
					t.Fatalf("allowed response=%s", resp.Body())
				}
				return
			}
			var body struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(resp.Body(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != 1 || body.Msg != test.want {
				t.Fatalf("response=%s", resp.Body())
			}
		})
	}
}

func TestAdminPermissionAnyOfStillAllowsOneGrantedPermission(t *testing.T) {
	h := server.New()
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Set("admin", &model.Admin{ID: 7, RoleIDs: []uint{9}})
		c.Next(ctx)
	})
	h.Use(adminPermWithLookup(func(_ context.Context, _ *gorm.DB, _ *model.Admin, _ []uint, key string) (bool, error) {
		return strings.HasSuffix(key, ":upgrade"), nil
	}))
	h.GET("/api/v2/admin/workflow-published-definitions", func(_ context.Context, c *app.RequestContext) { c.String(200, "allowed") })
	resp := ut.PerformRequest(h.Engine, "GET", "/api/v2/admin/workflow-published-definitions", nil).Result()
	if string(resp.Body()) != "allowed" {
		t.Fatalf("OR permission was not preserved: %s", resp.Body())
	}
}
