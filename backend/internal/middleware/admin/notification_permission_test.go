package admin

import (
	"os"
	"strings"
	"testing"
)

func TestNotificationRecipientOptionsRequiresSendPermission(t *testing.T) {
	source, err := os.ReadFile("route_permissions.go")
	if err != nil {
		t.Fatal(err)
	}
	compact := strings.Join(strings.Fields(string(source)), " ")
	if !strings.Contains(compact, `"GET /api/v2/admin/in-app-notifications/recipient-options": "notification:send"`) {
		t.Fatal("recipient options route must require notification:send")
	}
}

func TestDingTalkNotificationRoutesRequireDedicatedSendPermission(t *testing.T) {
	source, err := os.ReadFile("route_permissions.go")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(strings.Fields(string(source)), " ")
	for _, snippet := range []string{
		`"GET /api/v2/admin/dingtalk-notifications/recipient-options": "notification:dingtalk:send"`,
		`"POST /api/v2/admin/dingtalk-notifications": "notification:dingtalk:send"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("DingTalk notification permission mapping missing %q", snippet)
		}
	}
}

func TestNotificationStyleRoutesRequireDedicatedPermissions(t *testing.T) {
	source, err := os.ReadFile("route_permissions.go")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(strings.Fields(string(source)), " ")
	for _, snippet := range []string{
		`"GET /api/v2/admin/notification-styles": "notification:style:list"`,
		`"PUT /api/v2/admin/notification-styles": "notification:style:edit"`,
		`"POST /api/v2/admin/notification-styles/test/in-app": "notification:send"`,
		`"POST /api/v2/admin/notification-styles/test/dingtalk": "notification:dingtalk:send"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("notification style permission mapping missing %q", snippet)
		}
	}
}
