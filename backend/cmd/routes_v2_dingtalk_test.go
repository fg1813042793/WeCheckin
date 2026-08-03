package main

import (
	"os"
	"strings"
	"testing"
)

func TestV2DingTalkH5RoutesAreIsolated(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	required := []string{
		`h.Group("/api/v2/dingtalk/h5")`,
		`group.GET("/public-config", handler.PublicConfig)`,
		`group.POST("/login", handler.Login)`,
		`group.POST("/bind-self", handler.BindSelf)`,
		`auth.GET("/bootstrap", handler.Bootstrap)`,
		`auth.GET("/reviews/export", handler.ExportReviews)`,
		`auth.POST("/reviews/:id/submit-self", withBodyOrFormParam("id", "id", handler.SubmitSelf))`,
		`auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Finalize))`,
		`auth.GET("/template", handler.Template)`,
		`auth.PUT("/template", handler.SaveTemplate)`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2_dingtalk.go missing %s", want)
		}
	}
	if strings.Contains(text, "AdminAuth") || strings.Contains(text, "ClientAuth") {
		t.Fatalf("dingtalk h5 routes must use the module token guard, not existing admin/client middleware")
	}
}

func TestV2DingTalkH5LogoutBypassesBusinessAPIPermission(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		`authed := group.Group("", handler.Auth())`,
		`authed.POST("/logout", handler.Logout)`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk h5 logout should require auth but bypass ApiPerm, missing %s", want)
		}
	}
	if strings.Contains(text, `auth.POST("/logout", handler.Logout)`) {
		t.Fatalf("dingtalk h5 logout must not be blocked by business API permission middleware")
	}
}

func TestV2DingTalkH5SSOLoginBypassesExistingSessionAuth(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	route := `group.POST("/sso-login", handler.SSOLogin)`
	authGroup := `authed := group.Group("", handler.Auth())`
	routeIndex := strings.Index(text, route)
	authGroupIndex := strings.Index(text, authGroup)
	if routeIndex < 0 {
		t.Fatalf("dingtalk h5 routes missing %s", route)
	}
	if authGroupIndex < 0 {
		t.Fatalf("dingtalk h5 routes missing %s", authGroup)
	}
	if routeIndex > authGroupIndex {
		t.Fatalf("sso-login must be registered before authenticated groups")
	}
}

func TestV2RouteSuiteRegistersDingTalkH5Routes(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	if !strings.Contains(string(src), "registerV2DingTalkH5Routes(h)") {
		t.Fatalf("registerV2Routes must register dingtalk h5 route suite")
	}
}
