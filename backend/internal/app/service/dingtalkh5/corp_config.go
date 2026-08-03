package dingtalkh5

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	setupservice "wecheckin/backend/internal/app/service/setup"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type DingTalkH5CorpConfig struct {
	CorpID       string
	CorpName     string
	AppKey       string
	AppSecret    string
	AgentID      string
	Enabled      int
	AppSecretSet bool
}

func ListDingTalkH5CorpConfigsContext(ctx context.Context) ([]DingTalkH5CorpConfig, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return []DingTalkH5CorpConfig{}, nil
	}

	var rows []model.DingTalkH5CorpConfig
	if err := db.Order("`id` ASC").Find(&rows).Error; err != nil && !isMissingDingTalkH5MultiCorpTableError(err) {
		return nil, err
	}
	configs := make([]DingTalkH5CorpConfig, 0, len(rows))
	for _, row := range rows {
		configs = append(configs, dingTalkH5CorpConfigFromModel(row))
	}
	if len(configs) > 0 {
		return configs, nil
	}
	if legacy := legacyDingTalkH5CorpConfigContext(ctx); strings.TrimSpace(legacy.CorpID) != "" {
		return []DingTalkH5CorpConfig{legacy}, nil
	}
	return configs, nil
}

func SaveDingTalkH5CorpConfigsContext(ctx context.Context, configs []DingTalkH5CorpConfig) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	now := database.Now()
	saved := make([]DingTalkH5CorpConfig, 0, len(configs))
	savedCorpIDs := make([]string, 0, len(configs))
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			config = normalizeDingTalkH5CorpConfig(config)
			if config.CorpID == "" {
				continue
			}
			if config.Enabled != 0 {
				config.Enabled = 1
			}
			if err := upsertDingTalkH5CorpConfigDB(tx, config, now); err != nil {
				return err
			}
			saved = append(saved, config)
			savedCorpIDs = append(savedCorpIDs, config.CorpID)
		}
		return deleteOmittedDingTalkH5CorpConfigsDB(tx, savedCorpIDs)
	}); err != nil {
		return err
	}
	if len(saved) > 0 {
		return mirrorPrimaryDingTalkH5CorpConfigToSetupContext(ctx, saved[0])
	}
	return clearLegacyDingTalkH5CorpConfigContext(ctx)
}

