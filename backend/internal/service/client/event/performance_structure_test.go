package event

import (
	"os"
	"strings"
	"testing"
)

func TestEventListsBatchLoadRelatedUserInfo(t *testing.T) {
	checks := []struct {
		file     string
		forbid   []string
		required []string
	}{
		{
			file: "admin.go",
			forbid: []string{
				"db.Where(\"`user_mini_openid` = ?\", list[i].MiniOpenID)",
				"db.Where(\"`user_dept_user_id` = ?\", user.ID).First(&ud)",
				"getTopDeptNameContext(ctx, ud.DeptID)",
			},
			required: []string{
				"enrichEventParticipantsWithUserInfoContext(ctx, db, list)",
			},
		},
		{
			file: "dynamic.go",
			forbid: []string{
				"db.Where(\"`user_mini_openid` = ?\", list[i].UserID)",
			},
			required: []string{
				"loadEventUserInfoByOpenIDContext(ctx, db, userIDs)",
			},
		},
		{
			file: "score.go",
			forbid: []string{
				"db.Where(\"`user_mini_openid` = ?\", list[i].ParticipantID)",
				"db.Where(\"`user_dept_user_id` = ?\", user.ID).First(&ud)",
				"getTopDeptName(ud.DeptID)",
			},
			required: []string{
				"enrichEventScoresWithUserInfoContext(ctx, db, list)",
			},
		},
		{
			file: "my.go",
			forbid: []string{
				"db.Where(\"`event_role_event_id` = ? AND `event_role_user_id` = ?\", list[i].ID, userID)",
			},
			required: []string{
				"loadEventRolesForListWithDB(ctx, db, list, userID)",
				"event_role_event_id` IN ?",
			},
		},
	}

	for _, check := range checks {
		src, err := os.ReadFile(check.file)
		if err != nil {
			t.Fatalf("read %s: %v", check.file, err)
		}
		text := string(src)
		for _, snippet := range check.forbid {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must avoid per-row query snippet %q", check.file, snippet)
			}
		}
		for _, snippet := range check.required {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must batch related data with %q", check.file, snippet)
			}
		}
	}
}
