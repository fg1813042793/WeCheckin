package performance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../../../../internal/service/dingtalkh5/performance"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestDingTalkH5PerformanceUsesLayeredSubpackages(t *testing.T) {
	for _, file := range []string{
		filepath.Join("review", "service.go"),
		filepath.Join("template", "service.go"),
		filepath.Join("user", "service.go"),
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("dingtalk h5 performance facade must live in subpackage %s: %v", file, err)
		}
	}

	for _, dir := range []string{"auth", "config", "domain", "notification", "workbench"} {
		if _, err := os.Stat(dir); err == nil {
			t.Fatalf("performance package should only contain performance handler-aligned facades; move %s to the matching dingtalkh5 service package", dir)
		}
	}

	for _, file := range []string{
		filepath.Join("review", "reviews.go"),
		filepath.Join("review", "identity.go"),
		filepath.Join("review", "template_helpers.go"),
		filepath.Join("review", "review_query.go"),
		filepath.Join("review", "review_create.go"),
		filepath.Join("review", "review_employee.go"),
		filepath.Join("review", "review_flow.go"),
		filepath.Join("review", "notification", "notification.go"),
		filepath.Join("review", "scope", "scope.go"),
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("review implementation must be split by responsibility in performance/review package, missing %s: %v", file, err)
		}
	}

	for _, file := range []string{
		filepath.Join("review", "auth.go"),
		filepath.Join("review", "sso.go"),
		filepath.Join("review", "self_bind.go"),
		filepath.Join("review", "bootstrap_permissions.go"),
		filepath.Join("review", "permission_snapshot.go"),
		filepath.Join("review", "workbench.go"),
		filepath.Join("review", "defaults.go"),
		filepath.Join("review", "users.go"),
		filepath.Join("review", "user_store.go"),
		filepath.Join("review", "notification.go"),
		filepath.Join("review", "data_scope.go"),
		filepath.Join("review", "review_scope.go"),
	} {
		if _, err := os.Stat(file); err == nil {
			t.Fatalf("review package should not keep non-review dingtalk h5 service file %s", file)
		}
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"aliases.go":                true,
		"package_structure_test.go": true,
	}
	for _, file := range files {
		base := filepath.Base(file)
		if allowed[base] || strings.HasSuffix(base, "_test.go") {
			continue
		}
		t.Fatalf("performance root should only keep public facade files and tests; move %s into a business subpackage", base)
	}
}
