package passwordsession

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordChangesAndAccountDisableRevokeAllUserSessions(t *testing.T) {
	backendRoot := filepath.Join("..", "..")
	checks := []struct {
		path     string
		required []string
	}{
		{
			path: filepath.Join(backendRoot, "service", "accountsecurity", "password.go"),
			required: []string{
				"func ChangeOwnPasswordContext",
				"func ResetTemporaryPasswordContext",
				"passwordsession.InvalidateAllContext(ctx, userID)",
			},
		},
		{
			path: filepath.Join(backendRoot, "service", "admin", "adminmgr", "service.go"),
			required: []string{
				"func EditContext",
				"func SetStatusContext",
				"passwordsession.InvalidateAllContext(ctx, uint(uid))",
			},
		},
		{
			path: filepath.Join(backendRoot, "service", "admin", "adminuser", "service.go"),
			required: []string{
				"func invalidateUserSessionsContext",
				"func invalidateUserSessionsWhenDisabledContext",
				"passwordsession.InvalidateAllContext(ctx, uint(userID))",
				"invalidateUserSessionsContext(ctx, id)",
				"invalidateUserSessionsWhenDisabledContext(ctx, id, status)",
				"if err := DelUserContext(ctx, id); err != nil",
			},
		},
		{
			path: filepath.Join(backendRoot, "service", "dingtalkh5", "performance", "user", "service.go"),
			required: []string{
				"passwordChanged || next.Status != 1",
				"passwordsession.InvalidateAllContext(ctx, target.ID)",
			},
		},
	}

	for _, check := range checks {
		data, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		text := string(data)
		for _, snippet := range check.required {
			if !strings.Contains(text, snippet) {
				t.Errorf("%s missing session revocation contract %q", check.path, snippet)
			}
		}
	}
}

func TestSelfPasswordChangeDoesNotKeepCurrentTokenAlive(t *testing.T) {
	backendRoot := filepath.Join("..", "..")
	for _, path := range []string{
		filepath.Join(backendRoot, "handler", "admin", "mgr", "handler.go"),
		filepath.Join(backendRoot, "handler", "client", "passport", "password.go"),
		filepath.Join(backendRoot, "handler", "dingtalkh5", "account", "handler.go"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), "passwordsession.RefreshCurrentContext") {
			t.Errorf("%s must not refresh the current token after password change", path)
		}
	}
}
