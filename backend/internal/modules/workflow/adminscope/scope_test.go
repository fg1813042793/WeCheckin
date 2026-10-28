package adminscope

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

func TestBuildVisibilityExpandsCustomDepartmentsAndExtras(t *testing.T) {
	departments := []*model.Department{
		{ID: 10},
		{ID: 11, ParentID: 10},
		{ID: 12, ParentID: 11},
		{ID: 20},
		{ID: 21, ParentID: 20},
	}

	visibility := buildVisibility(
		42,
		permissionsupport.DataScope{Ready: true, Mode: 4, DeptIDs: []uint{10}},
		permissionsupport.DataScopeExtras{Ready: true, DeptIDs: []uint{20}, UserIDs: []uint{99, 99}},
		nil,
		departments,
	)

	if !visibility.Ready || visibility.All {
		t.Fatalf("visibility flags = %#v", visibility)
	}
	if !reflect.DeepEqual(visibility.DepartmentIDs, []uint{10, 11, 12, 20, 21}) {
		t.Fatalf("department ids = %#v", visibility.DepartmentIDs)
	}
	if !reflect.DeepEqual(visibility.UserIDs, []uint{99}) {
		t.Fatalf("user ids = %#v", visibility.UserIDs)
	}
}

func TestScopeAppliesVisibleDefinitionIDs(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &sql.DB{}, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}

	var rows []model.WorkflowDefinition
	statement := (Scope{DefinitionIDs: []uint{7, 9}}).
		ApplyDefinitionQuery(db.Model(&model.WorkflowDefinition{})).
		Find(&rows).Statement
	if !strings.Contains(statement.SQL.String(), "workflow_definitions.id IN") {
		t.Fatalf("definition scope missing: %s", statement.SQL.String())
	}
	if !reflect.DeepEqual(statement.Vars, []interface{}{uint(7), uint(9)}) {
		t.Fatalf("definition vars = %#v", statement.Vars)
	}

	statement = (Scope{}).ApplyDefinitionQuery(db.Model(&model.WorkflowDefinition{})).Find(&rows).Statement
	if !strings.Contains(statement.SQL.String(), "1 = 0") {
		t.Fatalf("empty definition scope must deny all: %s", statement.SQL.String())
	}
}

func TestBuildVisibilitySupportsSelfAndAllScopes(t *testing.T) {
	self := buildVisibility(42, permissionsupport.DataScope{Ready: true, Mode: 3}, permissionsupport.DataScopeExtras{}, nil, nil)
	if !reflect.DeepEqual(self.UserIDs, []uint{42}) || self.All {
		t.Fatalf("self visibility = %#v", self)
	}

	all := buildVisibility(42, permissionsupport.DataScope{Ready: true, Mode: 1}, permissionsupport.DataScopeExtras{}, nil, nil)
	if !all.Ready || !all.All || len(all.UserIDs) != 0 || len(all.DepartmentIDs) != 0 {
		t.Fatalf("all visibility = %#v", all)
	}
}
