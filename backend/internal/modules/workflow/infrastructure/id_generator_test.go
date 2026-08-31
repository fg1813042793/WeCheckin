package infrastructure

import (
	"strings"
	"testing"
)

func TestRandomIDGeneratorProducesPrefixedUniqueIDs(t *testing.T) {
	generator := NewRandomIDGenerator()
	first := generator.NewID("task")
	second := generator.NewID("task")
	if first == second {
		t.Fatalf("duplicate ids: %q", first)
	}
	if !strings.HasPrefix(first, "task-") || !strings.HasPrefix(second, "task-") {
		t.Fatalf("ids must keep prefix: %q, %q", first, second)
	}
}
