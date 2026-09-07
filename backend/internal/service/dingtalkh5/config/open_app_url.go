package config

import (
	"net/url"
	"strings"
)

// WrapDingTalkOpenAppURL opens an H5 operation URL inside the configured DingTalk workbench app.
func WrapDingTalkOpenAppURL(operationURL string, config DingTalkH5CorpConfig) string {
	operationURL = strings.TrimSpace(operationURL)
	if operationURL == "" {
		return ""
	}
	corpID := strings.TrimSpace(config.CorpID)
	appID := strings.TrimSpace(config.UnifiedAppID)
	if appID == "" {
		appID = strings.TrimSpace(config.AgentID)
		if appID != "" && !strings.HasPrefix(appID, "0_") {
			appID = "0_" + appID
		}
	}
	if corpID == "" || appID == "" {
		return operationURL
	}
	query := url.Values{}
	query.Set("corpid", corpID)
	query.Set("container_type", "work_platform")
	query.Set("app_id", appID)
	query.Set("redirect_type", "jump")
	query.Set("redirect_url", operationURL)
	return (&url.URL{
		Scheme:   "dingtalk",
		Host:     "dingtalkclient",
		Path:     "/action/openapp",
		RawQuery: query.Encode(),
	}).String()
}
