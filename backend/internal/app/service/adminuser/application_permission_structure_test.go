package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestUserManagementExposesUserPermissionOverrides(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"AllowPermissionKeys []string `json:\"allowPermissionKeys\"`",
		"DenyPermissionKeys  []string `json:\"denyPermissionKeys\"`",
		"permissionsupport.UserApplicationMenuPermissionKeySetsContext(ctx, db, user.ID)",
		"PermissionKeysTouched bool",
		"adminAccess.PermissionKeysTouched",
		"permissionsupport.SetUserApplicationMenuPermissionOverridesTx",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user management must expose and save direct permission overrides with %s", snippet)
		}
	}
}
