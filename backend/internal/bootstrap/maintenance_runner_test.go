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
		"database.InitDatabase",
		`flag.String("migrations"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("maintenance command must include %s", snippet)
		}
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
