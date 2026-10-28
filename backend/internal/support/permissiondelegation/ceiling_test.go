package permissiondelegation

import (
	"errors"
	"reflect"
	"testing"

	permissionsupport "wecheckin/backend/internal/support/permission"
)

func TestCeilingRejectsPermissionsOutsideEffectiveSet(t *testing.T) {
	ceiling := Ceiling{PermissionKeys: stringSet("admin:menu:workflow", "admin:api:workflow:list")}
	if err := ceiling.ValidatePermissionKeys([]string{"admin:menu:workflow"}); err != nil {
		t.Fatalf("owned permission rejected: %v", err)
	}
	if err := ceiling.ValidatePermissionKeys([]string{"admin:api:setup:edit"}); !errors.Is(err, ErrPermissionExceeded) {
		t.Fatalf("unowned permission error = %v", err)
	}
}

func TestCeilingReportsAllowedRoleDataScopes(t *testing.T) {
	tests := []struct {
		name    string
		ceiling Ceiling
		want    []int
	}{
		{name: "super", ceiling: Ceiling{All: true}, want: []int{1, 2, 3, 4}},
		{name: "all data", ceiling: Ceiling{DataScope: permissionsupport.DataScope{Ready: true, Mode: 1}}, want: []int{1, 2, 3, 4}},
		{name: "department", ceiling: Ceiling{DataScope: permissionsupport.DataScope{Ready: true, Mode: 2}, DepartmentIDs: []uint{10}}, want: []int{2, 3, 4}},
		{name: "custom", ceiling: Ceiling{DataScope: permissionsupport.DataScope{Ready: true, Mode: 4}, DepartmentIDs: []uint{10}}, want: []int{3, 4}},
		{name: "self", ceiling: Ceiling{DataScope: permissionsupport.DataScope{Ready: true, Mode: 3}}, want: []int{3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.ceiling.AllowedDataScopes(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("AllowedDataScopes() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestCeilingRestrictsRoleDataScope(t *testing.T) {
	departmentCeiling := Ceiling{
		PermissionKeys: stringSet(permissionsupport.AdminLoginPermissionKey),
		DataScope:      permissionsupport.DataScope{Ready: true, Mode: 2},
		DepartmentIDs:  []uint{10, 11, 12},
	}
	if err := departmentCeiling.ValidateRoleGrant(RoleGrant{AllowAdminLogin: 1, DataScope: 2}); err != nil {
		t.Fatalf("department scope rejected: %v", err)
	}
	if err := departmentCeiling.ValidateRoleGrant(RoleGrant{DataScope: 4, DeptIDs: []uint{11}}); err != nil {
		t.Fatalf("custom subset rejected: %v", err)
	}
	if err := departmentCeiling.ValidateRoleGrant(RoleGrant{DataScope: 1}); !errors.Is(err, ErrDataScopeExceeded) {
		t.Fatalf("all-data grant error = %v", err)
	}
	if err := departmentCeiling.ValidateRoleGrant(RoleGrant{DataScope: 4, DeptIDs: []uint{99}}); !errors.Is(err, ErrDepartmentExceeded) {
		t.Fatalf("outside department grant error = %v", err)
	}

	customCeiling := Ceiling{DataScope: permissionsupport.DataScope{Ready: true, Mode: 4}, DepartmentIDs: []uint{10}}
	if err := customCeiling.ValidateRoleGrant(RoleGrant{DataScope: 2}); !errors.Is(err, ErrDataScopeExceeded) {
		t.Fatalf("custom administrator delegated department-relative scope: %v", err)
	}
	if err := customCeiling.ValidateRoleGrant(RoleGrant{DataScope: 3}); err != nil {
		t.Fatalf("self scope rejected: %v", err)
	}
}

func TestCeilingProtectsReservedSuperAdminRole(t *testing.T) {
	child := Ceiling{PermissionKeys: stringSet(permissionsupport.AdminLoginPermissionKey), DataScope: permissionsupport.DataScope{Ready: true, Mode: 3}}
	if err := child.ValidateRoleGrant(RoleGrant{Name: "超级管理员", AllowAdminLogin: 1, DataScope: 3}); !errors.Is(err, ErrReservedRole) {
		t.Fatalf("reserved role error = %v", err)
	}

	super := Ceiling{All: true}
	if err := super.ValidateRoleGrant(RoleGrant{Name: "超级管理员", AllowAdminLogin: 1, DataScope: 1, PermissionKeys: []string{"admin:api:anything"}}); err != nil {
		t.Fatalf("super administrator grant rejected: %v", err)
	}
}
