package config

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5NotificationDiagnosisExposesSanitizedCallChain(t *testing.T) {
	src, err := os.ReadFile("notification_diagnosis.go")
	if err != nil {
		t.Fatalf("read notification_diagnosis.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"type DingTalkH5NotificationDiagnosis struct",
		"type DingTalkH5NotificationDiagnosisStep struct",
		"func DiagnoseDingTalkH5WorkNotificationContext",
		"loadDingTalkH5CorpConfigContext(ctx, corpID)",
		"diagnosisRecipientForCorpContext(ctx, config.CorpID, recipientUserID)",
		`"/gettoken?"`,
		`"/topapi/message/corpconversation/asyncsend_v2?"`,
		`"accessTokenReceived"`,
		`"appSecretSet"`,
		"dingTalkH5MaskLogValue",
		"agentId【%s】不合法",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("notification diagnosis should expose sanitized call chain with %q", want)
		}
	}
	for _, forbidden := range []string{
		`"access_token":`,
		`"appSecret": config.AppSecret`,
		`"appsecret": config.AppSecret`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("notification diagnosis must not expose secret/token with %q", forbidden)
		}
	}
}
