package formkitadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type TemplatePreset = map[string]string

func GetTemplatePresetsContext(ctx context.Context, adminID uint) ([]TemplatePreset, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	key := templatePresetKey(adminID)
	var entry model.Setup
	if err := db.Where("`setup_key` = ?", key).First(&entry).Error; err != nil {
		if IsNotFound(err) {
			return []TemplatePreset{}, nil
		}
		return nil, err
	}
	var out []TemplatePreset
	if entry.Value != "" {
		if err := json.Unmarshal([]byte(entry.Value), &out); err != nil {
			return nil, err
		}
	}
	if out == nil {
		out = []TemplatePreset{}
	}
	return out, nil
}

func SaveTemplatePresetsContext(ctx context.Context, adminID uint, presets []TemplatePreset) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	key := templatePresetKey(adminID)
	now := time.Now().Unix()
	bytes, err := json.Marshal(presets)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var entry model.Setup
		err := tx.Where("`setup_key` = ?", key).First(&entry).Error
		if err == nil {
			return tx.Model(&model.Setup{}).Where("`setup_key` = ?", key).Updates(map[string]interface{}{
				"setup_value":     string(bytes),
				"setup_edit_time": now,
			}).Error
		}
		if !IsNotFound(err) {
			return err
		}
		return tx.Create(&model.Setup{
			Key:      key,
			Value:    string(bytes),
			Type:     "template_presets",
			AddTime:  now,
			EditTime: now,
		}).Error
	})
}

func templatePresetKey(adminID uint) string {
	return fmt.Sprintf("template_presets_%d", adminID)
}
