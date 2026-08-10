package model_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../../internal/model"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestModelPackageHasNoImportTimePrints(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list model files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "fmt.Println") {
			t.Fatalf("%s must not print during package import", file)
		}
	}
}

func TestModelDefinitionsLiveInDomainSubpackages(t *testing.T) {
	required := []string{
		filepath.Join("account", "user.go"),
		filepath.Join("admin", "admin.go"),
		filepath.Join("content", "content.go"),
		filepath.Join("dingtalkh5", "performance.go"),
		filepath.Join("interaction", "enroll.go"),
		filepath.Join("interaction", "event.go"),
		filepath.Join("organization", "position.go"),
		filepath.Join("organization", "rbac.go"),
		filepath.Join("permission", "permission.go"),
		filepath.Join("system", "dict.go"),
		filepath.Join("system", "setup.go"),
		filepath.Join("assessment", "exam.go"),
		filepath.Join("assessment", "survey.go"),
	}
	for _, file := range required {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("model definitions must be grouped by domain; missing %s: %v", file, err)
		}
	}

	allowedRootFiles := map[string]bool{
		"aliases.go": true,
		"doc.go":     true,
		"models.go":  true,
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list model root files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if !allowedRootFiles[file] {
			t.Fatalf("model root should only keep docs, aliases and tests; move %s into a domain subpackage", file)
		}
	}
}
