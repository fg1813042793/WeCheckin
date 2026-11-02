package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSurveyNotificationTemplateRoutesAndPermissions(t *testing.T) {
	routes := readSurveyTemplateContractFile(t, "routes.go")
	middleware := readSurveyTemplateContractFile(t, filepath.Join("..", "..", "..", "middleware", "admin", "route_permissions.go"))
	catalog := readSurveyTemplateContractFile(t, filepath.Join("..", "..", "..", "support", "adminrouteperm", "catalog.go"))

	for _, snippet := range []string{
		`admin.GET("/survey-notification-templates", aSurvey.NotificationTemplatesGet)`,
		`admin.PUT("/survey-notification-templates", aSurvey.NotificationTemplatesSave)`,
	} {
		if !strings.Contains(routes, snippet) {
			t.Fatalf("survey template routes missing %q", snippet)
		}
	}
	for _, snippet := range []string{
		`"GET /api/v2/admin/survey-notification-templates"`,
		`"PUT /api/v2/admin/survey-notification-templates"`,
		`"survey-template:list"`,
		`"survey-template:edit"`,
	} {
		if !strings.Contains(middleware, snippet) {
			t.Fatalf("survey template middleware missing %q", snippet)
		}
	}
	for _, snippet := range []string{
		`{"survey-template:list", "问卷通知模版查看接口"}`,
		`{"survey-template:edit", "问卷通知模版维护接口"}`,
		`"survey-template:list":`,
		`"survey-template:edit":`,
		`case "survey", "survey-template", "response", "question-bank":`,
	} {
		if !strings.Contains(catalog, snippet) {
			t.Fatalf("survey template API catalog missing %q", snippet)
		}
	}
}

func readSurveyTemplateContractFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
