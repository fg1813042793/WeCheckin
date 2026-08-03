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
		"replaceSubjectGrantsTx(tx, SubjectRole, roleID, ApplicationMenuPermissionPrefixes(), keys, EffectAllow, nil, \"form\")",
		"RoleApplicationMenuKeyMapContext",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must support role app menu grants with %s", snippet)
		}
	}
	body := testFunctionBody(t, text, "SetRoleApplicationMenuPermissionsTx")
	for _, forbidden := range []string{
		"syncClientMenuPermissions(tx)",
		"syncDingTalkH5MenuPermissions(tx)",
		"syncDingTalkH5ButtonPermissions(tx)",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("role app menu grant save must not run slow catalog sync snippet %s", forbidden)
		}
	}
}
