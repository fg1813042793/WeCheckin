package dingtalk

import (
	"os"
	"strings"
	"testing"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/config"
)

func TestDingTalkSettingsExposeSSOConfigWithoutSecret(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	dtoSrc, err := os.ReadFile("../../../service/admin/dingtalk/dto.go")
	if err != nil {
		t.Fatalf("read dingtalk dto.go: %v", err)
	}
	text := string(src) + string(dtoSrc)
	for _, want := range []string{
		"DINGTALK_H5_CORP_ID",
		"DINGTALK_H5_APP_KEY",
		"appSecretSet",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk settings should include SSO config marker %q", want)
		}
	}
	if strings.Contains(text, `"appSecret":`) {
		t.Fatalf("dingtalk settings must not return raw appSecret")
	}
}

func TestDingTalkSettingsSaveKeepsExistingSecretWhenEmpty(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		`appSecret := strings.TrimSpace(c.PostForm("appSecret"))`,
		`if appSecret != ""`,
		`DINGTALK_H5_APP_SECRET`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk settings save should preserve appSecret with %q", want)
		}
	}
}

func TestDingTalkSettingsExposeMultiCorpConfigsWithoutRawSecrets(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	dtoSrc, err := os.ReadFile("../../../service/admin/dingtalk/dto.go")
	if err != nil {
		t.Fatalf("read dingtalk dto.go: %v", err)
	}
	text := string(src) + string(dtoSrc)
	for _, want := range []string{
		"corpConfigs",
		"listDingTalkH5CorpConfigsContext",
		"saveDingTalkH5CorpConfigsContext",
		"appSecretSet",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk settings should support multi corp configs with %q", want)
		}
	}
	if strings.Contains(text, `"appSecret": config.AppSecret`) {
		t.Fatalf("dingtalk settings must not return raw corp app secrets")
	}
}

func TestDingTalkSettingsExposeH5BrandConfigWithoutEditableGlobalURL(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	dtoSrc, err := os.ReadFile("../../../service/admin/dingtalk/dto.go")
	if err != nil {
		t.Fatalf("read dingtalk dto.go: %v", err)
	}
	combined := string(src) + string(dtoSrc) + string(viewSrc)
	for _, want := range []string{
		"DINGTALK_H5_APP_NAME",
		"DINGTALK_H5_LOGO_TEXT",
		"DINGTALK_H5_LOGO_URL",
		`json:"appName"`,
		`json:"logoText"`,
		`json:"logoUrl"`,
		`<el-tabs v-model="activeTab"`,
		`<el-tab-pane label="配置" name="app">`,
		`form.appName`,
		`form.logoText`,
		`form.logoUrl`,
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("dingtalk settings should expose configurable H5 branding with %q", want)
		}
	}
	viewText := string(viewSrc)
	appStart := strings.Index(viewText, `<el-tab-pane label="配置" name="app">`)
	if appStart < 0 {
		t.Fatalf("dingtalk setup app tab not found")
	}
	appEnd := strings.Index(viewText[appStart:], `</el-tab-pane>`)
	if appEnd < 0 {
		t.Fatalf("dingtalk setup app tab end not found")
	}
	appPane := viewText[appStart : appStart+appEnd]
	for _, forbidden := range []string{`label="H5 应用地址"`, `form.appUrl`} {
		if strings.Contains(appPane, forbidden) {
			t.Fatalf("global app config must not expose an editable H5 URL with %q", forbidden)
		}
	}
}

func TestDingTalkSettingsSeparatesCorpConfigFromLoginConfig(t *testing.T) {
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	text := string(viewSrc)
	for _, want := range []string{
		`<el-tab-pane label="企业应用" name="corp">`,
		`<el-tab-pane label="登录配置" name="login">`,
		`<el-tab-pane label="配置" name="app">`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk setup should render separate tab %q", want)
		}
	}
	loginStart := strings.Index(text, `<el-tab-pane label="登录配置" name="login">`)
	appStart := strings.Index(text, `<el-tab-pane label="配置" name="app">`)
	if loginStart < 0 || appStart < 0 || appStart <= loginStart {
		t.Fatalf("dingtalk setup login tab should be before app config tab")
	}
	loginPane := text[loginStart:appStart]
	if strings.Contains(loginPane, `label="企业应用"`) || strings.Contains(loginPane, `class="corp-config-list"`) {
		t.Fatalf("dingtalk setup login tab should not contain enterprise application form")
	}
}

