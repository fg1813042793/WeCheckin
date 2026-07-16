package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestAutoMigrateCanBeDisabledFromEnvironment(t *testing.T) {
	t.Setenv("WECHECKIN_AUTO_MIGRATE", "false")
	if autoMigrateEnabled() {
		t.Fatalf("auto migrate must be disabled when WECHECKIN_AUTO_MIGRATE=false")
	}

	t.Setenv("WECHECKIN_AUTO_MIGRATE", "0")
	if autoMigrateEnabled() {
		t.Fatalf("auto migrate must be disabled when WECHECKIN_AUTO_MIGRATE=0")
	}

	t.Setenv("WECHECKIN_AUTO_MIGRATE", "true")
	if !autoMigrateEnabled() {
		t.Fatalf("auto migrate must stay enabled by default-compatible truthy values")
	}
}

func TestBusinessInitializationReturnsStartupErrors(t *testing.T) {
	src, err := os.ReadFile("database.go")
	if err != nil {
		t.Fatalf("read database.go: %v", err)
	}

	text := string(src)
	if !strings.Contains(text, "func InitBusiness(enableExam bool) error") {
		t.Fatalf("InitBusiness must return an error so startup can fail on migration problems")
	}
	for _, snippet := range []string{"Migration warning", "continuing"} {
		if strings.Contains(text, snippet) {
			t.Fatalf("InitBusiness must not hide startup migration failures with %q", snippet)
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

func TestMenuSeedOnlyRunsBeforeFirstInitialization(t *testing.T) {
	cases := []struct {
		name          string
		existingMenus int64
		markerValue   string
		want          bool
	}{
		{name: "empty database before first startup", existingMenus: 0, markerValue: "", want: true},
		{name: "menus already exist without marker", existingMenus: 3, markerValue: "", want: false},
		{name: "marker already set", existingMenus: 0, markerValue: "1", want: false},
		{name: "truthy marker already set", existingMenus: 0, markerValue: "true", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldSeedMenus(tc.existingMenus, tc.markerValue); got != tc.want {
				t.Fatalf("shouldSeedMenus(%d, %q) = %v, want %v", tc.existingMenus, tc.markerValue, got, tc.want)
			}
		})
	}
}
