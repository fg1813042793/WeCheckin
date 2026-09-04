package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaintenanceRunnerRecordsExecutedTasks(t *testing.T) {
	src, err := os.ReadFile("maintenance.go")
	if err != nil {
		t.Fatalf("read maintenance.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func RunMaintenance",
		"schema_migrations",
		"runOnce",
		"bootstrap:auto_migrate",
		"bootstrap:seed_setups",
		"bootstrap:seed_permissions",
		"runVersionedSQLMigrations",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("maintenance runner must include %s", snippet)
		}
	}
}

func TestStandaloneMaintenanceCommandExists(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "cmd", "maintenance", "main.go"))
	if err != nil {
		t.Fatalf("read maintenance command: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"bootstrap.RunMaintenance",
		"config.LoadConfig",
		"database.ConnectDatabaseWithOptions",
		"ConnectTimeout:",
		"MaxOpenConns:",
		`flag.String("migrations"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("maintenance command must include %s", snippet)
		}
	}
	if strings.Contains(text, "database.InitDatabase") {
		t.Fatal("maintenance command must use the shared explicit database options")
	}
}

func TestSplitSQLStatementsKeepsSemicolonsInsideStrings(t *testing.T) {
	statements := splitSQLStatements(`
SET @ddl := 'SELECT ''a;b''';
PREPARE stmt FROM @ddl;
EXECUTE stmt;
-- comment with ; should not split
DEALLOCATE PREPARE stmt;
`)
	if len(statements) != 4 {
		t.Fatalf("splitSQLStatements returned %d statements, want 4: %#v", len(statements), statements)
	}
	if !strings.Contains(statements[0], "a;b") {
		t.Fatalf("first statement must keep semicolon inside string: %q", statements[0])
	}
}

func TestAppliedMigrationChecksumRepairIsNarrowlyAllowed(t *testing.T) {
	allowedVersion := "20260731162000_add_dingtalk_h5_review_scope_indexes"
	allowedChecksum := "2385c26de9616e0fb26ebeb7d617f827711d285e8a466a5658895765261f51ec"
	if !isAppliedMigrationChecksumRepairAllowed(allowedVersion, allowedChecksum) {
		t.Fatalf("%s is an idempotent index migration and should allow checksum repair", allowedVersion)
	}
	if isAppliedMigrationChecksumRepairAllowed(allowedVersion, strings.Repeat("0", 64)) {
		t.Fatalf("%s must not allow checksum repair for unexpected content", allowedVersion)
	}

	for _, version := range []string{
		"20260731170000_add_permission_grant_subject_status_index",
		"20260730085500_prepare_dingtalk_h5_legacy_audit_fields",
		"bootstrap:seed_permissions",
		"",
	} {
		if isAppliedMigrationChecksumRepairAllowed(version, allowedChecksum) {
			t.Fatalf("%s must not allow checksum repair", version)
		}
	}
}
