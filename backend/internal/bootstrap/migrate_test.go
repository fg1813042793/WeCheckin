package bootstrap

import (
	"errors"
	"strings"
	"testing"
)

func TestRunMigrationStepsLogsSuccessfulSteps(t *testing.T) {
	var logs []string
	steps := []migrationStep{
		{Name: "first", Run: func() error { return nil }},
		{Name: "second", Run: func() error { return nil }},
	}

	if err := runMigrationSteps(steps, func(format string, args ...interface{}) {
		logs = append(logs, format)
	}); err != nil {
		t.Fatalf("run migration steps: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("expected two migration logs, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "first") || !strings.Contains(logs[1], "second") {
		t.Fatalf("migration logs should include step names: %#v", logs)
	}
}

func TestRunMigrationStepsWrapsFailedStepName(t *testing.T) {
	expected := errors.New("boom")
	steps := []migrationStep{
		{Name: "ok", Run: func() error { return nil }},
		{Name: "bad_step", Run: func() error { return expected }},
		{Name: "never", Run: func() error { t.Fatal("should stop after first failure"); return nil }},
	}

	err := runMigrationSteps(steps, nil)
	if err == nil {
		t.Fatal("expected migration failure")
	}
	if !strings.Contains(err.Error(), "bad_step") || !errors.Is(err, expected) {
		t.Fatalf("failure should wrap step name and original error, got %v", err)
	}
}
