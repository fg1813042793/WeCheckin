package setup

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetSetup(key string) (*model.Setup, error) {
	return GetSetupContext(context.Background(), key)
}

func GetSetupContext(ctx context.Context, key string) (*model.Setup, error) {
	now := time.Now()
	if setup, ok := getSetupServiceCache(key, now); ok {
		return setup, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var setup model.Setup
	err := db.Where("`setup_key` = ?", key).First(&setup).Error
	if err != nil {
		return nil, err
	}
	setSetupServiceCache(setup, now)
	return &setup, nil
}

func SetSetup(key, value, typ, addIP string) error {
	return SetSetupContext(context.Background(), key, value, typ, addIP)
}

func SetSetupContext(ctx context.Context, key, value, typ, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var setup model.Setup
	result := db.Where("`setup_key` = ?", key).First(&setup)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			Type:    typ,
			AddTime: database.Now(),
		}
		err := db.Create(&setup).Error
		if err == nil {
			invalidateSetupServiceCache()
		}
		return err
	}
	if result.Error != nil {
		return result.Error
	}
	err := db.Model(&setup).Updates(map[string]interface{}{
		"setup_value":     value,
		"setup_type":      typ,
		"setup_edit_time": database.Now(),
	}).Error
	if err == nil {
		invalidateSetupServiceCache()
	}
	return err
}

func SetContentSetup(key, value, addIP string) error {
	return SetContentSetupContext(context.Background(), key, value, addIP)
}

func SetContentSetupContext(ctx context.Context, key, value, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var setup model.Setup
	result := db.Where("`setup_key` = ?", key).First(&setup)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			AddTime: database.Now(),
		}
		err := db.Create(&setup).Error
		if err == nil {
			invalidateSetupServiceCache()
		}
		return err
	}
	if result.Error != nil {
		return result.Error
	}
	err := db.Model(&setup).Update("setup_value", value).Error
	if err == nil {
		invalidateSetupServiceCache()
	}
	return err
}
