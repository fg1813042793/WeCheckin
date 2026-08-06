package dingtalk

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/performance"
	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/pkg/response"
	"wecheckin/backend/pkg/tokenutil"
)

const (
	defaultDingTalkH5TokenExpire = "168h"
	defaultDingTalkH5RedisPrefix = "dingtalk_h5_token:"
	defaultDingTalkH5AppName     = "OA管理"
	defaultDingTalkH5LogoText    = "OA"
)

type AdminDingTalkHandler struct{}

func NewAdminDingTalkHandler() *AdminDingTalkHandler { return &AdminDingTalkHandler{} }

func (h *AdminDingTalkHandler) GetSettings(ctx context.Context, c *app.RequestContext) {
	corpConfigs, err := listDingTalkH5CorpConfigsContext(ctx)
	if err != nil {
		response.Fail(c, "读取配置失败")
		return
	}
	expire, prefix := tokenutil.GetTokenConfig("dingtalk_h5")
	singleLogin := 0
	if tokenutil.IsDingTalkH5SingleLogin() {
		singleLogin = 1
	}
	appSecret := adminDingTalkSetupValue(ctx, "DINGTALK_H5_APP_SECRET")
	corpID := adminDingTalkSetupValue(ctx, "DINGTALK_H5_CORP_ID")
	appKey := adminDingTalkSetupValue(ctx, "DINGTALK_H5_APP_KEY")
	agentID := adminDingTalkSetupValue(ctx, "DINGTALK_H5_AGENT_ID")
	unifiedAppID := adminDingTalkSetupValue(ctx, "DINGTALK_H5_UNIFIED_APP_ID")
	notifyMode := adminDingTalkSetupValueDefault(ctx, "DINGTALK_H5_NOTIFY_MODE", "agent")
	notifyEnabled := boolToSwitch(dingtalkh5service.DingTalkH5NotificationEnabledContext(ctx))
	robotCode := adminDingTalkSetupValue(ctx, "DINGTALK_H5_ROBOT_CODE")
	if len(corpConfigs) > 0 {
		corpID = corpConfigs[0].CorpID
		appKey = corpConfigs[0].AppKey
		agentID = corpConfigs[0].AgentID
		unifiedAppID = corpConfigs[0].UnifiedAppID
		notifyMode = corpConfigs[0].NotifyMode
		notifyEnabled = corpConfigs[0].NotifyEnabled
		robotCode = corpConfigs[0].RobotCode
		appSecret = ""
		if corpConfigs[0].AppSecretSet {
			appSecret = "set"
		}
	}
	response.JSON(c, map[string]interface{}{
		"corpId":        corpID,
		"appKey":        appKey,
		"agentId":       agentID,
		"unifiedAppId":  unifiedAppID,
		"notifyMode":    notifyMode,
		"robotCode":     robotCode,
		"appSecretSet":  appSecret != "",
		"corpConfigs":   adminDingTalkCorpConfigResponses(corpConfigs),
		"tokenExpire":   expire.String(),
		"redisPrefix":   prefix,
		"singleLogin":   singleLogin,
		"selfBind":      boolToSwitch(dingtalkh5service.SelfBindEnabledContext(ctx)),
		"notifyEnabled": notifyEnabled,
		"appName":       adminDingTalkSetupValueDefault(ctx, "DINGTALK_H5_APP_NAME", defaultDingTalkH5AppName),
		"logoText":      adminDingTalkSetupValueDefault(ctx, "DINGTALK_H5_LOGO_TEXT", defaultDingTalkH5LogoText),
		"logoUrl":       adminDingTalkSetupValue(ctx, "DINGTALK_H5_LOGO_URL"),
		"appUrl":        adminDingTalkSetupValue(ctx, "DINGTALK_H5_APP_URL"),
	})
}

