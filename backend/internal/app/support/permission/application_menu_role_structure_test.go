package permission

import (
	"os"
	"strings"
	"testing"
)

func TestPermissionServiceSupportsRoleApplicationMenuGrants(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"SetRoleApplicationMenuPermissionsTx",
		"syncClientMenuPermissions(tx)",
		"syncDingTalkH5MenuPermissions(tx)",
		"replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationMenuPermissionPrefixes(), keys, EffectAllow, nil, \"form\")",
		"RoleApplicationMenuKeyMapContext",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must support role app menu grants with %s", snippet)
		}
	}
}
