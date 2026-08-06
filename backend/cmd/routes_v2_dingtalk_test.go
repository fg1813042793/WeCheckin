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
		`group.GET("/public-config", handler.Auth.PublicConfig)`,
		`group.POST("/login", handler.Auth.Login)`,
		`group.POST("/bind-self", handler.Auth.BindSelf)`,
		`auth.GET("/bootstrap", handler.Bootstrap.Bootstrap)`,
		`auth.GET("/reviews/export", handler.Review.ExportReviews)`,
		`auth.POST("/reviews/:id/submit-self", withBodyOrFormParam("id", "id", handler.Review.SubmitSelf))`,
		`auth.POST("/reviews/:id/finalize", withBodyOrFormParam("id", "id", handler.Review.Finalize))`,
		`auth.GET("/template", handler.Template.Template)`,
		`auth.PUT("/template", handler.Template.SaveTemplate)`,
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
		`authed := group.Group("", dingtalkh5mw.DingTalkH5Auth())`,
		`authed.POST("/logout", handler.Auth.Logout)`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk h5 logout should require auth but bypass ApiPerm, missing %s", want)
		}
	}
	if strings.Contains(text, `auth.POST("/logout", handler.Auth.Logout)`) {
		t.Fatalf("dingtalk h5 logout must not be blocked by business API permission middleware")
	}
}

func TestV2DingTalkH5SSOLoginBypassesExistingSessionAuth(t *testing.T) {
	src, err := os.ReadFile("routes_v2_dingtalk.go")
	if err != nil {
		t.Fatalf("read routes_v2_dingtalk.go: %v", err)
	}
	text := string(src)
	route := `group.POST("/sso-login", handler.Auth.SSOLogin)`
	authGroup := `authed := group.Group("", dingtalkh5mw.DingTalkH5Auth())`
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
