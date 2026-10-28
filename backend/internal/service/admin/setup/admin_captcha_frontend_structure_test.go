package setup

import (
	"os"
	"strings"
	"testing"
)

func TestAdminCaptchaFrontendAndSetupSwitch(t *testing.T) {
	loginPage, err := os.ReadFile("../../../../../admin/src/views/login/index.vue")
	if err != nil {
		t.Fatal(err)
	}
	setupPage, err := os.ReadFile("../../../../../admin/src/views/setup/index.vue")
	if err != nil {
		t.Fatal(err)
	}
	api, err := os.ReadFile("../../../../../admin/src/api/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(loginPage) + string(setupPage) + string(api)
	for _, snippet := range []string{
		"ADMIN_LOGIN_CAPTCHA_ENABLED",
		"captchaId",
		"captchaCode",
		"captchaImage",
		"refreshCaptcha",
		"adminApi.captcha()",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("admin captcha frontend missing %q", snippet)
		}
	}
}

