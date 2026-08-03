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
	setupCacheMu sync.RWMutex
	setupCache   = map[string]setupCacheEntry{}
)

func getDBSetup(key string) string {
	setupCacheMu.RLock()
	if e, ok := setupCache[key]; ok && time.Now().Before(e.expiresAt) {
		setupCacheMu.RUnlock()
		return e.value
	}
	setupCacheMu.RUnlock()

	db, cancel := database.WithContext(context.Background())
	defer cancel()
	if db == nil {
		return ""
	}
	var setup model.Setup
	if err := db.Where("setup_key = ?", key).First(&setup).Error; err != nil {
		return ""
	}

	setupCacheMu.Lock()
	setupCache[key] = setupCacheEntry{value: setup.Value, expiresAt: time.Now().Add(setupCacheTTL)}
	setupCacheMu.Unlock()
	return setup.Value
}

// InvalidateSetupCache 清除所有 setup 缓存。setup_set_content 后调用。
func InvalidateSetupCache() {
	setupCacheMu.Lock()
	setupCache = map[string]setupCacheEntry{}
	setupCacheMu.Unlock()
}

func GetTokenConfig(role string) (expire time.Duration, redisPrefix string) {
	role = strings.ToLower(strings.TrimSpace(role))
	roleUpper := strings.ToUpper(role)

	expireStr := getDBSetup("TOKEN_" + roleUpper + "_EXPIRE")
	redisPrefix = getDBSetup("TOKEN_" + roleUpper + "_REDIS_PREFIX")

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
	_, prefix := GetTokenConfig(role)
	return prefix + "a:" + token
}

func TokenSetKey(role, id string) string {
	_, prefix := GetTokenConfig(role)
	return prefix + "s:" + id
}

func IsAdminSingleLogin() bool {
	val := getDBSetup("ADMIN_SINGLE_LOGIN")
	return val == "1" || val == "true"
}

func IsUserSingleLogin() bool {
	val := getDBSetup("USER_SINGLE_LOGIN")
	return val == "1" || val == "true"
}

func IsDingTalkH5SingleLogin() bool {
	val := getDBSetup("DINGTALK_H5_SINGLE_LOGIN")
	return val == "1" || val == "true"
}
