package infrastructure

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestApplyInstanceFiltersIncludesTimeRanges(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	query := application.InstanceQuery{
		StartTimeFrom: 1000,
		StartTimeTo:   1999,
		EndTimeFrom:   2000,
		EndTimeTo:     2999,
	}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()
	for _, condition := range []string{
		"start_time >= ?",
		"start_time <= ?",
		"end_time >= ?",
		"end_time <= ?",
	} {
		if !strings.Contains(sqlText, condition) {
			t.Fatalf("query missing %q: %s", condition, sqlText)
		}
	}
	if want := []interface{}{int64(1000), int64(1999), int64(2000), int64(2999)}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}

func TestApplyInstanceFiltersIncludesDefinitionCategory(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	query := application.InstanceQuery{DefinitionCategory: " performance "}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()
	for _, fragment := range []string{"EXISTS", "workflow_definitions", "definition_category = ?"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("category query missing %q: %s", fragment, sqlText)
		}
	}
	if want := []interface{}{"performance"}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}

func TestApplyInstanceFiltersHidesStarterDeletedApplications(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	query := application.InstanceQuery{Scope: application.InstanceScopeStarted, ScopeUserID: "7"}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()
	for _, fragment := range []string{"starter_id = ?", "starter_deleted_at = ?"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("starter application query missing %q: %s", fragment, sqlText)
		}
	}
	if want := []interface{}{"7", int64(0)}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}
