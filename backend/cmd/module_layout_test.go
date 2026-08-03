package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoModuleLivesInBackend(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "go.mod"))
	if err != nil {
		t.Fatalf("backend/go.mod should exist: %v", err)
	}
	if !strings.HasPrefix(string(data), "module wecheckin/backend\n") {
		t.Fatalf("backend/go.mod should declare module wecheckin/backend, got first line %q", firstLine(string(data)))
	}

	for _, name := range []string{"go.mod", "go.sum"} {
		_, err := os.Stat(filepath.Join("..", "..", name))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("root %s should not exist after moving Go module into backend/", name)
		}
	}
}

func firstLine(text string) string {
	if idx := strings.IndexByte(text, '\n'); idx >= 0 {
		return text[:idx]
	}
	return text
}
