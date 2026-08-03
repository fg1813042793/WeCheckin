package dingtalk

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkSettingsExposeSSOConfigWithoutSecret(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
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
	text := string(src)
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

func TestDingTalkSettingsExposeH5BrandConfig(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	combined := string(src) + string(viewSrc)
	for _, want := range []string{
		"DINGTALK_H5_APP_NAME",
		"DINGTALK_H5_LOGO_TEXT",
		"DINGTALK_H5_LOGO_URL",
		`"appName":`,
		`"logoText":`,
		`"logoUrl":`,
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
}

func TestDingTalkSettingsSeparatesCorpConfigFromLoginConfig(t *testing.T) {
	viewSrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-setup/index.vue")
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
	viewSrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-setup/index.vue")
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

func TestDingTalkSettingsExposePerformanceNotificationConfig(t *testing.T) {
	handlerSrc, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	viewSrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-setup/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk setup page: %v", err)
	}
	combined := string(handlerSrc) + string(viewSrc)
	for _, want := range []string{
		"DINGTALK_H5_NOTIFY_ENABLED",
		"notifyEnabled",
		"agentId",
		"form.notifyEnabled",
		"绩效流程通知",
		"钉钉内部应用 AgentId",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("dingtalk settings should expose performance notification config with %q", want)
		}
	}
}
