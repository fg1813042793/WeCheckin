package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestPublishedDefinitionQueriesExcludeSoftDeletedDefinitions(t *testing.T) {
	source, err := os.ReadFile("definition_store.go")
	if err != nil {
		t.Fatalf("read definition store: %v", err)
	}
	text := string(source)
	if strings.Count(text, "definition_deleted_at = ?") < 3 {
		t.Fatal("published definition load, list and detail queries must exclude soft-deleted workflows")
	}
}
