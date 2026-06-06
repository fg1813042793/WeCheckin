package tokenutil

import (
	"strings"
	"sync"
	"time"

	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
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

	if database.DB == nil {
		return ""
	}
	var setup model.Setup
	if err := database.DB.Where("setup_key = ?", key).First(&setup).Error; err != nil {
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
	roleUpper := strings.ToUpper(role)

	expireStr := getDBSetup("TOKEN_" + roleUpper + "_EXPIRE")
	redisPrefix = getDBSetup("TOKEN_" + roleUpper + "_REDIS_PREFIX")

	if config.Cfg != nil {
		if expireStr == "" {
			if role == "admin" {
				expireStr = config.Cfg.Token.Admin.Expire
			} else {
				expireStr = config.Cfg.Token.User.Expire
			}
		}
		if redisPrefix == "" {
			if role == "admin" {
				redisPrefix = config.Cfg.Token.Admin.RedisPrefix
			} else {
				redisPrefix = config.Cfg.Token.User.RedisPrefix
			}
		}
	}

	if expireStr == "" {
		expireStr = "24h"
	}
	if redisPrefix == "" {
		if role == "admin" {
			redisPrefix = "admin_token:"
		} else {
			redisPrefix = "user_token:"
		}
	}

	expire, _ = time.ParseDuration(expireStr)
	if expire <= 0 {
		expire = 24 * time.Hour
	}
	return
}

func IsAdminSingleLogin() bool {
	val := getDBSetup("ADMIN_SINGLE_LOGIN")
	return val == "1" || val == "true"
}

func IsUserSingleLogin() bool {
	val := getDBSetup("USER_SINGLE_LOGIN")
	return val == "1" || val == "true"
}
