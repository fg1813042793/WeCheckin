package client

import (
	"os"
	"strings"
	"testing"
)

func TestClientRoutePermissionMapsKnownV2ClientRoutes(t *testing.T) {
	tests := []struct {
		method string
		path   string
		key    string
	}{
		{"GET", "/api/v2/me/bootstrap", "client:api:bootstrap:view"},
		{"GET", "/api/v2/me", "client:api:user:view"},
		{"POST", "/api/v2/enrollments/12/submissions", "client:api:enroll:submit"},
		{"POST", "/api/v2/events/8/scores", "client:api:event:score"},
		{"PUT", "/api/v2/exam-records/3/answers", "client:api:exam:answer"},
	}
	for _, tt := range tests {
		key, ok := clientRoutePermission(tt.method, tt.path)
		if !ok || key != tt.key {
			t.Fatalf("%s %s mapped to %q ok=%v, want %q", tt.method, tt.path, key, ok, tt.key)
		}
	}
}

func TestClientRoutePermissionMapsLegacyAuthenticatedClientRoutes(t *testing.T) {
	tests := []struct {
		method string
		path   string
		key    string
	}{
		{"GET", "/passport/my_detail", "client:api:user:view"},
		{"POST", "/passport/edit_base", "client:api:user:edit"},
		{"GET", "/fav/my_list", "client:api:favorite:view"},
		{"POST", "/fav/update", "client:api:favorite:edit"},
		{"GET", "/news/view", "client:api:news:view"},
		{"POST", "/enroll/enroll_submit", "client:api:enroll:submit"},
		{"POST", "/event/score_save", "client:api:event:score"},
		{"GET", "/survey/my_response", "client:api:survey:response"},
		{"POST", "/exam/save_answer", "client:api:exam:answer"},
	}
	for _, tt := range tests {
		key, ok := clientRoutePermission(tt.method, tt.path)
		if !ok || key != tt.key {
			t.Fatalf("%s %s mapped to %q ok=%v, want %q", tt.method, tt.path, key, ok, tt.key)
		}
	}
}

func TestClientRoutePermissionRejectsUndeclaredRoutes(t *testing.T) {
	if key, ok := clientRoutePermission("GET", "/api/v2/unknown"); ok || key != "" {
		t.Fatalf("unknown client route must not be implicitly allowed, got %q ok=%v", key, ok)
	}
}

func TestClientPermFailsClosedWhenPermissionTablesAreMissing(t *testing.T) {
	src, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	text := string(src)
	forbidden := "if !permissionsupport.TablesReady(db) {\n\t\t\tc.Next(ctx)\n\t\t\treturn\n\t\t}"
	if strings.Contains(text, forbidden) {
		t.Fatalf("ClientPerm must deny when unified permission tables are unavailable")
	}
	required := "if !permissionsupport.TablesReady(db) {\n\t\t\tc.JSON(consts.StatusOK, utils.H{\"code\": 1, \"msg\": \"无权限访问\"})\n\t\t\tc.Abort()\n\t\t\treturn\n\t\t}"
	if !strings.Contains(text, required) {
		t.Fatalf("ClientPerm must fail closed when unified permission tables are unavailable")
	}
}

func TestClientPermFailsClosedWhenClientAPIPermissionsAreNotReady(t *testing.T) {
	src, err := os.ReadFile("permission.go")
	if err != nil {
		t.Fatalf("read permission.go: %v", err)
	}
	text := string(src)
	forbidden := "if !ready {\n\t\t\tc.Next(ctx)\n\t\t\treturn\n\t\t}"
	if strings.Contains(text, forbidden) {
		t.Fatalf("ClientPerm must deny when the current subject has no client API grants")
	}
	required := "if !ready {\n\t\t\tc.JSON(consts.StatusOK, utils.H{\"code\": 1, \"msg\": \"无权限访问\"})\n\t\t\tc.Abort()\n\t\t\treturn\n\t\t}"
	if !strings.Contains(text, required) {
		t.Fatalf("ClientPerm must fail closed when the current subject has no client API grants")
	}
}