func (h *AdminDingTalkHandler) SaveSettings(ctx context.Context, c *app.RequestContext) {
	saveScope := normalizeDingTalkSettingsSaveScope(c.PostForm("scope"))
	if saveScope == "" {
		response.Fail(c, "保存范围错误")
		return
	}

	addIP := c.ClientIP()
	setupItems := []setupservice.SetupItem{}
	var corpConfigs []dingtalkh5service.DingTalkH5CorpConfig

	if dingTalkSettingsSaveScopeIncludes(saveScope, "login") {
		tokenExpire := strings.TrimSpace(c.PostForm("tokenExpire"))
		if tokenExpire == "" {
			tokenExpire = defaultDingTalkH5TokenExpire
		}
		if _, err := time.ParseDuration(tokenExpire); err != nil {
			response.Fail(c, "Token 过期时间格式错误")
			return
		}
		redisPrefix := strings.TrimSpace(c.PostForm("redisPrefix"))
		if redisPrefix == "" {
			redisPrefix = defaultDingTalkH5RedisPrefix
		}
		singleLogin := "0"
		if value := strings.TrimSpace(c.PostForm("singleLogin")); value == "1" || strings.EqualFold(value, "true") {
			singleLogin = "1"
		}
		selfBind := "0"
		if value := strings.TrimSpace(c.PostForm("selfBind")); value == "" || value == "1" || strings.EqualFold(value, "true") {
			selfBind = "1"
		}
		setupItems = append(setupItems,
			setupservice.SetupItem{Key: "TOKEN_DINGTALK_H5_EXPIRE", Value: tokenExpire, Type: "string"},
			setupservice.SetupItem{Key: "TOKEN_DINGTALK_H5_REDIS_PREFIX", Value: redisPrefix, Type: "string"},
			setupservice.SetupItem{Key: "DINGTALK_H5_SINGLE_LOGIN", Value: singleLogin, Type: "switch"},
			setupservice.SetupItem{Key: "DINGTALK_H5_SELF_BIND_ENABLED", Value: selfBind, Type: "switch"},
		)
	}

	if dingTalkSettingsSaveScopeIncludes(saveScope, "app") {
		appName := strings.TrimSpace(c.PostForm("appName"))
		if appName == "" {
			appName = defaultDingTalkH5AppName
		}
		logoText := strings.TrimSpace(c.PostForm("logoText"))
		if logoText == "" {
			logoText = defaultDingTalkH5LogoText
		}
		logoURL := strings.TrimSpace(c.PostForm("logoUrl"))
		appURL := strings.TrimSpace(c.PostForm("appUrl"))
		setupItems = append(setupItems,
			setupservice.SetupItem{Key: "DINGTALK_H5_APP_NAME", Value: appName, Type: "string"},
			setupservice.SetupItem{Key: "DINGTALK_H5_LOGO_TEXT", Value: logoText, Type: "string"},
			setupservice.SetupItem{Key: "DINGTALK_H5_LOGO_URL", Value: logoURL, Type: "string"},
			setupservice.SetupItem{Key: "DINGTALK_H5_APP_URL", Value: appURL, Type: "string"},
		)
	}

	if dingTalkSettingsSaveScopeIncludes(saveScope, "corp") {
		corpID := strings.TrimSpace(c.PostForm("corpId"))
		appKey := strings.TrimSpace(c.PostForm("appKey"))
		appSecret := strings.TrimSpace(c.PostForm("appSecret"))
		agentID := strings.TrimSpace(c.PostForm("agentId"))
		unifiedAppID := strings.TrimSpace(c.PostForm("unifiedAppId"))
		notifyMode := strings.TrimSpace(c.PostForm("notifyMode"))
		robotCode := strings.TrimSpace(c.PostForm("robotCode"))
		notifyEnabled := switchPostFormInt(c.PostForm("notifyEnabled"))
		var err error
		corpConfigs, err = parseDingTalkCorpConfigInputs(c.PostForm("corpConfigs"))
		if err != nil {
			response.Fail(c, "企业配置格式错误")
			return
		}
		if len(corpConfigs) == 0 && (corpID != "" || appKey != "" || appSecret != "") {
			corpConfigs = []dingtalkh5service.DingTalkH5CorpConfig{{
				CorpID:        corpID,
				CorpName:      corpID,
				AppKey:        appKey,
				AppSecret:     appSecret,
				AgentID:       agentID,
				UnifiedAppID:  unifiedAppID,
				NotifyEnabled: notifyEnabled,
				NotifyMode:    notifyMode,
				RobotCode:     robotCode,
				Enabled:       1,
			}}
		}
		if len(corpConfigs) > 0 {
			if corpID == "" {
				corpID = corpConfigs[0].CorpID
			}
			if appKey == "" {
				appKey = corpConfigs[0].AppKey
			}
			if appSecret == "" {
				appSecret = corpConfigs[0].AppSecret
			}
			if agentID == "" {
				agentID = corpConfigs[0].AgentID
			}
			if unifiedAppID == "" {
				unifiedAppID = corpConfigs[0].UnifiedAppID
			}
			if notifyMode == "" {
				notifyMode = corpConfigs[0].NotifyMode
			}
			if robotCode == "" {
				robotCode = corpConfigs[0].RobotCode
			}
			notifyEnabled = corpConfigs[0].NotifyEnabled
		}
		if notifyMode == "" {
			notifyMode = "agent"
		}
		if len(corpConfigs) == 0 {
			setupItems = append(setupItems,
				setupservice.SetupItem{Key: "DINGTALK_H5_CORP_ID", Value: corpID, Type: "string"},
				setupservice.SetupItem{Key: "DINGTALK_H5_APP_KEY", Value: appKey, Type: "string"},
				setupservice.SetupItem{Key: "DINGTALK_H5_AGENT_ID", Value: agentID, Type: "string"},
				setupservice.SetupItem{Key: "DINGTALK_H5_UNIFIED_APP_ID", Value: unifiedAppID, Type: "string"},
				setupservice.SetupItem{Key: "DINGTALK_H5_NOTIFY_ENABLED", Value: switchIntString(notifyEnabled), Type: "switch"},
				setupservice.SetupItem{Key: "DINGTALK_H5_NOTIFY_MODE", Value: notifyMode, Type: "string"},
				setupservice.SetupItem{Key: "DINGTALK_H5_ROBOT_CODE", Value: robotCode, Type: "string"},
			)
			if appSecret != "" {
				setupItems = append(setupItems, setupservice.SetupItem{Key: "DINGTALK_H5_APP_SECRET", Value: appSecret, Type: "password"})
			}
		}
	}

	if err := setupservice.SetSetupsContext(ctx, setupItems, addIP); err != nil {
		response.Fail(c, "保存失败")
		return
	}
	shouldSaveCorpConfigs := saveScope == "corp" || len(corpConfigs) > 0
	if shouldSaveCorpConfigs {
		if err := saveDingTalkH5CorpConfigsContext(ctx, corpConfigs); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	}
	tokenutil.InvalidateSetupCache()
	response.JSON(c, nil)
}

