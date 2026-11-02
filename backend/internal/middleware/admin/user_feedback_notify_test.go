package admin

import "testing"

func TestFeedbackNotificationRequiresExistingHandlePermission(t *testing.T) {
	got, ok := adminRoutePermission("POST", "/api/v2/admin/user-feedbacks/7/notify")
	if !ok || got != "user-feedback:handle" {
		t.Fatalf("notification permission=%q declared=%v", got, ok)
	}
}
