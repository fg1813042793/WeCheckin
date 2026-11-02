package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamNotificationTemplateRoutesAndPermissions(t *testing.T) {
	routes := readExamTemplateContractFile(t, "routes.go")
	middleware := readExamTemplateContractFile(t, filepath.Join("..", "..", "..", "middleware", "admin", "route_permissions.go"))
	catalog := readExamTemplateContractFile(t, filepath.Join("..", "..", "..", "support", "adminrouteperm", "catalog.go"))
	for _, snippet := range []string{`admin.GET("/exam-notification-templates", aExam.NotificationTemplatesGet)`, `admin.PUT("/exam-notification-templates", aExam.NotificationTemplatesSave)`} {
		if !strings.Contains(routes, snippet) {
			t.Fatalf("exam template routes missing %q", snippet)
		}
	}
	for _, snippet := range []string{"GET /api/v2/admin/exam-notification-templates", "PUT /api/v2/admin/exam-notification-templates", "exam-template:list", "exam-template:edit"} {
		if !strings.Contains(middleware, snippet) {
			t.Fatalf("exam template middleware missing %q", snippet)
		}
	}
	for _, snippet := range []string{`{"exam-template:list", "考试通知模版查看接口"}`, `{"exam-template:edit", "考试通知模版维护接口"}`, `"exam-template:list":`, `"exam-template:edit":`} {
		if !strings.Contains(catalog, snippet) {
			t.Fatalf("exam template API catalog missing %q", snippet)
		}
	}
}

func readExamTemplateContractFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
