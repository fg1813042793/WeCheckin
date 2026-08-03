package dingtalkh5

import (
	"context"
	"strings"
)

const (
	defaultDingTalkH5AppName  = "OA管理"
	defaultDingTalkH5LogoText = "OA"
)

func dingTalkH5AppConfigContext(ctx context.Context) DingTalkH5AppConfigDTO {
	appName := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_APP_NAME"))
	logoText := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_LOGO_TEXT"))
	logoURL := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, "DINGTALK_H5_LOGO_URL"))
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
	}
}

func PublicConfigContext(ctx context.Context) (*PublicConfigResponse, error) {
	appConfig := dingTalkH5AppConfigContext(ctx)
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
	}, nil
}
