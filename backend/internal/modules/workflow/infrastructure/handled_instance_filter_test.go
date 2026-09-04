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

func TestApplyInstanceFiltersHandledScopeRequiresTerminalTaskStatus(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	query := application.InstanceQuery{Scope: application.InstanceScopeHandled, ScopeUserID: " 7 "}
	var rows []workflowmodel.ProcessInstance
	statement := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).Find(&rows).Statement
	sqlText := strings.Join(strings.Fields(statement.SQL.String()), " ")
	for _, fragment := range []string{
		"scope_task.task_status IN (?, ?, ?, ?)",
		"scope_task.task_assignee_id = ? OR scope_task.handled_by = ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("handled instance query missing %q: %s", fragment, sqlText)
		}
	}
	want := []interface{}{
		workflowmodel.TaskStatusCompleted,
		workflowmodel.TaskStatusApproved,
		workflowmodel.TaskStatusRejected,
		workflowmodel.TaskStatusReturned,
		"7",
		"7",
	}
	if !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}
