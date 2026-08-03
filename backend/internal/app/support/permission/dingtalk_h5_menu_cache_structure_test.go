package permission

import (
	"os"
	"strings"
	"testing"
)

func TestDingTalkH5MenuPermissionUsesShortTTLCache(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"dingtalkH5MenuPermissionCacheTTL",
		"dingtalkH5MenuPermissionCacheKey(userID, roleIDs)",
		"getDingTalkH5MenuPermissionCache(userID, roleIDs)",
		"setDingTalkH5MenuPermissionCache(userID, roleIDs, keys, ready)",
		"invalidateDingTalkH5MenuPermissionCache()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 menu permission cache should include %q", snippet)
		}
	}
	for _, snippet := range []string{
		"SetRoleApplicationMenuPermissionsTx",
		"SetUserApplicationMenuPermissionOverridesTx",
	} {
		start := strings.Index(text, "func "+snippet)
		if start < 0 {
			t.Fatalf("%s missing", snippet)
		}
		body := text[start:]
		if end := strings.Index(body, "\n}\n\nfunc "); end >= 0 {
			body = body[:end+3]
		}
		if !strings.Contains(body, "invalidateDingTalkH5MenuPermissionCache()") {
			t.Fatalf("%s should invalidate dingtalk h5 menu permission cache", snippet)
		}
	}
}
