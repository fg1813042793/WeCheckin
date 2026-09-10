package config

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDingTalkOAuthClientExchangesCodeForUserIdentity(t *testing.T) {
	client := defaultDingTalkIdentityClient{baseURL: "https://api.test", httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1.0/oauth2/userAccessToken":
			if r.Method != http.MethodPost {
				t.Fatalf("token method = %s, want POST", r.Method)
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode token request: %v", err)
			}
			if body["clientId"] != "app-key" || body["clientSecret"] != "app-secret" || body["code"] != "oauth-code" || body["grantType"] != "authorization_code" {
				t.Fatalf("token request = %#v", body)
			}
			return jsonTestResponse(map[string]interface{}{"accessToken": "user-token", "expireIn": 7200})
		case "/v1.0/contact/users/me":
			if r.Method != http.MethodGet {
				t.Fatalf("profile method = %s, want GET", r.Method)
			}
			if got := r.Header.Get("x-acs-dingtalk-access-token"); got != "user-token" {
				t.Fatalf("profile access token = %q", got)
			}
			return jsonTestResponse(map[string]interface{}{
				"unionId": "union-1",
				"openId":  "open-1",
				"nick":    "测试用户",
			})
		default:
			t.Fatalf("unexpected DingTalk path: %s", r.URL.Path)
		}
		return jsonTestResponse(map[string]interface{}{})
	})}}

	identity, err := client.ExchangeOAuthCodeContext(context.Background(), DingTalkH5CorpConfig{
		CorpID: "corp-a", AppKey: "app-key", AppSecret: "app-secret",
	}, "oauth-code")
	if err != nil {
		t.Fatalf("ExchangeOAuthCodeContext error = %v", err)
	}
	if identity.UnionID != "union-1" || identity.OpenID != "open-1" || identity.Name != "测试用户" {
		t.Fatalf("identity = %#v", identity)
	}
}

func TestDingTalkOAuthClientRejectsMissingInputs(t *testing.T) {
	client := defaultDingTalkIdentityClient{}
	for _, test := range []struct {
		name   string
		config DingTalkH5CorpConfig
		code   string
		want   string
	}{
		{name: "code", config: DingTalkH5CorpConfig{AppKey: "key", AppSecret: "secret"}, code: " ", want: "钉钉授权码不能为空"},
		{name: "credentials", config: DingTalkH5CorpConfig{}, code: "code", want: "请先配置钉钉 H5 AppKey 和 AppSecret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.ExchangeOAuthCodeContext(context.Background(), test.config, test.code)
			if err == nil || err.Error() != test.want {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}
