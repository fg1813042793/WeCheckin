package infrastructure

import (
	"reflect"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

func TestApplyTaskFiltersIncludesAdminInstanceVisibility(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	visibility := &application.InstanceVisibility{Ready: true, UserIDs: []uint{42}, DepartmentIDs: []uint{10, 11}}
	query := application.TaskQuery{IDs: []string{"task-1"}, Visibility: visibility}
	var rows []workflowmodel.ProcessTask
	statement := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), query).Find(&rows).Statement
	normalized := strings.Join(strings.Fields(statement.SQL.String()), " ")

	for _, fragment := range []string{
		"workflow_process_tasks.id IN (",
		"CAST(starter_id AS UNSIGNED) IN (",
		"user_depts",
		"user_dept_dept_id IN (",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("task visibility query missing %q: %s", fragment, normalized)
		}
	}
	want := []interface{}{"task-1", uint(42), uint(10), uint(11)}
	if !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}

func TestUnifiedTaskQueryIncludesAdminInstanceVisibility(t *testing.T) {
	visibility := &application.InstanceVisibility{Ready: true, DepartmentIDs: []uint{10}}
	sqlText, args := unifiedTaskListSQL(application.TaskQuery{
		IDs: []string{"task-1"}, Visibility: visibility, IncludeFormRevisions: true, Page: 1, PageSize: 20,
	})
	normalized := strings.Join(strings.Fields(sqlText), " ")
	if strings.Count(normalized, "user_depts") != 2 {
		t.Fatalf("both task branches must enforce department visibility: %s", normalized)
	}
	if strings.Count(normalized, "task_instance.starter_id") != 2 {
		t.Fatalf("both task branches must filter by instance starter: %s", normalized)
	}
	if len(args) < 4 || args[len(args)-2] != 20 || args[len(args)-1] != 0 {
		t.Fatalf("unexpected unified task args: %#v", args)
	}
}

func TestNotificationQueryIncludesAdminInstanceVisibility(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	visibility := &application.InstanceVisibility{Ready: true, UserIDs: []uint{42}}
	query := application.NotificationQuery{IDs: []string{"notice-1"}, Visibility: visibility}
	var rows []workflowmodel.NotificationOutbox
	statement := applyNotificationQueryFilters(applyActiveNotificationFilter(db.Model(&workflowmodel.NotificationOutbox{})), query).Find(&rows).Statement
	normalized := strings.Join(strings.Fields(statement.SQL.String()), " ")

	for _, fragment := range []string{
		"workflow_notification_outbox.id IN (",
		"workflow_process_instances notification_instance",
		"CAST(notification_instance.starter_id AS UNSIGNED) IN (",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("notification visibility query missing %q: %s", fragment, normalized)
		}
	}
}
