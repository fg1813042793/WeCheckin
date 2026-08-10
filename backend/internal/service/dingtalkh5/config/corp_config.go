package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/pkg/database"
)

type DingTalkH5CorpConfig struct {
	CorpID        string
	CorpName      string
	AppKey        string
	AppSecret     string
	AgentID       string
	UnifiedAppID  string
	AppURL        string
	NotifyEnabled int
	NotifyMode    string
	RobotCode     string
	Enabled       int
	AppSecretSet  bool
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

func LoadDingTalkH5CorpConfigContext(ctx context.Context, corpID string) (DingTalkH5CorpConfig, error) {
	return loadDingTalkH5CorpConfigContext(ctx, corpID)
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

func LoadDingTalkH5UserBindingDB(db *gorm.DB, corpID, dingTalkUserID string) (*model.DingTalkH5UserBinding, error) {
	return loadDingTalkH5UserBindingDB(db, corpID, dingTalkUserID)
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

func LoadAnyDingTalkH5UserBindingDB(db *gorm.DB, corpID, dingTalkUserID string) (*model.DingTalkH5UserBinding, error) {
	return loadAnyDingTalkH5UserBindingDB(db, corpID, dingTalkUserID)
}

func loadLegacyDingTalkH5UserByIdentityDB(ctx context.Context, db *gorm.DB, config DingTalkH5CorpConfig, identity DingTalkUserIdentity) (*model.DingTalkH5PerfUser, error) {
	return nil, gorm.ErrRecordNotFound
}

func upsertDingTalkH5CorpConfigDB(db *gorm.DB, config DingTalkH5CorpConfig, now int64) error {
	var row model.DingTalkH5CorpConfig
	result := db.Where("`corp_id` = ?", config.CorpID).First(&row)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		row = model.DingTalkH5CorpConfig{
			CorpID:        config.CorpID,
			CorpName:      config.CorpName,
			AppKey:        config.AppKey,
			AppSecret:     config.AppSecret,
			AgentID:       config.AgentID,
			UnifiedAppID:  config.UnifiedAppID,
			AppURL:        config.AppURL,
			NotifyEnabled: config.NotifyEnabled,
			NotifyMode:    config.NotifyMode,
			RobotCode:     config.RobotCode,
			Enabled:       config.Enabled,
			AddTime:       now,
			EditTime:      now,
		}
		return db.Create(&row).Error
	}
	if result.Error != nil {
		return result.Error
	}
	updates := map[string]interface{}{
		"corp_name":      config.CorpName,
		"app_key":        config.AppKey,
		"agent_id":       config.AgentID,
		"unified_app_id": config.UnifiedAppID,
		"app_url":        config.AppURL,
		"notify_enabled": config.NotifyEnabled,
		"notify_mode":    config.NotifyMode,
		"robot_code":     config.RobotCode,
		"enabled":        config.Enabled,
		"edit_time":      now,
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
		CorpID:        strings.TrimSpace(row.CorpID),
		CorpName:      strings.TrimSpace(row.CorpName),
		AppKey:        strings.TrimSpace(row.AppKey),
		AppSecret:     strings.TrimSpace(row.AppSecret),
		AgentID:       strings.TrimSpace(row.AgentID),
		UnifiedAppID:  strings.TrimSpace(row.UnifiedAppID),
		AppURL:        strings.TrimSpace(row.AppURL),
		NotifyEnabled: normalizeDingTalkH5SwitchInt(row.NotifyEnabled),
		NotifyMode:    normalizeDingTalkH5NotifyMode(row.NotifyMode, row.AgentID, row.RobotCode),
		RobotCode:     strings.TrimSpace(row.RobotCode),
		Enabled:       row.Enabled,
		AppSecretSet:  strings.TrimSpace(row.AppSecret) != "",
	}
}

func normalizeDingTalkH5CorpConfig(config DingTalkH5CorpConfig) DingTalkH5CorpConfig {
	config.CorpID = strings.TrimSpace(config.CorpID)
	config.CorpName = strings.TrimSpace(config.CorpName)
	config.AppKey = strings.TrimSpace(config.AppKey)
	config.AppSecret = strings.TrimSpace(config.AppSecret)
	config.AgentID = strings.TrimSpace(config.AgentID)
	config.UnifiedAppID = strings.TrimSpace(config.UnifiedAppID)
	config.AppURL = strings.TrimSpace(config.AppURL)
	config.NotifyEnabled = normalizeDingTalkH5SwitchInt(config.NotifyEnabled)
	config.NotifyMode = normalizeDingTalkH5NotifyMode(config.NotifyMode, config.AgentID, config.RobotCode)
	config.RobotCode = strings.TrimSpace(config.RobotCode)
	if config.Enabled != 0 {
		config.Enabled = 1
	}
	config.AppSecretSet = strings.TrimSpace(config.AppSecret) != ""
	return config
}

func legacyDingTalkH5CorpConfigContext(ctx context.Context) DingTalkH5CorpConfig {
	appSecret := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_SECRET"))
	return DingTalkH5CorpConfig{
		CorpID:        strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_CORP_ID")),
		CorpName:      strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_CORP_ID")),
		AppKey:        strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_KEY")),
		AppSecret:     appSecret,
		AgentID:       strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_AGENT_ID")),
		UnifiedAppID:  strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_UNIFIED_APP_ID")),
		AppURL:        strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_URL")),
		NotifyEnabled: boolToDingTalkH5SwitchInt(DingTalkH5NotificationEnabledContext(ctx)),
		NotifyMode:    normalizeDingTalkH5NotifyMode(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_NOTIFY_MODE"), dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_AGENT_ID"), dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_ROBOT_CODE")),
		RobotCode:     strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_ROBOT_CODE")),
		Enabled:       1,
		AppSecretSet:  appSecret != "",
	}
}

func LegacyDingTalkH5CorpConfigContext(ctx context.Context) DingTalkH5CorpConfig {
	return legacyDingTalkH5CorpConfigContext(ctx)
}

func mirrorPrimaryDingTalkH5CorpConfigToSetupContext(ctx context.Context, config DingTalkH5CorpConfig) error {
	setupItems := []setupservice.SetupItem{
		{Key: "DINGTALK_H5_CORP_ID", Value: config.CorpID, Type: "string"},
		{Key: "DINGTALK_H5_APP_KEY", Value: config.AppKey, Type: "string"},
		{Key: "DINGTALK_H5_AGENT_ID", Value: config.AgentID, Type: "string"},
		{Key: "DINGTALK_H5_UNIFIED_APP_ID", Value: config.UnifiedAppID, Type: "string"},
		{Key: "DINGTALK_H5_APP_URL", Value: config.AppURL, Type: "string"},
		{Key: "DINGTALK_H5_NOTIFY_ENABLED", Value: dingTalkH5SwitchIntString(config.NotifyEnabled), Type: "switch"},
		{Key: "DINGTALK_H5_NOTIFY_MODE", Value: config.NotifyMode, Type: "string"},
		{Key: "DINGTALK_H5_ROBOT_CODE", Value: config.RobotCode, Type: "string"},
	}
	if strings.TrimSpace(config.AppSecret) != "" {
		setupItems = append(setupItems, setupservice.SetupItem{Key: "DINGTALK_H5_APP_SECRET", Value: config.AppSecret, Type: "password"})
	}
	return setupservice.SetSetupsContext(ctx, setupItems, "")
}

func clearLegacyDingTalkH5CorpConfigContext(ctx context.Context) error {
	setupItems := []setupservice.SetupItem{
		{Key: "DINGTALK_H5_CORP_ID", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_APP_KEY", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_AGENT_ID", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_UNIFIED_APP_ID", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_NOTIFY_ENABLED", Value: "0", Type: "switch"},
		{Key: "DINGTALK_H5_NOTIFY_MODE", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_ROBOT_CODE", Value: "", Type: "string"},
		{Key: "DINGTALK_H5_APP_SECRET", Value: "", Type: "password"},
	}
	return setupservice.SetSetupsContext(ctx, setupItems, "")
}

func normalizeDingTalkH5NotifyMode(mode, agentID, robotCode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "robot", "agent", "agent_fallback":
		return mode
	case "agent-fallback":
		return "agent_fallback"
	case "fallback", "old_first", "old-first":
		return "agent_fallback"
	}
	if strings.TrimSpace(robotCode) != "" {
		return "robot"
	}
	if strings.TrimSpace(agentID) != "" {
		return "agent"
	}
	return "agent"
}

func NormalizeDingTalkH5CorpConfig(config DingTalkH5CorpConfig) DingTalkH5CorpConfig {
	return normalizeDingTalkH5CorpConfig(config)
}

func NormalizeDingTalkH5NotifyMode(mode, agentID, robotCode string) string {
	return normalizeDingTalkH5NotifyMode(mode, agentID, robotCode)
}

func normalizeDingTalkH5SwitchInt(value int) int {
	if value != 0 {
		return 1
	}
	return 0
}

func boolToDingTalkH5SwitchInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func dingTalkH5SwitchIntString(value int) string {
	if normalizeDingTalkH5SwitchInt(value) == 1 {
		return "1"
	}
	return "0"
}

func dingTalkH5SetupValueContext(ctx context.Context, key string) string {
	if database.GetDB() == nil {
		return ""
	}
	setup, err := setupservice.GetSetupContext(ctx, key)
	if err != nil || setup == nil {
		return ""
	}
	return setup.Value
}

func SetupValueContext(ctx context.Context, key string) string {
	return dingTalkH5SetupValueContext(ctx, key)
}

func DingTalkH5NotificationEnabledContext(ctx context.Context) bool {
	value := strings.ToLower(strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_NOTIFY_ENABLED")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
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

func IsMissingDingTalkH5UserBindingTableError(err error) bool {
	return isMissingDingTalkH5UserBindingTableError(err)
}