func (h *AdminDingTalkHandler) TestNotification(ctx context.Context, c *app.RequestContext) {
	corpID := strings.TrimSpace(c.PostForm("corpId"))
	recipientUserID := strings.TrimSpace(c.PostForm("dingTalkUserId"))
	if corpID == "" {
		response.Fail(c, "请选择企业应用")
		return
	}
	result, err := dingtalkh5service.DiagnoseDingTalkH5WorkNotificationContext(ctx, corpID, recipientUserID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, result)
}

func normalizeDingTalkSettingsSaveScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "", "all":
		return "all"
	case "corp", "login", "app":
		return strings.ToLower(strings.TrimSpace(scope))
	default:
		return ""
	}
}

func dingTalkSettingsSaveScopeIncludes(scope, target string) bool {
	return scope == "all" || scope == target
}

type dingTalkCorpConfigInput struct {
	CorpID        string `json:"corpId"`
	CorpName      string `json:"corpName"`
	AppKey        string `json:"appKey"`
	AppSecret     string `json:"appSecret"`
	AgentID       string `json:"agentId"`
	UnifiedAppID  string `json:"unifiedAppId"`
	AppURL        string `json:"appUrl"`
	NotifyEnabled int    `json:"notifyEnabled"`
	NotifyMode    string `json:"notifyMode"`
	RobotCode     string `json:"robotCode"`
	Enabled       int    `json:"enabled"`
}

