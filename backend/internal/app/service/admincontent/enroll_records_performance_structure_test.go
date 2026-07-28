package admincontent

import (
	"os"
	"strings"
	"testing"
)

func TestEnrollRecordListsBatchLoadUserAndDeptInfo(t *testing.T) {
	src, err := os.ReadFile("enroll_records.go")
	if err != nil {
		t.Fatalf("read enroll_records.go: %v", err)
	}
	text := string(src)

	forbidden := []string{
		"db.Find(&users)",
		"db.Where(\"`user_mini_openid` = ?\", list[i].MiniOpenID)",
		"db.Where(\"`user_dept_user_id` = ?\", u.ID).First(&ud)",
		"dept.TopDeptNameContext(ctx, ud.DeptID)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("enroll record lists must avoid per-row query snippet %q", snippet)
		}
	}

	required := []string{
		"enrichEnrollUsersWithUserInfoContext(ctx, db, list)",
		"enrichEnrollJoinsWithUserInfoContext(ctx, db, list)",
		"loadUserDeptInfoByOpenIDContext(ctx, db, openIDs)",
		"COUNT(DISTINCT `enroll_join_day`)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("enroll record lists must batch load related data with %q", snippet)
		}
	}
}
