package infrastructure

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestShellHandlerRequiresGlobalSwitchAndRegisteredCommand(t *testing.T) {
	command := ShellCommand{ExecutablePath: "/usr/bin/printf", WorkingDir: "/tmp", ArgumentPattern: `^[A-Za-z0-9._-]+$`, MaxArgs: 3}
	disabled, err := NewShellHandler(ShellHandlerPolicy{Enabled: false, Commands: map[string]ShellCommand{"print": command}}, &fakeShellRunner{})
	if err != nil {
		t.Fatal(err)
	}
	if err := disabled.ValidateConfig(context.Background(), []byte(`{"commandKey":"print"}`)); !errors.Is(err, ErrShellHandlerDisabled) {
		t.Fatalf("disabled error = %v", err)
	}
	enabled, err := NewShellHandler(ShellHandlerPolicy{Enabled: true, Commands: map[string]ShellCommand{"print": command}}, &fakeShellRunner{})
	if err != nil {
		t.Fatal(err)
	}
	if err := enabled.ValidateConfig(context.Background(), []byte(`{"commandKey":"missing"}`)); !errors.Is(err, ErrShellCommandNotFound) {
		t.Fatalf("unknown command error = %v", err)
	}
}

func TestShellHandlerValidatesArgumentsAndEnvironmentThenRunsWithoutShell(t *testing.T) {
	runner := &fakeShellRunner{stdout: "ok"}
	handler, err := NewShellHandler(ShellHandlerPolicy{
		Enabled: true, MaxOutputBytes: 1024,
		Commands: map[string]ShellCommand{"print": {
			ExecutablePath: "/usr/bin/printf", WorkingDir: "/tmp", ArgumentPattern: `^[A-Za-z0-9._-]+$`, MaxArgs: 3,
			AllowedEnv: []string{"LANG"},
		}},
	}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if err := handler.ValidateConfig(context.Background(), []byte(`{"commandKey":"print","args":["a|cat"]}`)); err == nil {
		t.Fatal("shell syntax in an argument must fail the registered argument rule")
	}
	if err := handler.ValidateConfig(context.Background(), []byte(`{"commandKey":"print","env":{"SECRET":"value"}}`)); err == nil {
		t.Fatal("unregistered environment key must fail")
	}
	result, err := handler.Execute(context.Background(), application.RunContext{
		RunID: "run-shell", Task: application.TaskSnapshot{HandlerConfigJSON: `{"commandKey":"print","args":["hello"],"env":{"LANG":"C"}}`},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runner.path != "/usr/bin/printf" || len(runner.args) != 1 || runner.args[0] != "hello" || runner.env["LANG"] != "C" {
		t.Fatalf("runner invocation = %#v", runner)
	}
	if result.Summary != "ok" {
		t.Fatalf("summary = %q", result.Summary)
	}
}

func TestShellHandlerTruncatesOutputAndHonorsContextCancellation(t *testing.T) {
	runner := &fakeShellRunner{stdout: strings.Repeat("x", 64)}
	handler, err := NewShellHandler(ShellHandlerPolicy{
		Enabled: true, MaxOutputBytes: 16,
		Commands: map[string]ShellCommand{"print": {ExecutablePath: "/usr/bin/printf", WorkingDir: "/tmp", MaxArgs: 1}},
	}, runner)
	if err != nil {
		t.Fatal(err)
	}
	result, err := handler.Execute(context.Background(), application.RunContext{Task: application.TaskSnapshot{HandlerConfigJSON: `{"commandKey":"print"}`}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Summary) > 32 || !strings.Contains(result.Summary, "truncated") {
		t.Fatalf("truncated summary = %q", result.Summary)
	}

	runner.waitForContext = true
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = handler.Execute(ctx, application.RunContext{Task: application.TaskSnapshot{HandlerConfigJSON: `{"commandKey":"print"}`}})
	var handlerError *application.HandlerError
	if err == nil || !errors.As(err, &handlerError) || !handlerError.Temporary {
		t.Fatalf("cancellation error = %#v", err)
	}
}

type fakeShellRunner struct {
	path           string
	args           []string
	dir            string
	env            map[string]string
	stdout         string
	stderr         string
	err            error
	waitForContext bool
}

func (runner *fakeShellRunner) Run(ctx context.Context, path string, args []string, dir string, env map[string]string, _ int64) (string, string, error) {
	runner.path, runner.args, runner.dir, runner.env = path, append([]string(nil), args...), dir, env
	if runner.waitForContext {
		<-ctx.Done()
		return "", "", ctx.Err()
	}
	return runner.stdout, runner.stderr, runner.err
}
