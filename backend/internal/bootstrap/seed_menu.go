package bootstrap

import (
	"context"
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"
	permissionsupport "wecheckin-backend/backend/internal/app/support/permission"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

const menuSeedInitializedKey = "SYSTEM_MENU_SEED_INITIALIZED"

func seedMenus(enableExam bool) error {
	db, cancel := startupDB(context.Background())
	defer cancel()
	if db == nil {
		return errors.New("database is not initialized")
	}
	value, err := getMenuSeedInitializedValue(db)
	if err != nil {
		return err
	}
	if isMenuSeedInitialized(value) {
		log.Printf("seed permissions skipped: already initialized")
		return nil
	}

	if err := permissionsupport.SyncAdminMenuPermissionsContext(context.Background(), db, enableExam); err != nil {
		return err
	}
	return markMenuSeedInitialized(db)
}

func isMenuSeedInitialized(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func getMenuSeedInitializedValue(db *gorm.DB) (string, error) {
	var setup model.Setup
	err := db.Where("setup_key = ?", menuSeedInitializedKey).First(&setup).Error
	if err == nil {
		return setup.Value, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return "", err
}

func markMenuSeedInitialized(db *gorm.DB) error {
	now := database.Now()
	var setup model.Setup
	err := db.Where("setup_key = ?", menuSeedInitializedKey).First(&setup).Error
	if err == nil {
		if isMenuSeedInitialized(setup.Value) {
			return nil
		}
		if err := db.Model(&setup).Updates(map[string]interface{}{
			"setup_value":     "1",
			"setup_type":      "system",
			"setup_edit_time": now,
		}).Error; err != nil {
			return err
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := db.Create(&model.Setup{
		Key:      menuSeedInitializedKey,
		Value:    "1",
		Type:     "system",
		AddTime:  now,
		EditTime: now,
	}).Error; err != nil {
		return err
	}
	return nil
}
