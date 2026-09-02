package admin

import "testing"

func TestScheduledTaskRESTRoutesResolvePermissions(t *testing.T) {
	for _, test := range []struct {
		method string
		path   string
		want   string
	}{
		{method: "GET", path: "/api/v2/admin/scheduled-tasks", want: "scheduled-task:list"},
		{method: "POST", path: "/api/v2/admin/scheduled-tasks/cron-preview", want: "scheduled-task:list"},
		{method: "GET", path: "/api/v2/admin/scheduled-task-handlers", want: "scheduled-task:list"},
		{method: "GET", path: "/api/v2/admin/scheduled-task-workers", want: "scheduled-task:worker:list"},
		{method: "PATCH", path: "/api/v2/admin/scheduled-tasks/9/status", want: "scheduled-task:status"},
		{method: "POST", path: "/api/v2/admin/scheduled-tasks/9/run", want: "scheduled-task:run"},
		{method: "POST", path: "/api/v2/admin/scheduled-task-runs/run-1/retry", want: "scheduled-task:run:retry"},
		{method: "POST", path: "/api/v2/admin/scheduled-task-runs/run-1/cancel", want: "scheduled-task:run:cancel"},
	} {
		got, ok := adminRoutePermission(test.method, test.path)
		if !ok || got != test.want {
			t.Fatalf("%s %s = %q/%v, want %q", test.method, test.path, got, ok, test.want)
		}
	}
}
