package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestWorkflowInstanceUpgradePermissionMigration(t *testing.T) {
	data, err := os.ReadFile("20261102100000_add_workflow_instance_upgrade_permission.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, snippet := range []string{
		"admin:menu:workflow:instance:upgrade",
		"admin:api:workflow:instance:upgrade-preview",
		"admin:api:workflow:instance:upgrade",
		"workflow:instance:upgrade:preview",
		"workflow:instance:upgrade",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("migration missing %q", snippet)
		}
	}
	if strings.Contains(sql, "INSERT INTO `permission_grants`") {
		t.Fatal("instance upgrade permission must not be auto-granted")
	}
}
