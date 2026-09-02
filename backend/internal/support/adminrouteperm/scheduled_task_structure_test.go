package adminrouteperm

import "testing"

func TestScheduledTaskRoutePermissionCatalog(t *testing.T) {
	want := map[string]string{
		"scheduled-task:list":        "/api/v2/admin/scheduled-tasks",
		"scheduled-task:add":         "/api/v2/admin/scheduled-tasks",
		"scheduled-task:run":         "/api/v2/admin/scheduled-tasks/:id/run",
		"scheduled-task:run:list":    "/api/v2/admin/scheduled-task-runs",
		"scheduled-task:run:retry":   "/api/v2/admin/scheduled-task-runs/:id/retry",
		"scheduled-task:run:cancel":  "/api/v2/admin/scheduled-task-runs/:id/cancel",
		"scheduled-task:worker:list": "/api/v2/admin/scheduled-task-workers",
	}
	found := make(map[string]Declaration)
	for _, item := range Declarations() {
		found[item.Perms] = item
	}
	for perms, path := range want {
		item, ok := found[perms]
		if !ok || item.Path != path {
			t.Fatalf("route catalog %s = %#v, want %s", perms, item, path)
		}
	}
}
