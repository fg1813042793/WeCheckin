package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestSnapshotStatusMigrationPreservesHistoricalUncertainty(t *testing.T) {
	raw, err := os.ReadFile("20261029150000_add_audit_snapshot_status.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(raw)
	for _, required := range []string{"INFORMATION_SCHEMA.COLUMNS", "ADD COLUMN `log_snapshot_status`", "DEFAULT ''legacy_unknown''", "ADD COLUMN `log_snapshot_reason`", "PREPARE stmt"} {
		if !strings.Contains(sql, required) {
			t.Errorf("missing %s", required)
		}
	}
	for _, forbidden := range []string{"UPDATE `logs`", "DROP ", "TRUNCATE ", "DELETE FROM", "INSERT INTO"} {
		if strings.Contains(strings.ToUpper(sql), strings.ToUpper(forbidden)) {
			t.Errorf("unexpected data rewrite %s", forbidden)
		}
	}
}
