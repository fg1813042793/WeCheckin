package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRoleSupportsSchedulerWorkerAndAll(t *testing.T) {
	for _, value := range []string{"scheduler", "worker", "all"} {
		role, err := parseRole(value)
		if err != nil || string(role) != value {
			t.Fatalf("parseRole(%q) = %q, %v", value, role, err)
		}
	}
	if _, err := parseRole("http"); err == nil {
		t.Fatal("HTTP is not a taskd role")
	}
}

func TestRunRoleAllStartsSchedulerAndWorker(t *testing.T) {
	scheduler := &fakeComponent{}
	worker := &fakeComponent{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runRole(ctx, roleAll, scheduler, worker); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if scheduler.calls != 1 || worker.calls != 1 {
		t.Fatalf("scheduler/worker calls = %d/%d", scheduler.calls, worker.calls)
	}
}

func TestRunRoleAllRestartsFailedComponentWithoutCancelingSibling(t *testing.T) {
	scheduler := &recoveringComponent{restarted: make(chan struct{}, 1)}
	worker := &blockingComponent{canceled: make(chan struct{}, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runRole(ctx, roleAll, scheduler, worker) }()

	select {
	case <-scheduler.restarted:
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("scheduler was not restarted after an unexpected failure")
	}
	select {
	case <-worker.canceled:
		cancel()
		t.Fatal("worker was canceled when only the scheduler failed")
	default:
	}

	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("runRole() error = %v, want context cancellation", err)
	}
	if scheduler.calls.Load() < 2 {
		t.Fatalf("scheduler calls = %d, want at least 2", scheduler.calls.Load())
	}
}

func TestWaitForDependencyRetriesUntilConnectionRecovers(t *testing.T) {
	attempts := 0
	err := waitForDependency(context.Background(), "database", func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary network failure")
		}
		return nil
	}, time.Millisecond, 2*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("connection attempts = %d, want 3", attempts)
	}
}

func TestTaskdMainWiresDedicatedRuntimeWithoutHTTPServer(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		"signal.NotifyContext(",
		"database.ConnectDatabaseWithOptions(",
		"database.Options{",
		"ConnectTimeout:",
		"ReadTimeout:",
		"WriteTimeout:",
		"waitForDependency(ctx, \"database\"",
		"waitForDependency(ctx, \"Redis\"",
		"LogLevel:",
		"gormlogger.Warn",
		"redispkg.Init(",
		"scheduledtaskruntime.NewScheduler(",
		"RecoveryInterval:",
		"scheduledtaskruntime.NewWorker(",
		"scheduledtaskinfra.NewHandlerRegistry(",
		"scheduledtaskinfra.NewNotificationOutboxDispatchJob(",
		"notificationoutboxapp.NewService(",
		"notificationoutboxinfra.NewWebhookChannel(",
		"runRole(ctx, selectedRole, scheduler, worker)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("taskd main missing %q", want)
		}
	}
	if strings.Index(text, "signal.NotifyContext(") > strings.Index(text, "waitForDependency(ctx, \"database\"") {
		t.Fatal("taskd must create its cancellation context before waiting for the database")
	}
	for _, forbidden := range []string{"app/server", "server.Default(", ".Spin(", "server.WithHostPorts(", "database.InitDatabase(", ".LogMode("} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("taskd must not consume HTTP service resources: found %q", forbidden)
		}
	}
}

type fakeComponent struct{ calls int }

func (component *fakeComponent) Run(ctx context.Context) error {
	component.calls++
	return ctx.Err()
}

type recoveringComponent struct {
	calls     atomic.Int32
	restarted chan struct{}
}

func (component *recoveringComponent) Run(ctx context.Context) error {
	if component.calls.Add(1) == 1 {
		return errors.New("temporary component failure")
	}
	select {
	case component.restarted <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return ctx.Err()
}

type blockingComponent struct{ canceled chan struct{} }

func (component *blockingComponent) Run(ctx context.Context) error {
	<-ctx.Done()
	select {
	case component.canceled <- struct{}{}:
	default:
	}
	return ctx.Err()
}
