package role

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleListUsesRoleAuditFieldsForDataScope(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"RoleAuditFields",
		"access.DataScopeFilterForResourceWithDBContext(ctx, db, &admin, access.RoleAuditFields)",
		"permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)",
		"roleVisibleCondition",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role list must use role audit data-scope field with %q", snippet)
		}
	}
	for _, legacySnippet := range []string{
		"permissionsupport.RoleIDsByCustomDeptIDsContext",
		"`id` IN ? OR `id` = ?",
	} {
		if strings.Contains(text, legacySnippet) {
			t.Fatalf("role list must not use legacy role grant reverse lookup %q", legacySnippet)
		}
	}
}

func TestRoleCreationPersistsCreatorContext(t *testing.T) {
	handler, err := os.ReadFile(filepath.Join("..", "..", "..", "handler", "admin", "role", "handler.go"))
	if err != nil {
		t.Fatalf("read role handler: %v", err)
	}
	if !strings.Contains(string(handler), "AddWithAssignmentsContext(ctx, admin.ID, name") {
		t.Fatalf("role handler must pass current admin id when creating roles")
	}

	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func AddWithAssignmentsContext(ctx context.Context, adminID uint, name",
		"roleCreatorContext(ctx, db, adminID)",
		"CreateBy:        creatorID",
		"CreateDeptID:    creatorDeptID",
		"UpdateBy:        creatorID",
		"UpdateDeptID:    creatorDeptID",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role creation must persist creator context with %q", snippet)
		}
	}
}
