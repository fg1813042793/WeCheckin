package tokenutil

import (
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
