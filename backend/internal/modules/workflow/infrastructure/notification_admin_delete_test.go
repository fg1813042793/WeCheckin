package infrastructure

import (
	"reflect"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
)

func TestApplyActiveNotificationFilterHidesAdminDeletedRecords(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	var rows []workflowmodel.NotificationOutbox
	statement := applyActiveNotificationFilter(db.Model(&workflowmodel.NotificationOutbox{})).Find(&rows).Statement
	sqlText := strings.Join(strings.Fields(statement.SQL.String()), " ")

	if !strings.Contains(sqlText, "admin_deleted_at = ?") {
		t.Fatalf("active notification query must hide soft-deleted records: %s", sqlText)
	}
	if !reflect.DeepEqual(statement.Vars, []interface{}{int64(0)}) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, []interface{}{int64(0)})
	}
}

func TestNotificationDeleteRejectsSendingStatus(t *testing.T) {
	for _, test := range []struct {
		status string
		want   bool
	}{
		{status: workflowmodel.NotificationStatusPending, want: true},
		{status: workflowmodel.NotificationStatusFailed, want: true},
		{status: workflowmodel.NotificationStatusDead, want: true},
		{status: workflowmodel.NotificationStatusSent, want: true},
		{status: workflowmodel.NotificationStatusSending, want: false},
	} {
		if got := notificationCanBeDeleted(test.status); got != test.want {
			t.Fatalf("notificationCanBeDeleted(%q) = %v, want %v", test.status, got, test.want)
		}
	}
}