func loadDingTalkH5CorpConfigContext(ctx context.Context, corpID string) (DingTalkH5CorpConfig, error) {
	corpID = strings.TrimSpace(corpID)
	if corpID == "" {
		return DingTalkH5CorpConfig{}, fmt.Errorf("钉钉企业 CorpId 不能为空")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return DingTalkH5CorpConfig{CorpID: corpID}, nil
	}

	var row model.DingTalkH5CorpConfig
	err := db.Where("`corp_id` = ? AND `enabled` = 1", corpID).First(&row).Error
	if err == nil {
		config := dingTalkH5CorpConfigFromModel(row)
		if strings.TrimSpace(config.AppKey) == "" || strings.TrimSpace(config.AppSecret) == "" {
			return DingTalkH5CorpConfig{}, fmt.Errorf("请先配置钉钉 H5 AppKey 和 AppSecret")
		}
		return config, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !isMissingDingTalkH5MultiCorpTableError(err) {
		return DingTalkH5CorpConfig{}, err
	}

	legacy := legacyDingTalkH5CorpConfigContext(ctx)
	if legacy.CorpID == corpID {
		if strings.TrimSpace(legacy.AppKey) == "" || strings.TrimSpace(legacy.AppSecret) == "" {
			return DingTalkH5CorpConfig{}, fmt.Errorf("请先配置钉钉 H5 AppKey 和 AppSecret")
		}
		return legacy, nil
	}
	return DingTalkH5CorpConfig{}, fmt.Errorf("钉钉企业未配置或已停用")
}

func loadDingTalkH5UserBindingDB(db *gorm.DB, corpID, dingTalkUserID string) (*model.DingTalkH5UserBinding, error) {
	corpID = strings.TrimSpace(corpID)
	dingTalkUserID = strings.TrimSpace(dingTalkUserID)
	if corpID == "" || dingTalkUserID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var binding model.DingTalkH5UserBinding
	if err := db.Where("`corp_id` = ? AND `dingtalk_user_id` = ? AND `enabled` = 1", corpID, dingTalkUserID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func loadAnyDingTalkH5UserBindingDB(db *gorm.DB, corpID, dingTalkUserID string) (*model.DingTalkH5UserBinding, error) {
	corpID = strings.TrimSpace(corpID)
	dingTalkUserID = strings.TrimSpace(dingTalkUserID)
	if corpID == "" || dingTalkUserID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var binding model.DingTalkH5UserBinding
	if err := db.Where("`corp_id` = ? AND `dingtalk_user_id` = ?", corpID, dingTalkUserID).First(&binding).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func loadLegacyDingTalkH5UserByIdentityDB(ctx context.Context, db *gorm.DB, config DingTalkH5CorpConfig, identity DingTalkUserIdentity) (*model.DingTalkH5PerfUser, error) {
	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	legacy := legacyDingTalkH5CorpConfigContext(ctx)
	if strings.TrimSpace(legacy.CorpID) != strings.TrimSpace(config.CorpID) {
		return nil, gorm.ErrRecordNotFound
	}
	return loadPerfUserByAccountDB(db, identity.UserID)
}

func upsertDingTalkH5CorpConfigDB(db *gorm.DB, config DingTalkH5CorpConfig, now int64) error {
	var row model.DingTalkH5CorpConfig
	result := db.Where("`corp_id` = ?", config.CorpID).First(&row)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		row = model.DingTalkH5CorpConfig{
			CorpID:    config.CorpID,
			CorpName:  config.CorpName,
			AppKey:    config.AppKey,
			AppSecret: config.AppSecret,
			AgentID:   config.AgentID,
			Enabled:   config.Enabled,
			AddTime:   now,
			EditTime:  now,
		}
		return db.Create(&row).Error
	}
	if result.Error != nil {
		return result.Error
	}
	updates := map[string]interface{}{
		"corp_name": config.CorpName,
		"app_key":   config.AppKey,
		"agent_id":  config.AgentID,
		"enabled":   config.Enabled,
		"edit_time": now,
	}
	if strings.TrimSpace(config.AppSecret) != "" {
		updates["app_secret"] = config.AppSecret
	}
	return db.Model(&row).Updates(updates).Error
}

func deleteOmittedDingTalkH5CorpConfigsDB(db *gorm.DB, keepCorpIDs []string) error {
	if len(keepCorpIDs) == 0 {
		return db.Where("1 = 1").Delete(&model.DingTalkH5CorpConfig{}).Error
	}
	return db.Where("`corp_id` NOT IN ?", keepCorpIDs).Delete(&model.DingTalkH5CorpConfig{}).Error
}

func dingTalkH5CorpConfigFromModel(row model.DingTalkH5CorpConfig) DingTalkH5CorpConfig {
	return DingTalkH5CorpConfig{
		CorpID:       strings.TrimSpace(row.CorpID),
		CorpName:     strings.TrimSpace(row.CorpName),
		AppKey:       strings.TrimSpace(row.AppKey),
		AppSecret:    strings.TrimSpace(row.AppSecret),
		AgentID:      strings.TrimSpace(row.AgentID),
		Enabled:      row.Enabled,
		AppSecretSet: strings.TrimSpace(row.AppSecret) != "",
	}
}

func normalizeDingTalkH5CorpConfig(config DingTalkH5CorpConfig) DingTalkH5CorpConfig {
	config.CorpID = strings.TrimSpace(config.CorpID)
	config.CorpName = strings.TrimSpace(config.CorpName)
	config.AppKey = strings.TrimSpace(config.AppKey)
	config.AppSecret = strings.TrimSpace(config.AppSecret)
	config.AgentID = strings.TrimSpace(config.AgentID)
	if config.Enabled != 0 {
		config.Enabled = 1
	}
	config.AppSecretSet = strings.TrimSpace(config.AppSecret) != ""
	return config
}

func legacyDingTalkH5CorpConfigContext(ctx context.Context) DingTalkH5CorpConfig {
	appSecret := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_SECRET"))
	return DingTalkH5CorpConfig{
		CorpID:       strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_CORP_ID")),
		CorpName:     strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_CORP_ID")),
		AppKey:       strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_KEY")),
		AppSecret:    appSecret,
		AgentID:      strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_AGENT_ID")),
		Enabled:      1,
		AppSecretSet: appSecret != "",
	}
}

func mirrorPrimaryDingTalkH5CorpConfigToSetupContext(ctx context.Context, config DingTalkH5CorpConfig) error {
	for _, item := range []struct {
		key   string
		value string
		typ   string
	}{
		{key: "DINGTALK_H5_CORP_ID", value: config.CorpID, typ: "string"},
		{key: "DINGTALK_H5_APP_KEY", value: config.AppKey, typ: "string"},
		{key: "DINGTALK_H5_AGENT_ID", value: config.AgentID, typ: "string"},
	} {
		if err := setupservice.SetSetupContext(ctx, item.key, item.value, item.typ, ""); err != nil {
			return err
		}
	}
	if strings.TrimSpace(config.AppSecret) != "" {
		return setupservice.SetSetupContext(ctx, "DINGTALK_H5_APP_SECRET", config.AppSecret, "password", "")
	}
	return nil
}

func clearLegacyDingTalkH5CorpConfigContext(ctx context.Context) error {
	for _, item := range []struct {
		key string
		typ string
	}{
		{key: "DINGTALK_H5_CORP_ID", typ: "string"},
		{key: "DINGTALK_H5_APP_KEY", typ: "string"},
		{key: "DINGTALK_H5_AGENT_ID", typ: "string"},
		{key: "DINGTALK_H5_APP_SECRET", typ: "password"},
	} {
		if err := setupservice.SetSetupContext(ctx, item.key, "", item.typ, ""); err != nil {
			return err
		}
	}
	return nil
}

func dingTalkH5SetupValueContext(ctx context.Context, key string) string {
	setup, err := setupservice.GetSetupContext(ctx, key)
	if err != nil || setup == nil {
		return ""
	}
	return setup.Value
}

func isMissingDingTalkH5MultiCorpTableError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "dingtalk_h5_corp_configs") &&
		(strings.Contains(text, "doesn't exist") || strings.Contains(text, "no such table"))
}

func isMissingDingTalkH5UserBindingTableError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "dingtalk_h5_user_bindings") &&
		(strings.Contains(text, "doesn't exist") || strings.Contains(text, "no such table"))
}
