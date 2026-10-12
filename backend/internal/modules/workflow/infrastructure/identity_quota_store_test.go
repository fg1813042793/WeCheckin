package infrastructure

import (
	"reflect"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/workflowcore"
)

func TestStartQuotaUsageQueryCountsOnlyCompletedNonDeletedInstances(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	window := workflowcore.StartLimitWindow{StartsAt: 1000, EndsAt: 2000}
	var count int64
	statement := startQuotaUsageQuery(db, 9, " 7 ", window).Count(&count).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{
		"definition_id = ? AND starter_id = ?",
		"instance_status = ?",
		"admin_deleted_at = ?",
		"start_time >= ?",
		"start_time < ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("quota query missing %q: %s", fragment, sqlText)
		}
	}
	if strings.Contains(sqlText, "starter_deleted_at") {
		t.Fatalf("starter history deletion must not refund quota: %s", sqlText)
	}
	wantVars := []interface{}{uint(9), "7", workflowmodel.InstanceStatusCompleted, int64(0), int64(1000), int64(2000)}
	if !reflect.DeepEqual(statement.Vars, wantVars) {
		t.Fatalf("quota query vars = %#v, want %#v", statement.Vars, wantVars)
	}
}

func TestRunningStartInstanceQueryBlocksOnlyActiveNonDeletedInstances(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	var count int64
	statement := runningStartInstanceQuery(db, 9, " 7 ").Count(&count).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{
		"definition_id = ? AND starter_id = ?",
		"instance_status = ?",
		"admin_deleted_at = ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("running instance query missing %q: %s", fragment, sqlText)
		}
	}
	wantVars := []interface{}{uint(9), "7", workflowmodel.InstanceStatusRunning, int64(0)}
	if !reflect.DeepEqual(statement.Vars, wantVars) {
		t.Fatalf("running instance query vars = %#v, want %#v", statement.Vars, wantVars)
	}
}