func TestDingTalkSettingsCorpDeleteRequiresConfirmation(t *testing.T) {
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	text := string(viewSrc)
	for _, want := range []string{
		`import { ElMessage, ElMessageBox } from 'element-plus'`,
		`async function removeCorpConfig(index: number)`,
		`await ElMessageBox.confirm('确定删除该企业应用？', '提示', { type: 'warning' })`,
		`form.corpConfigs.splice(index, 1)`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk setup corp delete should require confirmation with %q", want)
		}
	}
}

func TestDingTalkSettingsExposeNotificationConfig(t *testing.T) {
	handlerSrc, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	combined := string(handlerSrc) + string(viewSrc)
	for _, want := range []string{
		"DINGTALK_H5_NOTIFY_ENABLED",
		"DINGTALK_H5_NOTIFY_MODE",
		"DINGTALK_H5_ROBOT_CODE",
		"DINGTALK_H5_UNIFIED_APP_ID",
		`NotifyEnabled`,
		"notifyEnabled",
		"notifyMode",
		"robotCode",
		"unifiedAppId",
		"agentId",
		"corpConfig.notifyEnabled",
		"notifyEnabled: item.notifyEnabled",
		"钉钉通知",
		"流程提醒和管理后台手动通知",
		"默认通知方式",
		"App ID 只用于通知点击打开应用",
		"AgentId + OA",
		"agent_fallback",
		"旧版优先，失败兜底新版",
		"sampleLink",
		"新版机器人通知",
		"旧版工作通知",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("dingtalk settings should expose notification config with %q", want)
		}
	}
	if strings.Contains(combined, "form.notifyEnabled") {
		t.Fatalf("performance notification switch should belong to enterprise application configs, not global app config")
	}
	viewText := string(viewSrc)
	appStart := strings.Index(viewText, `<el-tab-pane label="配置" name="app">`)
	if appStart < 0 {
		t.Fatalf("dingtalk setup app tab not found")
	}
	appPane := viewText[appStart:]
	if strings.Contains(appPane, "corpConfig.notifyEnabled") {
		t.Fatalf("app brand config tab should not contain enterprise notification switch")
	}
}

func TestDingTalkSettingsEnterpriseAppCarriesNotificationJumpURL(t *testing.T) {
	handlerSrc, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	dtoSrc, err := os.ReadFile("../../../service/admin/dingtalk/dto.go")
	if err != nil {
		t.Fatalf("read dingtalk dto.go: %v", err)
	}
	combined := string(handlerSrc) + string(dtoSrc) + string(viewSrc)
	for _, want := range []string{
		`AppURL        string ` + "`json:\"appUrl\"`",
		`AppURL:        strings.TrimSpace(input.AppURL)`,
		`AppURL: config.AppURL`,
		`corpConfig.appUrl`,
		`appUrl: item.appUrl.trim()`,
		`H5 应用地址`,
		`例如 https://.../?corpId=...`,
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("enterprise app config should carry notification jump URL with %q", want)
		}
	}
}

func TestDingTalkSettingsSaveUsesBatchedSetupWrites(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"setupservice.SetupItem",
		"setupservice.SetSetupsContext(ctx, setupItems, addIP)",
		"saveDingTalkH5CorpConfigsContext(ctx, corpConfigs)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk settings save should use batched setup writes with %q", want)
		}
	}
	if strings.Contains(text, "setupservice.SetSetupContext(ctx, item.key") {
		t.Fatalf("dingtalk settings save should not write setup keys one by one")
	}
}

