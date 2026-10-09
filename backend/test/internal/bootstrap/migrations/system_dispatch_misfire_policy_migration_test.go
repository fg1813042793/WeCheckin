package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemDispatchMisfirePolicyMigrationUsesFireOnce(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261009112000_fix_system_dispatch_misfire_policy.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"`misfire_policy` = 'fire_once'",
		"'system.notification-outbox-dispatch'",
		"'system.workflow-business-event-dispatch'",
		"`version` = `version` + 1",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("system dispatch migration missing %q", required)
		}
	}
}
