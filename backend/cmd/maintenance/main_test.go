package main

import (
	"os"
	"strings"
	"testing"
)

func TestMaintenanceUsesDatabaseOptionsAndStopsOnConnectionFailure(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{"logger.Init(", "database.ConnectDatabaseWithOptions(", "database.Options{", "LogWriter: logger.SQLWriter()", "log.Fatalf"} {
		if !strings.Contains(text, required) {
			t.Fatalf("maintenance main missing %q", required)
		}
	}
	if strings.Contains(text, "database.InitDatabase(") {
		t.Fatal("maintenance must not use fatal database compatibility wrapper")
	}
	if strings.Index(text, "database.ConnectDatabaseWithOptions(") > strings.Index(text, "bootstrap.RunMaintenance(") {
		t.Fatal("maintenance must connect before running migrations")
	}
	if strings.Index(text, "logger.Init(") > strings.Index(text, "database.ConnectDatabaseWithOptions(") {
		t.Fatal("maintenance must initialize file logging before creating the database logger")
	}
}
