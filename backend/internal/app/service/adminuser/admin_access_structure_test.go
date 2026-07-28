package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestUserManagementExposesAdminAccessFields(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	required := []string{
		"RoleID",
		"RoleName",
		"user_role_id",
		"loadRoleNameMapContext(ctx, db, list)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user management must expose backend access field or helper %q", snippet)
		}
	}

	forbidden := []string{
		"Account      string",
		"AdminType",
		"AdminEnabled",
		"AdminDeptIDs",
		"user_account",
		"user_admin_type",
		"user_admin_enabled",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("user management must not expose separate backend account field %q", snippet)
		}
	}
}

func TestUserManagementSavesAdminAccessInSameTransaction(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	required := []string{
		"type AdminAccessInput struct",
		"Password",
		"RoleID",
		"AllowPermissionKeys []string",
		"DenyPermissionKeys  []string",
		"saveUserAdminAccessTx(tx, user.ID, adminAccess)",
		"saveUserAdminAccessTx(tx, uint(uid), adminAccess)",
		"permissionsupport.SetUserApplicationMenuPermissionOverridesTx",
		"saveUserDeptsTx(tx, user.ID, deptIDs)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user add/edit must persist backend access with %q", snippet)
		}
	}

	if strings.Contains(text, "access.SaveAdminDeptsTx(tx, userID") {
		t.Fatalf("user admin access must not write a separate admin_depts relation")
	}
	for _, snippet := range []string{"AdminType", "AdminEnabled", "Account", "AdminDeptIDs", "user_admin_type", "user_admin_enabled", "user_account"} {
		if strings.Contains(text, snippet) {
			t.Fatalf("user admin access must be controlled by role, found old field %q", snippet)
		}
	}
}
