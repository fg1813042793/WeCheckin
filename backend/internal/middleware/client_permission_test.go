package middleware

import "testing"

func TestClientRoutePermissionMapsKnownV2ClientRoutes(t *testing.T) {
	tests := []struct {
		method string
		path   string
		key    string
	}{
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

func TestClientRoutePermissionRejectsUndeclaredRoutes(t *testing.T) {
	if key, ok := clientRoutePermission("GET", "/api/v2/unknown"); ok || key != "" {
		t.Fatalf("unknown client route must not be implicitly allowed, got %q ok=%v", key, ok)
	}
}
