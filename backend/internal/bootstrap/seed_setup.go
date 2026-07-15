package bootstrap

import (
	"log"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func seedSetups() {
	type setupDef struct {
		Key   string
		Value string
		Type  string
	}
	defs := []setupDef{
		{Key: "ADMIN_SINGLE_LOGIN", Value: "0", Type: "switch"},
		{Key: "USER_SINGLE_LOGIN", Value: "0", Type: "switch"},
		{Key: "TOKEN_ADMIN_EXPIRE", Value: "168h", Type: "string"},
		{Key: "TOKEN_ADMIN_REDIS_PREFIX", Value: "admin_token:", Type: "string"},
		{Key: "TOKEN_USER_EXPIRE", Value: "999d", Type: "string"},
		{Key: "TOKEN_USER_REDIS_PREFIX", Value: "user_token:", Type: "string"},
	}
	for _, d := range defs {
		var existing model.Setup
		if err := database.DB.Where("setup_key = ?", d.Key).First(&existing).Error; err != nil {
			setup := model.Setup{
				Key:     d.Key,
				Value:   d.Value,
				Type:    d.Type,
				AddTime: time.Now().UnixMilli(),
			}
			if err := database.DB.Create(&setup).Error; err != nil {
				log.Printf("seed setup %s error: %v", d.Key, err)
			}
		}
	}
}
