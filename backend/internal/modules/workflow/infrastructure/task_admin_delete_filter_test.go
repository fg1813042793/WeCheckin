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

func TestApplyTaskFiltersHidesAdminDeletedTasksOnlyForAdminList(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	var rows []workflowmodel.ProcessTask
	statement := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), application.TaskQuery{HideAdminDeleted: true}).Find(&rows).Statement
	sqlText := strings.Join(strings.Fields(statement.SQL.String()), " ")
	if !strings.Contains(sqlText, "workflow_process_tasks.admin_deleted_at = ?") {
		t.Fatalf("Admin task query must hide soft-deleted tasks: %s", sqlText)
	}
	if !reflect.DeepEqual(statement.Vars, []interface{}{int64(0)}) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, []interface{}{int64(0)})
	}

	statement = applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), application.TaskQuery{}).Find(&rows).Statement
	if strings.Contains(statement.SQL.String(), "workflow_process_tasks.admin_deleted_at") {
		t.Fatalf("runtime task query must preserve soft-deleted tasks for audit: %s", statement.SQL.String())
	}
}
