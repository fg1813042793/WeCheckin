package access

import (
	"os"
	"strings"
	"testing"
)

func TestDataScopeDoesNotFallbackToLegacyRoleFields(t *testing.T) {
	src, err := os.ReadFile("access.go")
	if err != nil {
		t.Fatalf("read access.go: %v", err)
	}
	text := string(src)

	required := []string{
		"permissionsupport.DataScopeContext",
		"UserDataScopeFilterWithDBContext",
		"VisibleDeptIDsWithDBContext",
		"adminDeptIDsWithDBContext(ctx, db, admin.ID)",
		"dataScopeFilterByMode",
		"visibleDeptIDsByScope",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("data scope must use unified permission path with %q", snippet)
		}
	}

	forbidden := []string{
		"role.DataScope",
		"db.First(&role, admin.RoleID)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("data scope must not fallback to legacy role fields, found %q", snippet)
		}
	}
}
