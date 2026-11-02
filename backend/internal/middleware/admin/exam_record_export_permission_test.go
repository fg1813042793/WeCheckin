package admin

import "testing"

func TestExamRecordExportRequiresDedicatedPermission(t *testing.T) {
	permission, ok := adminRoutePermission("GET", "/api/v2/admin/exams/8/records/99/export")
	if !ok || permission != "exam-record:export" {
		t.Fatalf("permission = %q, %v; want exam-record:export", permission, ok)
	}
}
