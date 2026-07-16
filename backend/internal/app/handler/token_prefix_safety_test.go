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

	err := filepath.WalkDir(".", func(file string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		text := string(src)
		for _, needle := range forbidden {
			if strings.Contains(text, needle) {
				t.Fatalf("%s must not hardcode token redis key %s", file, needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk handler files: %v", err)
	}
}
