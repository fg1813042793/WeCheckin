package admincontent

import (
	"os"
	"strings"
	"testing"
)

func TestEnrollExportLoadsOnlyReferencedUsers(t *testing.T) {
	src, err := os.ReadFile("export.go")
	if err != nil {
		t.Fatalf("read export.go: %v", err)
	}
	text := string(src)
	if strings.Contains(text, "db.Find(&users)") {
		t.Fatalf("export must not load the full users table")
	}
	required := []string{
		"uniqueNonEmptyStrings(userIDs)",
		"Where(\"`user_mini_openid` IN ?\"",
		"Select(\"user_mini_openid\", \"user_name\")",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("export must batch load referenced users with %q", snippet)
		}
	}
}
