package service

import (
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func SetSetup(key, value, typ, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`setup_key` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			Type:    typ,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Updates(map[string]interface{}{
		"setup_value":     value,
		"setup_type":      typ,
		"setup_edit_time": database.Now(),
	}).Error
}

func SetContentSetup(key, value, addIP string) error {
	var setup model.Setup
	result := database.DB.Where("`setup_key` = ?", key).First(&setup)
	if result.Error != nil {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			AddTime: database.Now(),
		}
		return database.DB.Create(&setup).Error
	}
	return database.DB.Model(&setup).Update("setup_value", value).Error
}
