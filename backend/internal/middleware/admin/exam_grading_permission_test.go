package admin

import (
	"testing"

	"wecheckin/backend/internal/support/adminmenuperm"
	"wecheckin/backend/internal/support/adminrouteperm"
)

func TestExamGradingRequiresDedicatedPermission(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		permission, ok := adminRoutePermission(method, "/api/v2/admin/exams/8/records/99/grading")
		if !ok || permission != "exam-record:grade" {
			t.Errorf("%s grading permission = %q, %v", method, permission, ok)
		}
	}
	for _, method := range []string{"PUT", "PATCH", "DELETE"} {
		if permission, ok := adminRoutePermission(method, "/api/v2/admin/exams/8/records/99/grading"); ok {
			t.Errorf("unexpected %s grading permission: %q", method, permission)
		}
	}
	if permission, ok := adminRoutePermission("GET", "/api/v2/admin/exams/8/records/99"); !ok || permission != "exam:list" {
		t.Fatalf("record read permission changed: %q, %v", permission, ok)
	}
}

func TestExamGradingPermissionDeclarations(t *testing.T) {
	foundAPI, foundMenu := false, false
	for _, item := range adminrouteperm.Declarations() {
		if item.Key != "admin:api:exam-record:grade" {
			continue
		}
		foundAPI = true
		if item.Perms != "exam-record:grade" || item.CategoryKey != "admin:api-category:exam" || item.Method != "POST" || item.Path != "/api/v2/admin/exams/:id/records/:recordId/grading" {
			t.Fatalf("grading API declaration = %+v", item)
		}
	}
	for _, item := range adminmenuperm.Declarations(true) {
		if item.Key != "admin:menu:exam:record-grade" {
			continue
		}
		foundMenu = true
		if item.Perms != "exam-record:grade" || item.ParentKey != "admin:menu:exam" || item.Type != adminmenuperm.TypeButton || item.Sort != 7 {
			t.Fatalf("grading menu declaration = %+v", item)
		}
	}
	if !foundAPI || !foundMenu {
		t.Fatalf("grading declarations missing: API=%v menu=%v", foundAPI, foundMenu)
	}
	for _, item := range adminmenuperm.Declarations(false) {
		if item.Key == "admin:menu:exam:record-grade" {
			t.Fatal("grading button must follow the exam feature switch")
		}
	}
}
