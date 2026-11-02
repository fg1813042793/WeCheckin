package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestInstanceVersionUpgradeUsesConditionalRunningUpdate(t *testing.T) {
	source, err := os.ReadFile("runtime_store.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, snippet := range []string{
		"func (store *GormStore) UpdateInstanceDefinitionVersion(",
		`Where("id = ? AND definition_version = ? AND instance_status = ?",`,
		`"definition_version": toVersion`,
		"application.ErrInstanceVersionConflict",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("runtime store missing %q", snippet)
		}
	}
}
