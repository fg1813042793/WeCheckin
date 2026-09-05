package infrastructure

import (
	"strings"
	"testing"
)

func TestInitiatorDepartmentScopeStoreStructure(t *testing.T) {
	source := readWorkflowPackageSource(t)
	for _, snippet := range []string{
		"func (store *GormStore) UserDepartmentIDs",
		`Table("user_depts")`,
		`Where("user_dept_user_id = ?", userID)`,
		`Pluck("user_dept_dept_id", &departmentIDs)`,
		"normalizeUintIDs(departmentIDs)",
	} {
		if !strings.Contains(source, snippet) {
			t.Fatalf("initiator department lookup must include %q", snippet)
		}
	}
}
