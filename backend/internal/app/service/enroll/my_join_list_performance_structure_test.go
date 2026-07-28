package enroll

import (
	"os"
	"strings"
	"testing"
)

func TestMyEnrollJoinListPagesEnrollIDsInSQL(t *testing.T) {
	src, err := os.ReadFile("records.go")
	if err != nil {
		t.Fatalf("read records.go: %v", err)
	}
	text := string(src)
	forbidden := []string{
		"db.Where(\"`enroll_join_user_id` = ?\", userID).Order(\"`enroll_join_add_time` DESC\").Find(&joins)",
		"db.Where(\"`enroll_user_mini_openid` = ?\", userID).Find(&enrollUsers)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("my enroll join list should not load all user history before pagination: %q", snippet)
		}
	}
	required := []string{
		"loadMyEnrollIDsPage(db, userID, page, pageSize)",
		"myEnrollIDsBaseSQL",
		"LIMIT ? OFFSET ?",
		"Total: total",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("my enroll join list should page enroll IDs in SQL with %q", snippet)
		}
	}
}
