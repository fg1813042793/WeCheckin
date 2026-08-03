package bootstrap

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
)

func seedSetups() error {
	db, cancel := startupDB(context.Background())
	defer cancel()
	if db == nil {
		return errors.New("database is not initialized")
	}

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
		{Key: "DINGTALK_H5_SINGLE_LOGIN", Value: "0", Type: "switch"},
		{Key: "DINGTALK_H5_CORP_ID", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_APP_KEY", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_APP_SECRET", Value: "", Type: "password"},
		{Key: "DINGTALK_H5_AGENT_ID", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_NOTIFY_ENABLED", Value: "0", Type: "switch"},
		{Key: "DINGTALK_H5_APP_NAME", Value: "OA管理", Type: "string"},
		{Key: "DINGTALK_H5_LOGO_TEXT", Value: "OA", Type: "string"},
		{Key: "DINGTALK_H5_LOGO_URL", Value: "", Type: "string"},
		{Key: "TOKEN_DINGTALK_H5_EXPIRE", Value: "168h", Type: "string"},
		{Key: "TOKEN_DINGTALK_H5_REDIS_PREFIX", Value: "dingtalk_h5_token:", Type: "string"},
	}
	for _, d := range defs {
		var existing model.Setup
		if err := db.Where("setup_key = ?", d.Key).First(&existing).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			setup := model.Setup{
				Key:     d.Key,
				Value:   d.Value,
				Type:    d.Type,
				AddTime: time.Now().UnixMilli(),
			}
			if err := db.Create(&setup).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
