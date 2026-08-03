package dingtalk

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	dingtalkh5service "wecheckin/backend/internal/app/service/dingtalkh5"
	setupservice "wecheckin/backend/internal/app/service/setup"
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
	if len(corpConfigs) > 0 {
		corpID = corpConfigs[0].CorpID
		appKey = corpConfigs[0].AppKey
		agentID = corpConfigs[0].AgentID
		appSecret = ""
		if corpConfigs[0].AppSecretSet {
			appSecret = "set"
		}
	}
	response.JSON(c, map[string]interface{}{
		"corpId":        corpID,
		"appKey":        appKey,
		"agentId":       agentID,
		"appSecretSet":  appSecret != "",
		"corpConfigs":   adminDingTalkCorpConfigResponses(corpConfigs),
		"tokenExpire":   expire.String(),
		"redisPrefix":   prefix,
		"singleLogin":   singleLogin,
		"selfBind":      boolToSwitch(dingtalkh5service.SelfBindEnabledContext(ctx)),
		"notifyEnabled": boolToSwitch(dingtalkh5service.DingTalkH5NotificationEnabledContext(ctx)),
		"appName":       adminDingTalkSetupValueDefault(ctx, "DINGTALK_H5_APP_NAME", defaultDingTalkH5AppName),
		"logoText":      adminDingTalkSetupValueDefault(ctx, "DINGTALK_H5_LOGO_TEXT", defaultDingTalkH5LogoText),
		"logoUrl":       adminDingTalkSetupValue(ctx, "DINGTALK_H5_LOGO_URL"),
	})
}

func (h *AdminDingTalkHandler) SaveSettings(ctx context.Context, c *app.RequestContext) {
	corpID := strings.TrimSpace(c.PostForm("corpId"))
	appKey := strings.TrimSpace(c.PostForm("appKey"))
	appSecret := strings.TrimSpace(c.PostForm("appSecret"))
	agentID := strings.TrimSpace(c.PostForm("agentId"))
	corpConfigs, err := parseDingTalkCorpConfigInputs(c.PostForm("corpConfigs"))
	if err != nil {
		response.Fail(c, "企业配置格式错误")
		return
	}
	if len(corpConfigs) == 0 && (corpID != "" || appKey != "" || appSecret != "") {
		corpConfigs = []dingtalkh5service.DingTalkH5CorpConfig{{
			CorpID:    corpID,
			CorpName:  corpID,
			AppKey:    appKey,
			AppSecret: appSecret,
			AgentID:   agentID,
			Enabled:   1,
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
	}
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
	notifyEnabled := "0"
	if value := strings.TrimSpace(c.PostForm("notifyEnabled")); value == "1" || strings.EqualFold(value, "true") {
		notifyEnabled = "1"
	}
	appName := strings.TrimSpace(c.PostForm("appName"))
	if appName == "" {
		appName = defaultDingTalkH5AppName
	}
	logoText := strings.TrimSpace(c.PostForm("logoText"))
	if logoText == "" {
		logoText = defaultDingTalkH5LogoText
	}
	logoURL := strings.TrimSpace(c.PostForm("logoUrl"))

	addIP := c.ClientIP()
	for _, item := range []struct {
		key   string
		value string
		typ   string
	}{
		{key: "DINGTALK_H5_CORP_ID", value: corpID, typ: "string"},
		{key: "DINGTALK_H5_APP_KEY", value: appKey, typ: "string"},
		{key: "DINGTALK_H5_AGENT_ID", value: agentID, typ: "string"},
		{key: "TOKEN_DINGTALK_H5_EXPIRE", value: tokenExpire, typ: "string"},
		{key: "TOKEN_DINGTALK_H5_REDIS_PREFIX", value: redisPrefix, typ: "string"},
		{key: "DINGTALK_H5_SINGLE_LOGIN", value: singleLogin, typ: "switch"},
		{key: "DINGTALK_H5_SELF_BIND_ENABLED", value: selfBind, typ: "switch"},
		{key: "DINGTALK_H5_NOTIFY_ENABLED", value: notifyEnabled, typ: "switch"},
		{key: "DINGTALK_H5_APP_NAME", value: appName, typ: "string"},
		{key: "DINGTALK_H5_LOGO_TEXT", value: logoText, typ: "string"},
		{key: "DINGTALK_H5_LOGO_URL", value: logoURL, typ: "string"},
	} {
		if err := setupservice.SetSetupContext(ctx, item.key, item.value, item.typ, addIP); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	}
	if appSecret != "" {
		if err := setupservice.SetSetupContext(ctx, "DINGTALK_H5_APP_SECRET", appSecret, "password", addIP); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	}
	if len(corpConfigs) > 0 {
		if err := saveDingTalkH5CorpConfigsContext(ctx, corpConfigs); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	}
	tokenutil.InvalidateSetupCache()
	response.JSON(c, nil)
}

type dingTalkCorpConfigInput struct {
	CorpID    string `json:"corpId"`
	CorpName  string `json:"corpName"`
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
	AgentID   string `json:"agentId"`
	Enabled   int    `json:"enabled"`
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
		configs = append(configs, dingtalkh5service.DingTalkH5CorpConfig{
			CorpID:    corpID,
			CorpName:  strings.TrimSpace(input.CorpName),
			AppKey:    strings.TrimSpace(input.AppKey),
			AppSecret: strings.TrimSpace(input.AppSecret),
			AgentID:   strings.TrimSpace(input.AgentID),
			Enabled:   enabled,
		})
	}
	return configs, nil
}

func adminDingTalkCorpConfigResponses(configs []dingtalkh5service.DingTalkH5CorpConfig) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		items = append(items, map[string]interface{}{
			"corpId":       config.CorpID,
			"corpName":     config.CorpName,
			"appKey":       config.AppKey,
			"agentId":      config.AgentID,
			"enabled":      config.Enabled,
			"appSecretSet": config.AppSecretSet,
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