func parseDingTalkCorpConfigInputs(raw string) ([]dingtalkh5service.DingTalkH5CorpConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var inputs []dingTalkCorpConfigInput
	if err := json.Unmarshal([]byte(raw), &inputs); err != nil {
		return nil, err
	}
	configs := make([]dingtalkh5service.DingTalkH5CorpConfig, 0, len(inputs))
	for _, input := range inputs {
		corpID := strings.TrimSpace(input.CorpID)
		if corpID == "" {
			continue
		}
		enabled := input.Enabled
		if enabled != 0 {
			enabled = 1
		}
		notifyEnabled := input.NotifyEnabled
		if notifyEnabled != 0 {
			notifyEnabled = 1
		}
		configs = append(configs, dingtalkh5service.DingTalkH5CorpConfig{
			CorpID:        corpID,
			CorpName:      strings.TrimSpace(input.CorpName),
			AppKey:        strings.TrimSpace(input.AppKey),
			AppSecret:     strings.TrimSpace(input.AppSecret),
			AgentID:       strings.TrimSpace(input.AgentID),
			UnifiedAppID:  strings.TrimSpace(input.UnifiedAppID),
			AppURL:        strings.TrimSpace(input.AppURL),
			NotifyEnabled: notifyEnabled,
			NotifyMode:    strings.TrimSpace(input.NotifyMode),
			RobotCode:     strings.TrimSpace(input.RobotCode),
			Enabled:       enabled,
		})
	}
	return configs, nil
}

func adminDingTalkCorpConfigResponses(configs []dingtalkh5service.DingTalkH5CorpConfig) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		items = append(items, map[string]interface{}{
			"corpId":        config.CorpID,
			"corpName":      config.CorpName,
			"appKey":        config.AppKey,
			"agentId":       config.AgentID,
			"unifiedAppId":  config.UnifiedAppID,
			"appUrl":        config.AppURL,
			"notifyEnabled": config.NotifyEnabled,
			"notifyMode":    config.NotifyMode,
			"robotCode":     config.RobotCode,
			"enabled":       config.Enabled,
			"appSecretSet":  config.AppSecretSet,
		})
	}
	return items
}

func listDingTalkH5CorpConfigsContext(ctx context.Context) ([]dingtalkh5service.DingTalkH5CorpConfig, error) {
	return dingtalkh5service.ListDingTalkH5CorpConfigsContext(ctx)
}

func saveDingTalkH5CorpConfigsContext(ctx context.Context, configs []dingtalkh5service.DingTalkH5CorpConfig) error {
	return dingtalkh5service.SaveDingTalkH5CorpConfigsContext(ctx, configs)
}

func boolToSwitch(value bool) int {
	if value {
		return 1
	}
	return 0
}

func switchPostFormInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "1" || strings.EqualFold(value, "true") {
		return 1
	}
	return 0
}

func switchIntString(value int) string {
	if value != 0 {
		return "1"
	}
	return "0"
}

func adminDingTalkSetupValue(ctx context.Context, key string) string {
	setup, err := setupservice.GetSetupContext(ctx, key)
	if err != nil || setup == nil {
		return ""
	}
	return setup.Value
}

func adminDingTalkSetupValueDefault(ctx context.Context, key, fallback string) string {
	value := strings.TrimSpace(adminDingTalkSetupValue(ctx, key))
	if value == "" {
		return fallback
	}
	return value
}
