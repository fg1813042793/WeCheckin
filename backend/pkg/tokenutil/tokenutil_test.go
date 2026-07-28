package tokenutil

import (
	"os"
	"strings"
	"testing"

	"wecheckin-backend/backend/internal/config"
)

func TestTokenRedisKeysUseConfiguredPrefix(t *testing.T) {
	oldCfg := config.Cfg
	t.Cleanup(func() { config.Cfg = oldCfg })

	config.Cfg = &config.Config{
		Token: config.TokenConfig{
			User:  config.TokenRoleConfig{Expire: "24h", RedisPrefix: "custom_user_token:"},
			Admin: config.TokenRoleConfig{Expire: "24h", RedisPrefix: "custom_admin_token:"},
		},
	}

	if got := TokenAuthKey("user", "abc"); got != "custom_user_token:a:abc" {
		t.Fatalf("user auth key = %q", got)
	}
	if got := TokenSetKey("admin", "42"); got != "custom_admin_token:s:42" {
		t.Fatalf("admin set key = %q", got)
	}
}

func TestTokenConfigReadsSetupWithQueryContext(t *testing.T) {
	src, err := os.ReadFile("tokenutil.go")
	if err != nil {
		t.Fatalf("read tokenutil.go: %v", err)
	}
	if strings.Contains(string(src), "database.DB.") {
		t.Fatalf("tokenutil must use database.WithContext instead of direct database.DB calls")
	}
}
