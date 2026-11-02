package admin

import "testing"

func TestNotificationTemplatePermissionsSeparateSendingAndEditing(t *testing.T) {
	for _, tt := range []struct{ method, path, want string }{
		{"GET", "/api/v2/admin/dingtalk-notification-templates", "notification:style:list"},
		{"GET", "/api/v2/admin/dingtalk-notification-templates/options", "notification:dingtalk:send"},
		{"GET", "/api/v2/admin/dingtalk-notification-templates/workflow-options", "notification:style:edit"},
		{"POST", "/api/v2/admin/dingtalk-notification-templates", "notification:style:edit"},
		{"PUT", "/api/v2/admin/dingtalk-notification-templates/42", "notification:style:edit"},
		{"DELETE", "/api/v2/admin/dingtalk-notification-templates/42", "notification:style:edit"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			got, ok := adminRoutePermission(tt.method, tt.path)
			if !ok || got != tt.want {
				t.Fatalf("permission = %q (%v), want %q", got, ok, tt.want)
			}
		})
	}
}
