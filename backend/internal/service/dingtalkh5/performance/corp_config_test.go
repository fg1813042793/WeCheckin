package performance

import (
	"os"
	"strings"
	"testing"
)

func TestSaveCorpConfigsReplacesPersistedList(t *testing.T) {
	src, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"func deleteOmittedDingTalkH5CorpConfigsDB",
		"`corp_id` NOT IN ?",
		"deleteOmittedDingTalkH5CorpConfigsDB(tx, savedCorpIDs)",
		"db.Transaction(func(tx *gorm.DB) error",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saving corp configs should replace omitted configs with %q", want)
		}
	}
}

func TestSaveCorpConfigsAllowsClearingAllConfigs(t *testing.T) {
	src, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		`return db.Where("1 = 1").Delete(&model.DingTalkH5CorpConfig{}).Error`,
		"return clearLegacyDingTalkH5CorpConfigContext(ctx)",
		"func clearLegacyDingTalkH5CorpConfigContext(ctx context.Context) error",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saving an empty corp config list should clear all configs with %q", want)
		}
	}
}

func TestSaveCorpConfigsMirrorsLegacySetupInBatch(t *testing.T) {
	src, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"setupItems := []setupservice.SetupItem{",
		"setupservice.SetSetupsContext(ctx, setupItems, \"\")",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("corp config legacy mirror should batch setup writes with %q", want)
		}
	}
	for _, forbidden := range []string{
		"setupservice.SetSetupContext(ctx, item.key",
		"setupservice.SetSetupContext(ctx, \"DINGTALK_H5_APP_SECRET\"",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("corp config legacy mirror should not write setup keys one by one with %q", forbidden)
		}
	}
}

func TestCorpConfigsCarryIndependentNotificationSwitch(t *testing.T) {
	serviceSrc, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	migrationSrc, err := os.ReadFile("../../../../migrations/20260803110000_add_dingtalk_h5_corp_notify_enabled.sql")
	if err != nil {
		t.Fatalf("read corp notify migration: %v", err)
	}
	combined := string(serviceSrc) + string(migrationSrc)
	for _, want := range []string{
		"NotifyEnabled int",
		"`notify_enabled`",
		"notify_enabled",
		"DINGTALK_H5_NOTIFY_ENABLED",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("corp config should carry independent notification switch with %q", want)
		}
	}
}

func TestCorpConfigsCarryEnterpriseAppURL(t *testing.T) {
	serviceSrc, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	modelSrc, err := os.ReadFile("../../../model/dingtalk_h5_performance.go")
	if err != nil {
		t.Fatalf("read dingtalk h5 model: %v", err)
	}
	migrationSrc, err := os.ReadFile("../../../../migrations/20260803143000_add_dingtalk_h5_corp_app_url.sql")
	if err != nil {
		t.Fatalf("read corp app url migration: %v", err)
	}
	combined := string(serviceSrc) + string(modelSrc) + string(migrationSrc)
	for _, want := range []string{
		"AppURL",
		"`app_url`",
		"app_url",
		"DINGTALK_H5_APP_URL",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("corp config should carry enterprise app URL with %q", want)
		}
	}
}

func TestNormalizeDingTalkH5NotifyModeKeepsAgentFallback(t *testing.T) {
	tests := []struct {
		name      string
		mode      string
		agentID   string
		robotCode string
		want      string
	}{
		{name: "strict agent", mode: "agent", agentID: "342080997", robotCode: "robot-code", want: "agent"},
		{name: "robot", mode: "robot", agentID: "342080997", robotCode: "robot-code", want: "robot"},
		{name: "agent fallback", mode: "agent_fallback", agentID: "342080997", robotCode: "robot-code", want: "agent_fallback"},
		{name: "hyphen alias", mode: "agent-fallback", agentID: "342080997", robotCode: "robot-code", want: "agent_fallback"},
		{name: "legacy robot code", mode: "", agentID: "", robotCode: "robot-code", want: "robot"},
		{name: "legacy agent id", mode: "", agentID: "342080997", robotCode: "", want: "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDingTalkH5NotifyMode(tt.mode, tt.agentID, tt.robotCode); got != tt.want {
				t.Fatalf("normalizeDingTalkH5NotifyMode(%q, %q, %q) = %q, want %q", tt.mode, tt.agentID, tt.robotCode, got, tt.want)
			}
		})
	}
}
