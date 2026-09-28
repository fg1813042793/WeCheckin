package infrastructure

import (
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestInitiatorDepartmentScopeStoreStructure(t *testing.T) {
	source := readWorkflowPackageSource(t)
	for _, snippet := range []string{
		"func (store *GormStore) UserDepartmentIDs",
		`Table("user_depts")`,
		`Where("user_dept_user_id = ?", userID)`,
		`Pluck("user_dept_dept_id", &departmentIDs)`,
		`Select("id", "dept_parent_id")`,
		"departmentIDsWithAncestors",
		"normalizeUintIDs(departmentIDs)",
	} {
		if !strings.Contains(source, snippet) {
			t.Fatalf("initiator department lookup must include %q", snippet)
		}
	}
}

func TestDepartmentIDsWithAncestorsIncludesSelectedParentScopes(t *testing.T) {
	departments := []model.Department{
		{ID: 1, ParentID: 0},
		{ID: 2, ParentID: 1},
		{ID: 3, ParentID: 2},
		{ID: 4, ParentID: 3},
	}
	got := departmentIDsWithAncestors([]uint{4}, departments)
	want := []uint{4, 3, 2, 1}
	if len(got) != len(want) {
		t.Fatalf("department scope = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("department scope = %#v, want %#v", got, want)
		}
	}
}

func TestDepartmentIDsWithAncestorsStopsOnCycles(t *testing.T) {
	departments := []model.Department{{ID: 1, ParentID: 2}, {ID: 2, ParentID: 1}}
	got := departmentIDsWithAncestors([]uint{1}, departments)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("cyclic department scope = %#v", got)
	}
}
