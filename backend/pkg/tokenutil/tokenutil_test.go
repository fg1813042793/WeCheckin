package tokenutil

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/config"
)

func TestTokenConfigContextSkipsDatabaseWhenCanceled(t *testing.T) {
	oldCfg := config.Cfg
	oldLoader := querySetupValueContext
	InvalidateSetupCache()
	t.Cleanup(func() {
		config.Cfg = oldCfg
		querySetupValueContext = oldLoader
		InvalidateSetupCache()
	})
	config.Cfg = &config.Config{Token: config.TokenConfig{
		Admin: config.TokenRoleConfig{Expire: "2h", RedisPrefix: "admin:"},
	}}
	queries := 0
	querySetupValueContext = func(context.Context, string) string {
		queries++
		return "unexpected"
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	expire, prefix := GetTokenConfigContext(ctx, "admin")
	if expire != 2*time.Hour || prefix != "admin:" {
		t.Fatalf("GetTokenConfigContext() = (%s, %q), want config fallback", expire, prefix)
	}
	if queries != 0 {
		t.Fatalf("database query count = %d, want 0", queries)
	}
}

func TestTokenContextAPIsUseSetupCacheWithoutDatabase(t *testing.T) {
	oldLoader := querySetupValueContext
	InvalidateSetupCache()
	t.Cleanup(func() {
		querySetupValueContext = oldLoader
		InvalidateSetupCache()
	})
	setupCacheMu.Lock()
	for key, value := range map[string]string{
		"TOKEN_ADMIN_EXPIRE":       "3h",
		"TOKEN_ADMIN_REDIS_PREFIX": "cached-admin:",
		"ADMIN_SINGLE_LOGIN":       "true",
		"USER_SINGLE_LOGIN":        "1",
		"DINGTALK_H5_SINGLE_LOGIN": "true",
	} {
		setupCache[key] = setupCacheEntry{value: value, expiresAt: time.Now().Add(time.Minute)}
	}
	setupCacheMu.Unlock()
	querySetupValueContext = func(context.Context, string) string {
		t.Fatal("cache hit must not query database")
		return ""
	}

	expire, prefix := GetTokenConfigContext(context.Background(), "admin")
	if expire != 3*time.Hour || prefix != "cached-admin:" {
		t.Fatalf("GetTokenConfigContext() = (%s, %q)", expire, prefix)
	}
	if !IsAdminSingleLoginContext(context.Background()) || !IsUserSingleLoginContext(context.Background()) || !IsDingTalkH5SingleLoginContext(context.Background()) {
		t.Fatal("single-login Context APIs must read cached true values")
	}
}

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