func TestDingTalkSettingsSaveOnlyPersistsActiveTabScope(t *testing.T) {
	handlerSrc, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	handlerText := string(handlerSrc)
	viewText := string(viewSrc)
	for _, want := range []string{
		`saveScope := normalizeDingTalkSettingsSaveScope(c.PostForm("scope"))`,
		`if dingTalkSettingsSaveScopeIncludes(saveScope, "login") {`,
		`if dingTalkSettingsSaveScopeIncludes(saveScope, "app") {`,
		`if dingTalkSettingsSaveScopeIncludes(saveScope, "corp") {`,
		`func normalizeDingTalkSettingsSaveScope(scope string) string`,
		`func dingTalkSettingsSaveScopeIncludes(scope, target string) bool`,
	} {
		if !strings.Contains(handlerText, want) {
			t.Fatalf("dingtalk settings save should scope backend writes with %q", want)
		}
	}
	if strings.Contains(handlerText, `setupItems := []setupservice.SetupItem{
		{Key: "TOKEN_DINGTALK_H5_EXPIRE"`) {
		t.Fatalf("dingtalk settings save should not build one unconditional setup payload for all tabs")
	}
	appScopeStart := strings.Index(handlerText, `if dingTalkSettingsSaveScopeIncludes(saveScope, "app") {`)
	corpScopeStart := strings.Index(handlerText, `if dingTalkSettingsSaveScopeIncludes(saveScope, "corp") {`)
	if appScopeStart < 0 || corpScopeStart < 0 || corpScopeStart <= appScopeStart {
		t.Fatalf("dingtalk settings handler should define app scope before corp scope")
	}
	if strings.Contains(handlerText[appScopeStart:corpScopeStart], "DINGTALK_H5_APP_URL") {
		t.Fatalf("app brand settings must not overwrite the enterprise H5 URL compatibility key")
	}

	for _, want := range []string{
		`function buildSettingsPayload()`,
		`switch (activeTab.value)`,
		`scope: 'corp'`,
		`scope: 'login'`,
		`scope: 'app'`,
	} {
		if !strings.Contains(viewText, want) {
			t.Fatalf("dingtalk setup page should submit only active tab scope with %q", want)
		}
	}
	loginStart := strings.Index(viewText, `case 'login':`)
	appStart := strings.Index(viewText, `case 'app':`)
	if loginStart < 0 || appStart < 0 || appStart <= loginStart {
		t.Fatalf("dingtalk setup page should define separate login and app payload cases")
	}
	loginPayload := viewText[loginStart:appStart]
	for _, forbidden := range []string{"corpConfigs", "notifyEnabled", "appName", "logoText", "logoUrl", "appUrl"} {
		if strings.Contains(loginPayload, forbidden) {
			t.Fatalf("login settings payload should not include %s", forbidden)
		}
	}
	defaultStart := strings.Index(viewText, `default:`)
	if defaultStart < 0 || defaultStart <= appStart {
		t.Fatalf("dingtalk setup page should define app payload before default case")
	}
	appPayload := viewText[appStart:defaultStart]
	if strings.Contains(appPayload, "notifyEnabled") {
		t.Fatalf("app brand settings payload should not include notifyEnabled")
	}
	if strings.Contains(appPayload, "appUrl") {
		t.Fatalf("app brand settings payload should not include the enterprise H5 operation URL")
	}
}

func TestValidateDingTalkCorpConfigsRequiresURLForEnabledApps(t *testing.T) {
	tests := []struct {
		name    string
		configs []dingtalkh5service.DingTalkH5CorpConfig
		wantErr bool
	}{
		{name: "empty list may clear all enterprise apps"},
		{
			name: "enabled enterprise requires H5 URL",
			configs: []dingtalkh5service.DingTalkH5CorpConfig{{
				CorpID: "ding-enabled", Enabled: 1,
			}},
			wantErr: true,
		},
		{
			name: "disabled enterprise may omit H5 URL",
			configs: []dingtalkh5service.DingTalkH5CorpConfig{{
				CorpID: "ding-disabled", Enabled: 0,
			}},
		},
		{
			name: "enabled enterprise accepts H5 URL",
			configs: []dingtalkh5service.DingTalkH5CorpConfig{{
				CorpID: "ding-enabled", AppURL: "https://oa.example.com/?corpId=ding-enabled", Enabled: 1,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDingTalkCorpConfigs(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDingTalkCorpConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDingTalkSettingsCorpScopeAllowsClearingCorpConfigs(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		`shouldSaveCorpConfigs := saveScope == "corp" || len(corpConfigs) > 0`,
		`if shouldSaveCorpConfigs {`,
		`saveDingTalkH5CorpConfigsContext(ctx, corpConfigs)`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("corp settings save should allow clearing corp configs with %q", want)
		}
	}
}

func TestDingTalkSettingsExposeNotificationDiagnosisAction(t *testing.T) {
	handlerSrc, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	handlerText := string(handlerSrc)
	viewText := string(viewSrc)
	for _, want := range []string{
		"func (h *AdminDingTalkHandler) TestNotification",
		"DiagnoseDingTalkH5WorkNotificationContext(ctx, corpID, recipientUserID)",
	} {
		if !strings.Contains(handlerText, want) {
			t.Fatalf("dingtalk handler should expose notification diagnosis action with %q", want)
		}
	}
	for _, want := range []string{
		"测试通知",
		"testCorpNotification(corpConfig)",
		"/api/v2/admin/dingtalk/settings/notification-test",
		"通知诊断结果",
		"调用链路",
		"复制证据",
		"diagnosisResult",
	} {
		if !strings.Contains(viewText, want) {
			t.Fatalf("dingtalk setup page should expose notification diagnosis UI with %q", want)
		}
	}
}
