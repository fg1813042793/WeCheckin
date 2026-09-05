package infrastructure

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	workflowmodel "wecheckin/backend/internal/model/workflow"
)

func TestWorkflowOverviewQueryUsesOnlyCountSubqueries(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      &sql.DB{},
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	statement := workflowOverviewQuery(db, " 7 ").Statement
	sqlText := strings.Join(strings.Fields(statement.SQL.String()), " ")
	for _, fragment := range []string{
		"SELECT COUNT(*) FROM workflow_process_tasks overview_task",
		"SELECT COUNT(*) FROM workflow_process_instances handled_instance",
		"SELECT COUNT(*) FROM workflow_process_instances started_instance",
		"SELECT COUNT(*) FROM workflow_process_instances copied_instance",
		"overview_task.task_assignee_id = ? AND overview_task.task_status = ?",
		"handled_task.task_status IN (?, ?, ?, ?)",
		"handled_task.task_assignee_id = ? OR handled_task.handled_by = ?",
		"started_instance.starter_id = ? AND started_instance.starter_deleted_at = 0",
		"copied_participant.user_id = ? AND copied_participant.participant_role = ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("overview query missing %q: %s", fragment, sqlText)
		}
	}
	for _, forbidden := range []string{"SELECT *", " ORDER BY ", " LIMIT "} {
		if strings.Contains(sqlText, forbidden) {
			t.Fatalf("overview query must not read list records (%q): %s", forbidden, sqlText)
		}
	}

	wantVars := []interface{}{
		"7", workflowmodel.TaskStatusPending,
		workflowmodel.TaskStatusCompleted,
		workflowmodel.TaskStatusApproved,
		workflowmodel.TaskStatusRejected,
		workflowmodel.TaskStatusReturned,
		"7", "7",
		"7",
		"7", workflowmodel.ParticipantRoleCC,
	}
	if !reflect.DeepEqual(statement.Vars, wantVars) {
		t.Fatalf("overview query vars = %#v, want %#v", statement.Vars, wantVars)
	}
}
