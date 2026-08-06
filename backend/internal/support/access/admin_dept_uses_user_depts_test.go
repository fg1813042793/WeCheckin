package access

import (
	"os"
	"strings"
	"testing"
)

func TestAdminDeptAccessUsesUserDeptsTable(t *testing.T) {
	src, err := os.ReadFile("access.go")
	if err != nil {
		t.Fatalf("read access.go: %v", err)
	}
	text := string(src)

	required := []string{
		"model.UserDept",
		"user_dept_user_id",
		"UserID: adminID",
		"DeptID: deptID",
		"permissionsupport.DataScopeContext",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin data-scope dept access must use user_depts with %q", snippet)
		}
	}

	forbidden := []string{
		"model.AdminDept",
		"admin_depts",
		"admin_dept_admin_id",
		"admin_dept_dept_id",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("admin data-scope dept access must not use obsolete admin_depts snippet %q", snippet)
		}
	}
}
