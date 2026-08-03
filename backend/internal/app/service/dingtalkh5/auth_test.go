package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

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
	src, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
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
		"bootstrapForUserDB(ctx, db, user)",
	} {
		if !strings.Contains(helperBody, snippet) {
			t.Fatalf("shared token issuance should sanitize device and bootstrap session with %q", snippet)
		}
	}
}

func TestLogoutContextCleansRedisSessionByStoredSessionUser(t *testing.T) {
	src, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
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
