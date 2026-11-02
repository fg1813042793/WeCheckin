package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5CaptchaRouteIsPublicAndDoesNotAlterSSOOrBinding(t *testing.T) {
	routesSource, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	routesText := string(routesSource)
	for _, snippet := range []string{
		`group.GET("/captcha", handler.Auth.Captcha)`,
		`group.POST("/login", handler.Auth.Login)`,
		`group.POST("/sso-login", handler.Auth.SSOLogin)`,
		`group.POST("/bind-self", handler.Auth.BindSelf)`,
	} {
		if !strings.Contains(routesText, snippet) {
			t.Fatalf("DingTalk H5 routes missing %q", snippet)
		}
	}

	handlerSource, err := os.ReadFile("../../../handler/dingtalkh5/auth/handler.go")
	if err != nil {
		t.Fatal(err)
	}
	handlerText := string(handlerSource)
	loginBody := routeFunctionBody(handlerText, "func (h *Handler) Login")
	if !strings.Contains(loginBody, "LoginWithCaptchaContext") {
		t.Fatal("password login must use captcha-aware service")
	}
	for _, name := range []string{"func (h *Handler) SSOLogin", "func (h *Handler) BindSelf"} {
		body := routeFunctionBody(handlerText, name)
		if strings.Contains(body, "Captcha") || strings.Contains(body, "captcha") {
			t.Fatalf("%s must remain captcha-free", name)
		}
	}
	if !strings.Contains(routeFunctionBody(handlerText, "func (h *Handler) Captcha"), `c.Header("Cache-Control", "no-store")`) {
		t.Fatal("captcha response must disable caching")
	}
}

func routeFunctionBody(source, signature string) string {
	start := strings.Index(source, signature)
	if start < 0 {
		return ""
	}
	rest := source[start:]
	next := strings.Index(rest[len(signature):], "\nfunc ")
	if next < 0 {
		return rest
	}
	return rest[:len(signature)+next]
}
