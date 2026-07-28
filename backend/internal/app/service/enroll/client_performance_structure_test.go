package enroll

import (
	"os"
	"strings"
	"testing"
)

func TestEnrollClientListLimitsJoinLookupToCurrentPage(t *testing.T) {
	src, err := os.ReadFile("client.go")
	if err != nil {
		t.Fatalf("read client.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"db.Where(\"`enroll_join_user_id` = ?\", userID).Find(&joins)",
		"db.Where(\"`enroll_user_mini_openid` = ?\", userID).Find(&enrollUsers)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("enroll client list should not load all user history: %q", snippet)
		}
	}
	required := []string{
		"enrollListIDs(list)",
		"`enroll_join_enroll_id` IN ?",
		"`enroll_user_enroll_id` IN ?",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("enroll client list should limit join lookup to current page with %q", snippet)
		}
	}
}
