package permission

import (
	"strings"
	"testing"
)

func TestPermissionServiceExposesDirectUserPermissionKeySets(t *testing.T) {
	text := readPermissionPackageSource(t)
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
