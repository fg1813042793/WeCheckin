package httpadmin

import (
	"encoding/json"
	"testing"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
)

func TestRiskPermissionRequiresNotificationSendForInAppNotificationJob(t *testing.T) {
	permission := riskPermission(
		scheduledtaskmodel.HandlerTypeGo,
		json.RawMessage(`{"handlerKey":"notification.in_app.send","params":{}}`),
	)
	if permission != "notification:send" {
		t.Fatalf("riskPermission() = %q, want notification:send", permission)
	}
}

func TestRiskPermissionDoesNotApplyNotificationPermissionToOtherGoJobs(t *testing.T) {
	permission := riskPermission(
		scheduledtaskmodel.HandlerTypeGo,
		json.RawMessage(`{"handlerKey":"scheduled-task.cleanup","params":{}}`),
	)
	if permission != "" {
		t.Fatalf("riskPermission() = %q, want empty", permission)
	}
}
