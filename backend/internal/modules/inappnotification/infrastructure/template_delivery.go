package infrastructure

import (
	"net/url"
	"strconv"
	"strings"

	"wecheckin/backend/internal/modules/inappnotification/application"
	configsvc "wecheckin/backend/internal/service/dingtalkh5/config"
)

func dingTalkSendTemplatePayload(config configsvc.DingTalkH5CorpConfig, template application.DingTalkSendTemplate, title, content string) (configsvc.DingTalkWorkNotificationPayload, error) {
	if strings.TrimSpace(config.CorpID) == "" || (strings.TrimSpace(config.UnifiedAppID) == "" && strings.TrimSpace(config.AgentID) == "") {
		return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
	}
	targetURL := template.Target.URL
	switch template.Target.Kind {
	case "app_home":
		targetURL = strings.TrimSpace(config.AppURL)
		if !validTemplateAppURL(targetURL) {
			return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
		}
	case "workflow_start":
		baseURL := strings.TrimSpace(config.AppURL)
		if !validTemplateAppURL(baseURL) {
			return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
		}
		parsed, _ := url.Parse(baseURL)
		query, err := url.ParseQuery(parsed.RawQuery)
		if err != nil {
			return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
		}
		for _, key := range []string{"view", "key", "page"} {
			query.Del(key)
		}
		workflowView := "workflow:start:" + strconv.FormatUint(uint64(template.Target.DefinitionID), 10)
		query.Set("corpId", strings.TrimSpace(config.CorpID))
		query.Set("view", workflowView)
		parsed.RawQuery = query.Encode()
		parsed.Fragment = "/pages/index/index"
		parsed.RawFragment = ""
		if template.Target.DefinitionID == 0 {
			return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateWorkflowUnavailable
		}
		parsed.Fragment += "?view=" + workflowView
		targetURL = parsed.String()
	case "url":
		if !application.ValidTemplateHTTPSURL(targetURL) {
			return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
		}
	default:
		return configsvc.DingTalkWorkNotificationPayload{}, application.ErrTemplateTargetUnavailable
	}
	return configsvc.DingTalkWorkNotificationPayload{Title: title, Content: content, MessageType: template.MessageType,
		URL: configsvc.WrapDingTalkOpenAppURL(targetURL, config), ButtonTitle: template.ButtonTitle, HeadColor: template.HeadColor, SourceName: "WeCheckin 通知"}, nil
}

// Existing enterprise app configuration permits HTTP for local installations; custom template URLs require HTTPS.

func validTemplateAppURL(value string) bool {
	if strings.HasPrefix(value, "http://") {
		return application.ValidTemplateHTTPSURL("https://" + strings.TrimPrefix(value, "http://"))
	}
	return application.ValidTemplateHTTPSURL(value)
}
