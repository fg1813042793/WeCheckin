package passwordsession

import "testing"

func TestRestrictedSessionAllowlist(t *testing.T) {
	tests := []struct {
		audience Audience
		method   string
		path     string
		allowed  bool
	}{
		{AudienceAdmin, "PATCH", "/api/v2/admin/me/password", true},
		{AudienceAdmin, "POST", "/api/v2/admin/auth/logout", true},
		{AudienceAdmin, "GET", "/api/v2/admin/home", false},
		{AudienceClient, "PATCH", "/api/v2/me/password", true},
		{AudienceClient, "POST", "/api/v2/me/logout", true},
		{AudienceClient, "GET", "/api/v2/news", false},
		{AudienceDingTalkH5, "PATCH", "/api/v2/dingtalk/h5/account/password", true},
		{AudienceDingTalkH5, "POST", "/api/v2/dingtalk/h5/logout", true},
		{AudienceDingTalkH5, "GET", "/api/v2/dingtalk/h5/workbench", false},
	}
	for _, tc := range tests {
		if got := Allowed(tc.audience, tc.method, tc.path); got != tc.allowed {
			t.Fatalf("Allowed(%s, %s, %s)=%v want %v", tc.audience, tc.method, tc.path, got, tc.allowed)
		}
	}
}

func TestRequiredResponseUsesStableCode(t *testing.T) {
	if ErrorCode != "PASSWORD_CHANGE_REQUIRED" || ErrorMessage == "" {
		t.Fatalf("restriction contract = %q %q", ErrorCode, ErrorMessage)
	}
}
