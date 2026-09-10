package passport

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestDingTalkAuthorizationResponseIncludesNativeSDKParameters(t *testing.T) {
	raw, err := json.Marshal(DingTalkAuthorizationResponse{
		AuthorizationURL: "https://login.dingtalk.com/oauth2/auth",
		State:            "state-1",
		ClientID:         "app-key",
		RedirectURI:      "https://api.example.com/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{`"clientId":"app-key"`, `"redirectUri":"https://api.example.com/callback"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("authorization response %s missing %s", text, want)
		}
	}
}

func TestBuildDingTalkAuthorizationURL(t *testing.T) {
	got, err := buildDingTalkAuthorizationURL("app-key", "https://client.example.com/login", "state-1")
	if err != nil {
		t.Fatalf("buildDingTalkAuthorizationURL error = %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "login.dingtalk.com" || parsed.Path != "/oauth2/auth" {
		t.Fatalf("authorization endpoint = %s", parsed.String())
	}
	query := parsed.Query()
	for key, want := range map[string]string{
		"client_id": "app-key", "redirect_uri": "https://client.example.com/login", "response_type": "code",
		"scope": "openid", "state": "state-1", "prompt": "consent",
	} {
		if got := query.Get(key); got != want {
			t.Fatalf("query %s = %q, want %q", key, got, want)
		}
	}
}

func TestValidateDingTalkOAuthRedirectURI(t *testing.T) {
	for _, value := range []string{
		"https://client.example.com/login",
		"http://localhost:8081/",
		"http://127.0.0.1:8081/",
		"http://192.168.50.6:8083/native-auth",
		"http://client.example.com/login",
	} {
		if err := validateDingTalkOAuthRedirectURI(value); err != nil {
			t.Fatalf("redirect URI %q should be valid: %v", value, err)
		}
	}
	for _, value := range []string{
		"", "javascript:alert(1)", "https://user:pass@client.example.com/login", "wecheckin://auth/dingtalk", "wecheckin://other/path",
	} {
		if err := validateDingTalkOAuthRedirectURI(value); err == nil {
			t.Fatalf("redirect URI %q should be rejected", value)
		}
	}
}

func TestDingTalkClientLoginUsesExistingBindingAndClientSession(t *testing.T) {
	src, err := os.ReadFile("dingtalk_login.go")
	if err != nil {
		t.Fatalf("read dingtalk_login.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"ExchangeOAuthCodeContext(ctx, config, authCode)",
		"`corp_id` = ? AND `union_id` = ? AND `enabled` = 1",
		"onlineservice.StoreUserSessionContext(ctx, user, token, addIP, device)",
		"ErrDingTalkBindingRequired",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("DingTalk client login missing %q", want)
		}
	}
	for _, forbidden := range []string{"Create(&user)", "BindSelfContext", "StoreDingTalkH5SessionContext"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("DingTalk client login must not auto-create, self-bind, or issue H5 sessions: %q", forbidden)
		}
	}
}
