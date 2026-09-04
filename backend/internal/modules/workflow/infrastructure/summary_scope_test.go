package infrastructure

import (
	"strings"
	"testing"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

func TestInstanceVisibilityWhereDefaultsToDeny(t *testing.T) {
	where, _ := instanceVisibilityWhere(&workflowapp.InstanceVisibility{})
	if where != "1 = 0" {
		t.Fatalf("where = %q, want deny-all", where)
	}
}

func TestInstanceVisibilityWhereSupportsUsersAndDepartments(t *testing.T) {
	where, args := instanceVisibilityWhere(&workflowapp.InstanceVisibility{
		Ready:         true,
		UserIDs:       []uint{7, 9},
		DepartmentIDs: []uint{11, 12},
	})
	for _, snippet := range []string{"CAST(starter_id AS UNSIGNED) IN ?", "user_depts", "user_dept_dept_id IN ?"} {
		if !strings.Contains(where, snippet) {
			t.Fatalf("where %q missing %q", where, snippet)
		}
	}
	if len(args) != 2 {
		t.Fatalf("args = %#v, want two query arguments", args)
	}
}

func TestInstanceVisibilityWhereAllowsAll(t *testing.T) {
	where, args := instanceVisibilityWhere(&workflowapp.InstanceVisibility{Ready: true, All: true})
	if where != "" || len(args) != 0 {
		t.Fatalf("where = %q args = %#v, want unrestricted", where, args)
	}
}
