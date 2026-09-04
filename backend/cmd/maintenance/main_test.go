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
	for _, required := range []string{"database.ConnectDatabaseWithOptions(", "database.Options{", "log.Fatalf"} {
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
}
