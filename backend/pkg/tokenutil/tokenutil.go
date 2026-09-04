package tokenutil

import (
	"context"
	"strings"
	"sync"
	"time"

	"wecheckin/backend/internal/config"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

const setupCacheTTL = 30 * time.Second

type setupCacheEntry struct {
	value     string
	expiresAt time.Time
}

var (
	setupCacheMu           sync.RWMutex
	setupCache             = map[string]setupCacheEntry{}
	querySetupValueContext = queryDBSetupValueContext
)

func getDBSetup(key string) string {
	return getDBSetupContext(context.Background(), key)
}

func getDBSetupContext(ctx context.Context, key string) string {
	setupCacheMu.RLock()
	if e, ok := setupCache[key]; ok && time.Now().Before(e.expiresAt) {
		setupCacheMu.RUnlock()
		return e.value
	}
	setupCacheMu.RUnlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return ""
	}
	value := querySetupValueContext(ctx, key)
	if value == "" {
		return ""
	}

	setupCacheMu.Lock()
	setupCache[key] = setupCacheEntry{value: value, expiresAt: time.Now().Add(setupCacheTTL)}
	setupCacheMu.Unlock()
	return value
}

func queryDBSetupValueContext(ctx context.Context, key string) string {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return ""
	}
	var setup model.Setup
	if err := db.Where("setup_key = ?", key).First(&setup).Error; err != nil {
		return ""
	}
	return setup.Value
}

// InvalidateSetupCache 清除所有 setup 缓存。setup_set_content 后调用。
func InvalidateSetupCache() {
	setupCacheMu.Lock()
	setupCache = map[string]setupCacheEntry{}
	setupCacheMu.Unlock()
}

func GetTokenConfig(role string) (expire time.Duration, redisPrefix string) {
	return GetTokenConfigContext(context.Background(), role)
}

func GetTokenConfigContext(ctx context.Context, role string) (expire time.Duration, redisPrefix string) {
	role = strings.ToLower(strings.TrimSpace(role))
	roleUpper := strings.ToUpper(role)

	expireStr := getDBSetupContext(ctx, "TOKEN_"+roleUpper+"_EXPIRE")
	redisPrefix = getDBSetupContext(ctx, "TOKEN_"+roleUpper+"_REDIS_PREFIX")

	if config.Cfg != nil {
		if expireStr == "" {
			switch role {
			case "admin":
				expireStr = config.Cfg.Token.Admin.Expire
			case "dingtalk_h5":
				expireStr = "168h"
			default:
				expireStr = config.Cfg.Token.User.Expire
			}
		}
		if redisPrefix == "" {
			switch role {
			case "admin":
				redisPrefix = config.Cfg.Token.Admin.RedisPrefix
			case "dingtalk_h5":
				redisPrefix = "dingtalk_h5_token:"
			default:
				redisPrefix = config.Cfg.Token.User.RedisPrefix
			}
		}
	}

	if expireStr == "" {
		expireStr = "24h"
	}
	if redisPrefix == "" {
		switch role {
		case "admin":
			redisPrefix = "admin_token:"
		case "dingtalk_h5":
			redisPrefix = "dingtalk_h5_token:"
		default:
			redisPrefix = "user_token:"
		}
	}

	expire, _ = time.ParseDuration(expireStr)
	if expire <= 0 {
		expire = 24 * time.Hour
	}
	return
}

func TokenAuthKey(role, token string) string {
	return TokenAuthKeyContext(context.Background(), role, token)
}

func TokenAuthKeyContext(ctx context.Context, role, token string) string {
	_, prefix := GetTokenConfigContext(ctx, role)
	return prefix + "a:" + token
}

func TokenSetKey(role, id string) string {
	return TokenSetKeyContext(context.Background(), role, id)
}

func TokenSetKeyContext(ctx context.Context, role, id string) string {
	_, prefix := GetTokenConfigContext(ctx, role)
	return prefix + "s:" + id
}

func IsAdminSingleLogin() bool {
	return IsAdminSingleLoginContext(context.Background())
}

func IsAdminSingleLoginContext(ctx context.Context) bool {
	val := getDBSetupContext(ctx, "ADMIN_SINGLE_LOGIN")
	return val == "1" || val == "true"
}

func IsUserSingleLogin() bool {
	return IsUserSingleLoginContext(context.Background())
}

func IsUserSingleLoginContext(ctx context.Context) bool {
	val := getDBSetupContext(ctx, "USER_SINGLE_LOGIN")
	return val == "1" || val == "true"
}

func IsDingTalkH5SingleLogin() bool {
	return IsDingTalkH5SingleLoginContext(context.Background())
}

func IsDingTalkH5SingleLoginContext(ctx context.Context) bool {
	val := getDBSetupContext(ctx, "DINGTALK_H5_SINGLE_LOGIN")
	return val == "1" || val == "true"
}
