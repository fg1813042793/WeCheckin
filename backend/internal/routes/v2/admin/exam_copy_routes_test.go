package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamCopyRoutePermissionsSwaggerAndMigration(t *testing.T) {
	routes := readExamCopyContractFile(t, "routes.go")
	middleware := readExamCopyContractFile(t, filepath.Join("..", "..", "..", "middleware", "admin", "route_permissions.go"))
	catalog := readExamCopyContractFile(t, filepath.Join("..", "..", "..", "support", "adminrouteperm", "catalog.go"))
	menu := readExamCopyContractFile(t, filepath.Join("..", "..", "..", "support", "adminmenuperm", "declarations.go"))
	swagger := readExamCopyContractFile(t, filepath.Join("..", "swagger", "swagger.go"))
	migration := readExamCopyContractFile(t, filepath.Join("..", "..", "..", "..", "migrations", "20261031120000_add_exam_copy_permission.sql"))

	for _, snippet := range []string{`admin.POST("/exams/:id/copy"`, `aExam.Copy`} {
		if !strings.Contains(routes, snippet) {
			t.Fatalf("exam copy route missing %q", snippet)
		}
	}
	for _, snippet := range []string{`path: "/api/v2/admin/exams/:id/copy"`, `perm: "exam:copy"`} {
		if !strings.Contains(middleware, snippet) {
			t.Fatalf("exam copy middleware missing %q", snippet)
		}
	}
	for _, snippet := range []string{`{"exam:copy", "考试复制接口"}`, `"exam:copy":`} {
		if !strings.Contains(catalog, snippet) {
			t.Fatalf("exam copy API catalog missing %q", snippet)
		}
	}
	for _, snippet := range []string{"admin:menu:exam:copy", "exam:copy"} {
		if !strings.Contains(menu, snippet) || !strings.Contains(migration, snippet) {
			t.Fatalf("exam copy permission declaration missing %q", snippet)
		}
	}
	if !strings.Contains(swagger, "@Router /api/v2/admin/exams/{id}/copy [post]") {
		t.Fatal("exam copy swagger declaration is required")
	}
}

func readExamCopyContractFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
