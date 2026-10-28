package admin

import (
	"os"
	"strings"
	"testing"
)

func TestAdminCaptchaRouteIsPublicAndNoStore(t *testing.T) {
	src, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, snippet := range []string{
		`auth := h.Group("/api/v2/admin/auth", adminmw.NoStore())`,
		`auth.GET("/captcha", aMgr.AdminCaptcha)`,
		`auth.POST("/login", aMgr.AdminLogin)`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("captcha route missing %q", snippet)
		}
	}
}

