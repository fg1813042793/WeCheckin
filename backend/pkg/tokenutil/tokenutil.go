package tokenutil

import (
	"log"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func getDBSetup(key string) string {
	if database.DB == nil {
		return ""
	}
	var setup model.Setup
	if err := database.DB.Where("setup_key = ?", key).First(&setup).Error; err != nil {
		return ""
	}
	return setup.Value
}

func GetTokenConfig(role string) (expire time.Duration, redisPrefix string) {
	roleUpper := strings.ToUpper(role)

	expireStr := getDBSetup("TOKEN_" + roleUpper + "_EXPIRE")
	prefix := getDBSetup("TOKEN_" + roleUpper + "_REDIS_PREFIX")

	log.Printf("[GetTokenConfig] role=%s dbExpire=%q dbPrefix=%q cfg=%v", role, expireStr, prefix, config.Cfg)

	if config.Cfg != nil {
		if expireStr == "" {
			if role == "admin" {
				expireStr = config.Cfg.Token.Admin.Expire
			} else {
				expireStr = config.Cfg.Token.User.Expire
			}
		}
		if prefix == "" {
			if role == "admin" {
				prefix = config.Cfg.Token.Admin.RedisPrefix
				log.Printf("[GetTokenConfig] fallback to cfg admin prefix=%q", prefix)
			} else {
				prefix = config.Cfg.Token.User.RedisPrefix
			}
		}
	}

	if expireStr == "" {
		expireStr = "24h"
	}
	if prefix == "" {
		if role == "admin" {
			prefix = "admin_token:"
			log.Printf("[GetTokenConfig] fallback to hardcoded prefix=%q", prefix)
		} else {
			prefix = "user_token:"
		}
	}

	log.Printf("[GetTokenConfig] final prefix=%q expire=%q", prefix, expireStr)

	expire, _ = time.ParseDuration(expireStr)
	if expire <= 0 {
		expire = 24 * time.Hour
	}
	return
}

func GetAdminPrefix() string {
	_, prefix := GetTokenConfig("admin")
	if prefix == "" {
		prefix = "admin_token:"
	}
	return prefix
}

func GetUserPrefix() string {
	_, prefix := GetTokenConfig("user")
	if prefix == "" {
		prefix = "user_token:"
	}
	return prefix
}
