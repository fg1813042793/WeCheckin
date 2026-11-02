package adminauth

import "testing"

func TestAdminCaptchaUsesIndependentConfiguration(t *testing.T) {
	if AdminLoginCaptchaSetupKey != "ADMIN_LOGIN_CAPTCHA_ENABLED" {
		t.Fatalf("setup key = %q", AdminLoginCaptchaSetupKey)
	}
	if adminCaptchaRedisPrefix != "admin:login:captcha:" {
		t.Fatalf("redis prefix = %q", adminCaptchaRedisPrefix)
	}
}
