package workflowsummary

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

func TestBuildInstanceVisibilityUsesBaseAndExtraScopes(t *testing.T) {
	departments := []*model.Department{
		{ID: 10},
		{ID: 11, ParentID: 10},
		{ID: 12, ParentID: 11},
		{ID: 20},
	}
	tests := []struct {
		name      string
		scope     permissionsupport.DataScope
		extras    permissionsupport.DataScopeExtras
		userDepts []uint
		wantAll   bool
		wantUsers []uint
		wantDepts []uint
		wantReady bool
	}{
		{name: "all", scope: permissionsupport.DataScope{Mode: 1, Ready: true}, wantAll: true, wantReady: true},
		{name: "department descendants", scope: permissionsupport.DataScope{Mode: 2, Ready: true}, userDepts: []uint{10}, wantDepts: []uint{10, 11, 12}, wantReady: true},
		{name: "self", scope: permissionsupport.DataScope{Mode: 3, Ready: true}, wantUsers: []uint{7}, wantReady: true},
		{name: "custom and extras", scope: permissionsupport.DataScope{Mode: 4, DeptIDs: []uint{11}, Ready: true}, extras: permissionsupport.DataScopeExtras{UserIDs: []uint{9}, DeptIDs: []uint{20}, Ready: true}, wantUsers: []uint{9}, wantDepts: []uint{11, 12, 20}, wantReady: true},
		{name: "extras without base", extras: permissionsupport.DataScopeExtras{UserIDs: []uint{8}, Ready: true}, wantUsers: []uint{8}, wantReady: true},
		{name: "missing scope", wantReady: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := buildInstanceVisibility(7, test.scope, test.extras, test.userDepts, departments)
			if got.All != test.wantAll || got.Ready != test.wantReady {
				t.Fatalf("visibility = %#v", got)
			}
			if !reflect.DeepEqual(got.UserIDs, test.wantUsers) {
				t.Fatalf("user IDs = %#v, want %#v", got.UserIDs, test.wantUsers)
			}
			if !reflect.DeepEqual(got.DepartmentIDs, test.wantDepts) {
				t.Fatalf("department IDs = %#v, want %#v", got.DepartmentIDs, test.wantDepts)
			}
		})
	}
}
