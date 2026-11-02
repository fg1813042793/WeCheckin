package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamRecordExportRoutePermissionAndMigration(t *testing.T) {
	routes := readExamRecordExportContractFile(t, "routes.go")
	middleware := readExamRecordExportContractFile(t, filepath.Join("..", "..", "..", "middleware", "admin", "route_permissions.go"))
	catalog := readExamRecordExportContractFile(t, filepath.Join("..", "..", "..", "support", "adminrouteperm", "catalog.go"))
	menu := readExamRecordExportContractFile(t, filepath.Join("..", "..", "..", "support", "adminmenuperm", "declarations.go"))
	swagger := readExamRecordExportContractFile(t, filepath.Join("..", "swagger", "swagger.go"))
	migration := readExamRecordExportContractFile(t, filepath.Join("..", "..", "..", "..", "migrations", "20261031110000_add_exam_record_export_permission.sql"))

	for _, snippet := range []string{`admin.GET("/exams/:id/records/:recordId/export"`, `aExam.RecordExport`} {
		if !strings.Contains(routes, snippet) {
			t.Fatalf("exam record export route missing %q", snippet)
		}
	}
	for _, snippet := range []string{`path: "/api/v2/admin/exams/:id/records/:recordId/export"`, `perm: "exam-record:export"`} {
		if !strings.Contains(middleware, snippet) {
			t.Fatalf("exam record export middleware missing %q", snippet)
		}
	}
	for _, snippet := range []string{`{"exam-record:export", "考试答卷导出接口"}`, `"exam-record:export":`} {
		if !strings.Contains(catalog, snippet) {
			t.Fatalf("exam record export catalog missing %q", snippet)
		}
	}
	for _, snippet := range []string{"admin:menu:exam:record-export", "exam-record:export"} {
		if !strings.Contains(menu, snippet) || !strings.Contains(migration, snippet) {
			t.Fatalf("exam record export permission declaration missing %q", snippet)
		}
	}
	if !strings.Contains(swagger, "@Router /api/v2/admin/exams/{id}/records/{recordId}/export [get]") {
		t.Fatal("exam record export swagger declaration is required")
	}
}

func readExamRecordExportContractFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
