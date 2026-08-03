package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wecheckin/backend/pkg/database"
)

func TestStartupInitializationDoesNotDropLegacyUserFormFields(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list bootstrap files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := strings.ToLower(string(src))
		if strings.Contains(text, "drop table") {
			t.Fatalf("%s must not contain DROP TABLE statements", file)
		}
		if strings.Contains(text, "user_form_fields") {
			t.Fatalf("%s must not clean legacy user_form_fields table", file)
		}
	}
}

func TestBootstrapStartupCodeStaysSplitByResponsibility(t *testing.T) {
	required := []string{"migrate.go", "seed_setup.go", "seed_menu.go"}
	for _, file := range required {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("bootstrap startup code must keep %s split by responsibility: %v", file, err)
		}
	}
}

func TestStartupNoLongerUsesAutoMigrateEnvironmentSwitch(t *testing.T) {
	for _, file := range []string{"migrate.go", "maintenance.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "WECHECKIN_AUTO_MIGRATE") {
			t.Fatalf("%s must not use WECHECKIN_AUTO_MIGRATE; run backend/init.sh for maintenance tasks", file)
		}
	}
}

func TestMaintenanceInitializationReturnsErrors(t *testing.T) {
	src, err := os.ReadFile("maintenance.go")
	if err != nil {
		t.Fatalf("read maintenance.go: %v", err)
	}

	text := string(src)
	if !strings.Contains(text, "func RunMaintenance(options MaintenanceOptions) error") {
		t.Fatalf("RunMaintenance must return an error so maintenance failures are visible")
	}
	for _, snippet := range []string{"Migration warning", "continuing"} {
		if strings.Contains(text, snippet) {
			t.Fatalf("RunMaintenance must not hide migration failures with %q", snippet)
		}
	}
}

func TestRawMigrationSQLErrorsAreChecked(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}

	for i, line := range strings.Split(string(src), "\n") {
		if strings.Contains(line, "database.DB.Exec(") && !strings.Contains(line, ".Error") {
			t.Fatalf("migrate.go:%d raw migration SQL must check .Error: %s", i+1, strings.TrimSpace(line))
		}
	}
}

func TestBootstrapDatabaseOperationsUseQueryContext(t *testing.T) {
	files := []string{"migrate.go", "seed_menu.go", "seed_setup.go"}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "database.DB.") {
			t.Fatalf("%s startup database operations must use database.WithContext", file)
		}
	}
}

func TestBootstrapDatabaseContextUsesStartupTimeout(t *testing.T) {
	if startupDatabaseTimeout <= database.DefaultQueryTimeout {
		t.Fatalf("startup database timeout must exceed request query timeout: got %s, default query %s", startupDatabaseTimeout, database.DefaultQueryTimeout)
	}

	files := []string{"migrate.go", "seed_menu.go", "seed_setup.go"}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		if strings.Contains(text, "database.WithContext(context.Background())") {
			t.Fatalf("%s must not use request-level database.WithContext for startup operations", file)
		}
		if !strings.Contains(text, "startupDB(context.Background())") {
			t.Fatalf("%s startup database operations must use startupDB(context.Background())", file)
		}
	}
}

func TestMenuSeedUsesIdempotentPermissionUpserts(t *testing.T) {
	src, err := os.ReadFile("seed_menu.go")
	if err != nil {
		t.Fatalf("read seed_menu.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissionsupport.SyncAdminMenuPermissionsContext(context.Background(), db, enableExam)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("menu seed must be idempotent permission upsert with %s", snippet)
		}
	}
}
