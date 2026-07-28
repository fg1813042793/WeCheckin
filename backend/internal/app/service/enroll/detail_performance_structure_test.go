package enroll

import (
	"os"
	"strings"
	"testing"
)

func TestEnrollDetailRankLoadsOnlyReferencedUsers(t *testing.T) {
	src, err := os.ReadFile("detail.go")
	if err != nil {
		t.Fatalf("read detail.go: %v", err)
	}
	text := string(src)
	if strings.Contains(text, "db.Find(&allUsers)") {
		t.Fatalf("enroll detail rank should not load the full users table")
	}
	required := []string{
		"rankUserOpenIDs(enrollUsers)",
		"Where(\"`user_mini_openid` IN ?\"",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("enroll detail rank should batch load referenced users with %q", snippet)
		}
	}
}
