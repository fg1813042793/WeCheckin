package migrations

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"wecheckin/backend/internal/model/notification"
)

func TestDingTalkTemplateMigrationMatchesModelAndPreservesExistingData(t *testing.T) {
	data, err := os.ReadFile("20261029120000_add_dingtalk_notification_templates.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, want := range []string{"CREATE TABLE IF NOT EXISTS `dingtalk_notification_templates`", "UNIQUE KEY `uk_dingtalk_notification_template_name` (`name`)", "VARCHAR(80)", "VARCHAR(255)", "VARCHAR(2048)", "VARCHAR(20)", "VARCHAR(8)", "DEFAULT 'FF1677FF'", "DEFAULT '打开页面'", "DEFAULT CHARSET=utf8mb4"} {
		if !strings.Contains(sql, want) {
			t.Errorf("migration missing %s", want)
		}
	}
	typ := reflect.TypeOf(notificationmodel.DingTalkTemplate{})
	for i := 0; i < typ.NumField(); i++ {
		for _, tag := range strings.Split(typ.Field(i).Tag.Get("gorm"), ";") {
			if strings.HasPrefix(tag, "column:") && !strings.Contains(sql, "`"+strings.TrimPrefix(tag, "column:")+"`") {
				t.Errorf("missing column for %s", typ.Field(i).Name)
			}
		}
	}
	for _, bad := range []string{"DROP TABLE", "DELETE FROM", "TRUNCATE", "INSERT INTO"} {
		if strings.Contains(strings.ToUpper(sql), bad) {
			t.Errorf("unexpected data mutation %s", bad)
		}
	}
}
