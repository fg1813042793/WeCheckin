package mgr

import (
	"os"
	"strings"
	"testing"
)

func TestAdminLoginHandlerUsesCaptchaContract(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func (h *AdminMgrHandler) AdminCaptcha",
		`c.PostForm("captchaId")`,
		`c.PostForm("captchaCode")`,
		"adminauthservice.LoginWithCaptchaContext(ctx",
		"adminauthservice.GenerateCaptchaContext(ctx)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("captcha handler contract missing %q", snippet)
		}
	}
}

