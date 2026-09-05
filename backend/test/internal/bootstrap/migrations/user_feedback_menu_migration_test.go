package migrations_test

import (
	"strings"
	"testing"
)

func TestUserFeedbackH5MenuCorrectionUsesNewTargetedMigration(t *testing.T) {
	legacy := readUserFeedbackMigration(t, "20260905101000_add_user_feedback_permissions.sql")
	if !strings.Contains(legacy, "('dingtalk_h5:menu:feedback', '用户反馈', 'dingtalk_h5', 'menu'") ||
		!strings.Contains(legacy, "'', 'feedback', 'mine', ''") {
		t.Fatal("published feedback permission migration must remain immutable")
	}

	correction := readUserFeedbackMigration(t, "20260905102000_update_user_feedback_h5_menu.sql")
	requireFragments(t, correction, []string{
		"UPDATE `permissions`",
		"`permission_name` = '我的反馈'",
		"`permission_icon` = 'chat'",
		"`permission_key` = 'dingtalk_h5:menu:feedback'",
		"`permission_platform` = 'dingtalk_h5'",
		"`permission_type` = 'menu'",
	})
	if strings.Count(correction, "UPDATE `permissions`") != 1 {
		t.Fatal("feedback menu correction must update permissions exactly once")
	}
	lower := strings.ToLower(correction)
	for _, forbidden := range []string{"insert into", "delete from", "permission_grants", "alter table"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("feedback menu correction contains forbidden operation %q", forbidden)
		}
	}
}
