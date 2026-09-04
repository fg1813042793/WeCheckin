package setup

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/tokenutil"
)

type SetupItem struct {
	Key   string
	Value string
	Type  string
}

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
	err := db.Where("`setup_key` = ?", key).Take(&setup).Error
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
	return SetSetupsContext(ctx, []SetupItem{{Key: key, Value: value, Type: typ}}, addIP)
}

func SetSetupsContext(ctx context.Context, items []SetupItem, addIP string) error {
	if len(items) == 0 {
		return nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()

	now := database.Now()
	rowsByKey := make(map[string]model.Setup, len(items))
	keys := make([]string, 0, len(items))
	for _, item := range items {
		if item.Key == "" {
			continue
		}
		if _, ok := rowsByKey[item.Key]; !ok {
			keys = append(keys, item.Key)
		}
		rowsByKey[item.Key] = model.Setup{
			Key:      item.Key,
			Value:    item.Value,
			Type:     item.Type,
			AddTime:  now,
			EditTime: now,
		}
	}
	if len(rowsByKey) == 0 {
		return nil
	}

	rows := make([]model.Setup, 0, len(rowsByKey))
	for _, key := range keys {
		rows = append(rows, rowsByKey[key])
	}
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "setup_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"setup_value",
			"setup_type",
			"setup_edit_time",
			"updated_at",
		}),
	}).CreateInBatches(rows, 100).Error
	if err == nil {
		invalidateRelatedSetupCachesForKeys(keys)
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
	result := db.Where("`setup_key` = ?", key).Take(&setup)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		setup = model.Setup{
			Key:     key,
			Value:   value,
			AddTime: database.Now(),
		}
		err := db.Create(&setup).Error
		if err == nil {
			invalidateRelatedSetupCaches(key)
		}
		return err
	}
	if result.Error != nil {
		return result.Error
	}
	err := db.Model(&setup).Update("setup_value", value).Error
	if err == nil {
		invalidateRelatedSetupCaches(key)
	}
	return err
}

func invalidateRelatedSetupCaches(key string) {
	invalidateSetupServiceCache()
	tokenutil.InvalidateSetupCache()
	if key == "STATIC_DOMAIN" {
		media.InvalidateStaticDomainCache()
	}
}

func invalidateRelatedSetupCachesForKeys(keys []string) {
	invalidateSetupServiceCache()
	tokenutil.InvalidateSetupCache()
	for _, key := range keys {
		if key == "STATIC_DOMAIN" {
			media.InvalidateStaticDomainCache()
			return
		}
	}
}
