package migrations

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminrouteperm"
)

func TestAuditLogMigrationPreservesLegacyAndPermissionIsolation(t *testing.T) {
	data, err := os.ReadFile("20261028180000_expand_audit_logs.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	snapshotSQL, err := os.ReadFile("20261029150000_add_audit_snapshot_status.sql")
	if err != nil {
		t.Fatal(err)
	}
	schemaSQL := sql + string(snapshotSQL)
	for _, part := range []string{"INFORMATION_SCHEMA.COLUMNS", "INFORMATION_SCHEMA.STATISTICS", "CONCAT('legacy-', `id`)", "WHEN `log_type` = 1 THEN 'login' ELSE 'operation'", "`log_outcome` = 'unknown'", "`log_source` = 'admin'", "`log_actor_type` = 'admin'", "UNIQUE INDEX `uidx_logs_event_id`", "ON DUPLICATE KEY UPDATE"} {
		if !strings.Contains(sql, part) {
			t.Errorf("migration missing %q", part)
		}
	}
	for _, forbidden := range []string{"DROP TABLE", "DROP COLUMN", "DELETE FROM", "TRUNCATE", "INSERT INTO `permission_grants`", "CREATE TABLE `logs`"} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("unsafe migration: %s", forbidden)
		}
	}
	typ := reflect.TypeOf(model.Log{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name == "ID" {
			break
		}
		for _, tag := range strings.Split(field.Tag.Get("gorm"), ";") {
			if strings.HasPrefix(tag, "column:") && !strings.Contains(schemaSQL, "ADD COLUMN `"+strings.TrimPrefix(tag, "column:")+"`") {
				t.Errorf("column missing for %s", field.Name)
			}
		}
	}
	for _, code := range []string{"log:security", "log:export", "log:security-export"} {
		for _, prefix := range []string{"admin:api:", "admin:menu:"} {
			if !strings.Contains(sql, "'"+prefix+code+"'") {
				t.Errorf("missing %s%s", prefix, code)
			}
		}
		found := false
		for _, declaration := range adminrouteperm.Declarations() {
			if declaration.Perms == code {
				found = true
				if !strings.Contains(sql, "'"+declaration.Path+"'") {
					t.Errorf("route missing: %s", declaration.Path)
				}
			}
		}
		if !found {
			t.Errorf("catalog missing %s", code)
		}
	}
}
