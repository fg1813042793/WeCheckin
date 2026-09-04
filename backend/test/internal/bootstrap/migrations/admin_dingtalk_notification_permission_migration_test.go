package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminDingTalkNotificationMigrationRegistersPermissionWithoutAutoGrant(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_admin_dingtalk_notification_permission.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("admin DingTalk notification permission migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read admin DingTalk notification permission migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"admin:menu:notification:dingtalk-send",
		"admin:api:notification:dingtalk-send",
		"notification:dingtalk:send",
		"/api/v2/admin/dingtalk-notifications",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin DingTalk notification permission migration missing %q", snippet)
		}
	}
	if strings.Contains(strings.ToLower(text), "permission_grants") {
		t.Fatal("admin DingTalk notification permission migration must not auto-grant external notification permission")
	}
}
