package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartTaskdScriptDefaultsToAllAndAllowsRoleOverride(t *testing.T) {
	const path = "../../start-taskd.sh"
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat start-taskd.sh: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatal("start-taskd.sh must be executable")
	}

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read start-taskd.sh: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		`Go 1.24`,
		`export GOCACHE="${GOCACHE:-${BACKEND_DIR}/../.cache/go-build}"`,
		`TASKD_ROLE="${TASKD_ROLE:-all}"`,
		`--role|--role=*`,
		`TASKD_ARGS=("--role=${TASKD_ROLE}" "${TASKD_ARGS[@]}")`,
		`go run ./cmd/taskd "${TASKD_ARGS[@]}"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("start-taskd.sh must contain %q", snippet)
		}
	}
	if strings.Contains(text, `go run ./cmd "$@"`) {
		t.Fatal("start-taskd.sh must not start the HTTP service")
	}
}

func TestStartTaskdScriptPropagatesTaskdFailure(t *testing.T) {
	binDir := t.TempDir()
	fakeGo := filepath.Join(binDir, "go")
	if err := os.WriteFile(fakeGo, []byte("#!/bin/sh\nif [ \"$1 $2\" = \"mod download\" ]; then exit 0; fi\nexit 23\n"), 0o755); err != nil {
		t.Fatalf("write fake go command: %v", err)
	}

	command := exec.Command("bash", "../../start-taskd.sh", "--help")
	command.Env = append(os.Environ(), "PATH="+binDir+":/usr/bin:/bin")
	err := command.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("start-taskd.sh error = %v, want exit status 23", err)
	}
	if exitError.ExitCode() != 23 {
		t.Fatalf("start-taskd.sh exit code = %d, want 23", exitError.ExitCode())
	}
}
