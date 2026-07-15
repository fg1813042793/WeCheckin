package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
