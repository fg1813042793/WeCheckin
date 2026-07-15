package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlersDoNotHardcodeTokenRedisKeys(t *testing.T) {
	forbidden := []string{
		`"user_token:a:"`,
		`"admin_token:a:"`,
		`"user_token:s:"`,
		`"admin_token:s:"`,
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob handler files: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s must not hardcode token redis key %s", file, needle)
			}
		}
	}
}
