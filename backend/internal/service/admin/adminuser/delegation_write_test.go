package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestAdminUserWritesValidateDelegationCeiling(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func AddUserWithManagerAndAdminAccessForAdminContext(",
		"permissiondelegation.ResolveByAdminIDContext(ctx, tx, adminID)",
		"ceiling.ValidateRoleIDsContext(ctx, tx, roleIDs)",
		"ceiling.ValidatePermissionKeys",
		"ceiling.ValidateDepartmentIDsContext",
		"ceiling.ValidateUserIDsContext",
		"ceiling.ValidateUserAdminAccessContext(ctx, tx, uint(uid))",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user delegation enforcement missing %q", snippet)
		}
	}
}

