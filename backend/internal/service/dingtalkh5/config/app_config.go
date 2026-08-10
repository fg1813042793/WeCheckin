package config

import (
	"context"
	"strings"
)

const (
	defaultDingTalkH5AppName  = "钉钉H5应用"
	defaultDingTalkH5LogoText = "H5"

	DefaultDingTalkH5AppName  = defaultDingTalkH5AppName
	DefaultDingTalkH5LogoText = defaultDingTalkH5LogoText
)

type DingTalkH5AppConfigDTO struct {
	AppTitle string `json:"appTitle"`
	AppName  string `json:"appName"`
	LogoText string `json:"logoText"`
	LogoURL  string `json:"logoUrl"`
	AppURL   string `json:"appUrl"`
}

type PublicConfigResponse struct {
	CorpID    string                 `json:"corpId"`
	CorpName  string                 `json:"corpName"`
	AppConfig DingTalkH5AppConfigDTO `json:"appConfig"`
	AppTitle  string                 `json:"appTitle"`
	AppName   string                 `json:"appName"`
	LogoText  string                 `json:"logoText"`
	LogoURL   string                 `json:"logoUrl"`
	AppURL    string                 `json:"appUrl"`
}

func AppConfigContext(ctx context.Context) DingTalkH5AppConfigDTO {
	appName := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_NAME"))
	logoText := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_LOGO_TEXT"))
	logoURL := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_LOGO_URL"))
	appURL := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_URL"))
	if appName == "" {
		appName = defaultDingTalkH5AppName
	}
	if logoText == "" {
		logoText = defaultDingTalkH5LogoText
	}
	return DingTalkH5AppConfigDTO{
		AppTitle: appName,
		AppName:  appName,
		LogoText: logoText,
		LogoURL:  logoURL,
		AppURL:   appURL,
	}
}

func PublicConfigContext(ctx context.Context) (*PublicConfigResponse, error) {
	appConfig := AppConfigContext(ctx)
	configs, err := ListDingTalkH5CorpConfigsContext(ctx)
	if err != nil {
		return nil, err
	}

	var corp DingTalkH5CorpConfig
	for _, config := range configs {
		if config.Enabled == 0 {
			continue
		}
		corp = config
		break
	}

	return &PublicConfigResponse{
		CorpID:    corp.CorpID,
		CorpName:  corp.CorpName,
		AppConfig: appConfig,
		AppTitle:  appConfig.AppTitle,
		AppName:   appConfig.AppName,
		LogoText:  appConfig.LogoText,
		LogoURL:   appConfig.LogoURL,
		AppURL:    appConfig.AppURL,
	}, nil
}
