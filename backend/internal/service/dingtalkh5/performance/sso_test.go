package performance

import (
	"context"
	"os"
	"strings"
	"testing"
)

type fakeDingTalkIdentityClient struct {
	identity   DingTalkUserIdentity
	err        error
	seenCorpID string
}

func (f fakeDingTalkIdentityClient) ExchangeAuthCodeContext(ctx context.Context, config DingTalkH5CorpConfig, authCode string) (DingTalkUserIdentity, error) {
	return f.identity, f.err
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
	_, err := loginByAuthCodeContext(context.Background(), fakeDingTalkIdentityClient{identity: DingTalkUserIdentity{}}, "corp-a", "code-1", "127.0.0.1", "UA")
	if err == nil || err.Error() != "钉钉身份异常" {
		t.Fatalf("empty DingTalk userid error = %v, want 钉钉身份异常", err)
	}
}

func TestDingTalkH5SSOImplementationKeepsAuthorizationInCurrentSystem(t *testing.T) {
	src, err := os.ReadFile("sso.go")
	if err != nil {
		t.Fatalf("read sso.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"loadDingTalkH5UserBindingDB(db, corpID, identity.UserID)",
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
	src, err := os.ReadFile("dingtalk_oapi.go")
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
	src, err := os.ReadFile("sso.go")
	if err != nil {
		t.Fatalf("read sso.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"func LoginByAuthCodeContext(ctx context.Context, corpID, authCode, addIP, device string)",
		"loadDingTalkH5CorpConfigContext(ctx, corpID)",
		"client.ExchangeAuthCodeContext(ctx, config, authCode)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("multi-corp SSO must include %q", want)
		}
	}
}
