package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppliedMigrationChecksumRepairAllowlistMatchesCurrentFiles(t *testing.T) {
	for version := range appliedMigrationChecksumRepairAllowlist {
		path := filepath.Join("..", "..", "migrations", version+".sql")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read allowlisted migration %s: %v", version, err)
		}
		digest := sha256.Sum256(content)
		checksum := hex.EncodeToString(digest[:])
		if !isAppliedMigrationChecksumRepairAllowed(version, checksum) {
			t.Fatalf("allowlisted migration %s must use current checksum %s", version, checksum)
		}
	}
}

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
	allowed := map[string]string{
		"20260727174000_seed_mishi_org_users":                 "b559652f583b1e7d8e41fc72135234b7e97e5ffa32044d8646120caac0016aa9",
		"20260731162000_add_dingtalk_h5_review_scope_indexes": "2385c26de9616e0fb26ebeb7d617f827711d285e8a466a5658895765261f51ec",
		"20260901140000_add_workflow_definition_logo":         "28622beda1bdd9f21756b7e6285dd89e745774babcb7219937db9ccb59e4ad96",
		"20260903133000_add_workflow_task_admin_delete":       "aff5bec2c37ba63e9d9fe876739098dced70c9e7b04722c881c09f8b02aaf0d7",
	}
	for version, checksum := range allowed {
		if !isAppliedMigrationChecksumRepairAllowed(version, checksum) {
			t.Fatalf("%s should allow checksum repair for its known previous content", version)
		}
		if isAppliedMigrationChecksumRepairAllowed(version, strings.Repeat("0", 64)) {
			t.Fatalf("%s must not allow checksum repair for unexpected content", version)
		}
	}

	for _, version := range []string{
		"20260731170000_add_permission_grant_subject_status_index",
		"20260730085500_prepare_dingtalk_h5_legacy_audit_fields",
		"bootstrap:seed_permissions",
		"",
	} {
		if isAppliedMigrationChecksumRepairAllowed(version, allowed["20260731162000_add_dingtalk_h5_review_scope_indexes"]) {
			t.Fatalf("%s must not allow checksum repair", version)
		}
	}
}
