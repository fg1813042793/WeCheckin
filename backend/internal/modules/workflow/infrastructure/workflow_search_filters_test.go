package infrastructure

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

func openWorkflowSearchDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &sql.DB{}, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}
	return db
}

func TestApplyInstanceFiltersIncludesLiteralStarterName(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	query := application.InstanceQuery{StarterName: " 研发%_!张 "}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{"users", "user_name LIKE ?", "ESCAPE '!'"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("starter name query missing %q: %s", fragment, sqlText)
		}
	}
	if want := []interface{}{"%研发!%!_!!张%"}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}

func TestApplyInstanceFiltersIncludesLiteralDefinitionName(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	query := application.InstanceQuery{DefinitionName: " 绩效%_! "}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{"workflow_definitions", "definition_name LIKE ?", "ESCAPE '!'"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("definition name query missing %q: %s", fragment, sqlText)
		}
	}
	if want := []interface{}{"%绩效!%!_!!%"}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}

func TestApplyTaskFiltersIncludesInstanceSearchFilters(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	query := application.TaskQuery{
		Status: "pending", DefinitionName: "绩效", DefinitionCategory: "performance", StarterName: "张",
		StartTimeFrom: 1000, StartTimeTo: 1999,
	}
	var rows []workflowmodel.ProcessTask
	statement := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), query).Find(&rows).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{
		"task_status = ?", "instance_id IN", "workflow_process_instances",
		"workflow_definitions", "definition_name LIKE ?", "definition_category = ?", "users",
		"user_name LIKE ?", "start_time >= ?", "start_time <= ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("task query missing %q: %s", fragment, sqlText)
		}
	}
	if want := []interface{}{"pending", "%绩效%", "performance", "%张%", int64(1000), int64(1999)}; !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}
