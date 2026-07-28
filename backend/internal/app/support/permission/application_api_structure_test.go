package permission

import (
	"os"
	"strings"
	"testing"
)

func TestPermissionServiceSyncsApplicationAPIPermissions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`"wecheckin-backend/backend/internal/app/support/appapiperm"`,
		"syncClientAPIPermissions(db)",
		"syncDingTalkH5APIPermissions(db)",
		"appapiperm.ClientAPICategories",
		"appapiperm.ClientAPIDeclarations",
		"appapiperm.DingTalkH5APICategories",
		"appapiperm.DingTalkH5APIDeclarations",
		"ApplicationAPIPermissionPrefixes",
		"SubjectAPIPermissionReadyContext",
		"RoleApplicationAPIKeyMapContext",
		"SetRoleApplicationAPIPermissionsTx",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("permission service must sync application API permissions with %s", snippet)
		}
	}
}
