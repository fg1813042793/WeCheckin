package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
)

type fakeDingTalkIdentityClient struct {
	identity configsvc.DingTalkUserIdentity
	err      error
}

func (f fakeDingTalkIdentityClient) ExchangeAuthCodeContext(ctx context.Context, config configsvc.DingTalkH5CorpConfig, authCode string) (configsvc.DingTalkUserIdentity, error) {
	return f.identity, f.err
}

func TestSessionDeviceTrimsAndCapsUserAgent(t *testing.T) {
	const sessionDeviceColumnLength = 255
	raw := "  " + strings.Repeat("钉", sessionDeviceColumnLength+20) + "  "

	got := sessionDevice(raw)

	if len([]rune(got)) != sessionDeviceColumnLength {
		t.Fatalf("session device length = %d, want %d", len([]rune(got)), sessionDeviceColumnLength)
	}
	if maxSessionDeviceLength != sessionDeviceColumnLength {
		t.Fatalf("session device max length = %d, want table column length %d", maxSessionDeviceLength, sessionDeviceColumnLength)
	}
	if strings.HasPrefix(got, " ") || strings.HasSuffix(got, " ") {
		t.Fatalf("session device should be trimmed, got %q", got)
	}
}

func TestLoginContextDoesNotRunFullSeederOnEveryLogin(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	body := functionBody(string(src), "func LoginContext")
	for _, forbidden := range []string{
		"EnsureSeedContext(ctx)",
		"passwordutil.Hash(\"123456\")",
		"upsertDefaultPerfUser",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("LoginContext should not run full seed work on every login, found %q", forbidden)
		}
	}
	for _, snippet := range []string{
		"issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("LoginContext should reuse shared token issuance with %q", snippet)
		}
	}
	helperBody := functionBody(string(src), "func issueDingTalkH5LoginResponseContext")
	for _, snippet := range []string{
		"safeDevice := sessionDevice(device)",
		"StoreDingTalkH5SessionContext(ctx, onlineUserFromDingTalkH5User(user), token, addIP, safeDevice)",
		"bootstrapsvc.BootstrapContext(ctx, user)",
	} {
		if !strings.Contains(helperBody, snippet) {
			t.Fatalf("shared token issuance should sanitize device and bootstrap session with %q", snippet)
		}
	}
}

func TestLogoutContextCleansRedisSessionByStoredSessionUser(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	body := functionBody(string(src), "func LogoutContext")
	if strings.Contains(body, "_ = onlineservice.RemoveUserSessionContext") {
		t.Fatalf("LogoutContext must not ignore Redis session cleanup errors")
	}
	for _, snippet := range []string{
		"onlineservice.DingTalkH5SessionAccountContext(ctx, token)",
		"loadPerfUserByAccountDB(db, account)",
		"onlineservice.RemoveDingTalkH5SessionContext(ctx, userID, token)",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("LogoutContext should clear Redis reliably with %q", snippet)
		}
	}
}

func TestAuthenticateContextUsesLeanPerfUserColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	body := functionBody(text, "func AuthenticateContext")
	for _, snippet := range []string{
		"Select(perfUserSelectColumns)",
		"`user_mini_openid` = ? AND `user_status` = 1",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("AuthenticateContext should avoid SELECT * with %q", snippet)
		}
	}
	if !strings.Contains(text, "const perfUserSelectColumns =") {
		t.Fatalf("auth service should define lean selected columns")
	}
}

func TestLoginByAuthCodeRejectsInvalidAuthCodeBeforeCallingDingTalk(t *testing.T) {
	_, err := loginByAuthCodeContext(context.Background(), fakeDingTalkIdentityClient{}, "corp-a", "  ", "127.0.0.1", "UA")
	if err == nil || err.Error() != "免登授权码不能为空" {
		t.Fatalf("empty auth code error = %v, want 免登授权码不能为空", err)
	}
}

func TestLoginByAuthCodeRejectsEmptyCorpIDBeforeCallingDingTalk(t *testing.T) {
	_, err := loginByAuthCodeContext(context.Background(), fakeDingTalkIdentityClient{}, "  ", "code-1", "127.0.0.1", "UA")
	if err == nil || err.Error() != "钉钉企业 CorpId 不能为空" {
		t.Fatalf("empty corp id error = %v, want 钉钉企业 CorpId 不能为空", err)
	}
}

func TestLoginByAuthCodeRejectsEmptyDingTalkUserID(t *testing.T) {
	_, err := loginByAuthCodeContext(context.Background(), fakeDingTalkIdentityClient{identity: configsvc.DingTalkUserIdentity{}}, "corp-a", "code-1", "127.0.0.1", "UA")
	if err == nil || err.Error() != "钉钉身份异常" {
		t.Fatalf("empty DingTalk userid error = %v, want 钉钉身份异常", err)
	}
}

func TestDingTalkH5SSOImplementationKeepsAuthorizationInCurrentSystem(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"configsvc.LoadDingTalkH5UserBindingDB(db, corpID, identity.UserID)",
		"loadPerfUserByIDDB(db, binding.UserID)",
		"hydratePerfUserWithUserDeptDB(db, &user)",
		"issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)",
		"钉钉账号未开通绩效系统，请联系管理员",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("SSO login should preserve local user binding and authorization bootstrap, missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"Create(&user)",
		"upsertDefaultPerfUser",
		"SetRoleApplicationMenuPermissions",
		"SetRoleApplicationAPIPermissions",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("SSO login must not auto-create users or grant permissions, found %q", forbidden)
		}
	}
}

func TestDingTalkOAPIClientUsesConfigAndNoHardcodedSecret(t *testing.T) {
	src, err := os.ReadFile("../config/dingtalk_oapi.go")
	if err != nil {
		t.Fatalf("read dingtalk_oapi.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"DingTalkH5CorpConfig",
		"config.AppKey",
		"config.AppSecret",
		"/gettoken",
		"/topapi/v2/user/getuserinfo",
		"ExchangeAuthCodeContext",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("DingTalk OAPI client missing %q", want)
		}
	}
	if strings.Contains(text, "appsecret=\"") || strings.Contains(text, "appSecret := \"") {
		t.Fatalf("DingTalk OAPI client must not hardcode app secrets")
	}
}

func TestDingTalkH5SSOSourceRequiresCorpID(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"func LoginByAuthCodeContext(ctx context.Context, corpID, authCode, addIP, device string)",
		"configsvc.LoadDingTalkH5CorpConfigContext(ctx, corpID)",
		"client.ExchangeAuthCodeContext(ctx, config, authCode)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("multi-corp SSO must include %q", want)
		}
	}
}

func functionBody(src, signature string) string {
	start := strings.Index(src, signature)
	if start < 0 {
		return ""
	}
	body := src[start:]
	depth := 0
	for index, r := range body {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[:index+1]
			}
		}
	}
	return body
}
