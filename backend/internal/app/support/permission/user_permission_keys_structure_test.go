package permission

import (
	"os"
	"strings"
	"testing"
)

func TestPermissionServiceExposesDirectUserPermissionKeySets(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"UserApplicationMenuPermissionKeySetsContext",
		"SetUserApplicationMenuPermissionOverridesTx",
		"directSubjectPermissionKeySetsContext",
		"SubjectUser",
		"ApplicationMenuPermissionPrefixes()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must expose direct user permission key sets with %s", snippet)
		}
	}
}
