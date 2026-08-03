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
		`"wecheckin/backend/internal/app/support/appapiperm"`,
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
	body := testFunctionBody(t, text, "SetRoleApplicationAPIPermissionsTx")
	for _, forbidden := range []string{
		"syncClientAPIPermissions(tx)",
		"syncDingTalkH5APIPermissions(tx)",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("role app API grant save must not run slow catalog sync snippet %s", forbidden)
		}
	}
}
