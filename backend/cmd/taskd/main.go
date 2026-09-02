package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"wecheckin/backend/internal/config"
	scheduledtaskinfra "wecheckin/backend/internal/modules/scheduledtask/infrastructure"
	scheduledtaskruntime "wecheckin/backend/internal/modules/scheduledtask/runtime"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowinfra "wecheckin/backend/internal/modules/workflow/infrastructure"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
	redispkg "wecheckin/backend/pkg/redis"
)

var taskdVersion = "dev"

type role string

const (
	roleScheduler role = "scheduler"
	roleWorker    role = "worker"
	roleAll       role = "all"
)

type component interface {
	Run(context.Context) error
}

func parseRole(value string) (role, error) {
	switch role(value) {
	case roleScheduler, roleWorker, roleAll:
		return role(value), nil
	default:
		return "", fmt.Errorf("invalid taskd role %q", value)
	}
}

func runRole(ctx context.Context, selected role, scheduler, worker component) error {
	switch selected {
	case roleScheduler:
		if scheduler == nil {
			return errors.New("scheduler component is required")
		}
		return superviseComponent(ctx, "scheduler", scheduler)
	case roleWorker:
		if worker == nil {
			return errors.New("worker component is required")
		}
		return superviseComponent(ctx, "worker", worker)
	case roleAll:
		if scheduler == nil || worker == nil {
			return errors.New("scheduler and worker components are required")
		}
		var group sync.WaitGroup
		errorsCh := make(chan error, 2)
		for _, item := range []struct {
			name      string
			component component
		}{{name: "scheduler", component: scheduler}, {name: "worker", component: worker}} {
			item := item
			group.Add(1)
			go func() {
				defer group.Done()
				errorsCh <- superviseComponent(ctx, item.name, item.component)
			}()
		}
		group.Wait()
		close(errorsCh)
		var errs []error
		for err := range errorsCh {
			if err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	default:
		return fmt.Errorf("unsupported taskd role %q", selected)
	}
}

func superviseComponent(ctx context.Context, name string, target component) error {
	delay := time.Second
	const maximumDelay = 30 * time.Second
	for {
		err := target.Run(ctx)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err == nil {
			err = errors.New("component stopped unexpectedly")
		}
		waitDelay := jitteredTaskdDelay(delay)
		log.Printf("taskd component %s stopped, restarting in %s: %v", name, waitDelay, err)
		timer := time.NewTimer(waitDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if delay < maximumDelay {
			delay *= 2
			if delay > maximumDelay {
				delay = maximumDelay
			}
		}
	}
}

func waitForDependency(ctx context.Context, name string, connect func() error, retryDelays ...time.Duration) error {
	initialDelay := time.Second
	maximumDelay := 30 * time.Second
	if len(retryDelays) > 0 && retryDelays[0] > 0 {
		initialDelay = retryDelays[0]
	}
	if len(retryDelays) > 1 && retryDelays[1] >= initialDelay {
		maximumDelay = retryDelays[1]
	}
	delay := initialDelay
	failures := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := connect()
		if err == nil {
			if failures > 0 {
				log.Printf("taskd dependency %s recovered after %d failure(s)", name, failures)
			}
			return nil
		}
		failures++
		waitDelay := jitteredTaskdDelay(delay)
		log.Printf("taskd dependency %s unavailable, retrying in %s: %v", name, waitDelay, err)
		timer := time.NewTimer(waitDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if delay < maximumDelay {
			delay *= 2
			if delay > maximumDelay {
				delay = maximumDelay
			}
		}
	}
}

func jitteredTaskdDelay(base time.Duration) time.Duration {
	minimum := base * 4 / 5
	spread := base - minimum
	if spread <= 0 {
		return base
	}
	return minimum + time.Duration(rand.Int63n(int64(spread)+1))
}

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("taskd stopped: %v", err)
	}
}

func run() error {
	env := flag.String("env", "", "运行环境 (dev/prod)")
	roleValue := flag.String("role", string(roleAll), "运行角色 (scheduler/worker/all)")
	workerIDValue := flag.String("worker-id", "", "Worker 节点 ID；同一集群内必须唯一")
	flag.Parse()

	selectedRole, err := parseRole(*roleValue)
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfig(*env)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := logger.Init(cfg.Log.Dir, cfg.Log.Level, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	if err := waitForDependency(ctx, "database", func() error {
		return database.ConnectDatabase(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	}); err != nil {
		return err
	}
	if err := waitForDependency(ctx, "Redis", func() error {
		return redispkg.Init(cfg.Redis)
	}); err != nil {
		return err
	}
	defer redispkg.Close()

	db := database.GetDB()
	db.Logger = db.Logger.LogMode(gormlogger.Warn)
	workflowStore := workflowinfra.NewGormStore(db)
	notificationRepository := workflowinfra.NewGormNotificationRepository(db)
	notificationDispatcher := workflowapp.NewNotificationDispatcher(
		notificationRepository,
		workflowinfra.NewDingTalkNotificationChannel(db, nil),
	)
	workflowRuntime := workflowapp.NewServiceWithNotifications(
		workflowStore,
		workflowinfra.NewAssigneeResolver(db),
		workflowinfra.NewRandomIDGenerator(),
		workflowapp.DefaultLifecycleEventPublisher(),
		notificationDispatcher,
	)
	taskStore := scheduledtaskinfra.NewGormStore(db)
	registry, err := scheduledtaskinfra.NewHandlerRegistry(
		cfg.ScheduledTask,
		workflowRuntime,
		scheduledtaskinfra.NewCleanupJob(taskStore, cfg.ScheduledTask.RunRetentionDays, cfg.ScheduledTask.LogRetentionDays, nil),
		scheduledtaskinfra.NewWorkflowNotificationDispatchJob(workflowRuntime),
	)
	if err != nil {
		return fmt.Errorf("initialize handlers: %w", err)
	}
	queue := scheduledtaskinfra.NewRedisStreamQueue(redispkg.RDB, cfg.ScheduledTask.RedisKeyPrefix)

	scheduler := scheduledtaskruntime.NewScheduler(taskStore, queue, scheduledtaskruntime.SchedulerConfig{
		PollInterval:     time.Duration(cfg.ScheduledTask.SchedulerPollSeconds) * time.Second,
		RecoveryInterval: time.Duration(cfg.ScheduledTask.SchedulerRecoverySeconds) * time.Second,
		BatchSize:        100, DispatchRecoveryAge: 30 * time.Second,
		StaleRunAge: time.Duration(cfg.ScheduledTask.RecoveryTimeoutSeconds) * time.Second,
	})
	workerID := *workerIDValue
	if workerID == "" {
		workerID = defaultWorkerID()
	}
	executor := scheduledtaskinfra.LoggingExecutor{
		Registry: registry,
		Loggers: scheduledtaskinfra.NewGormRunLoggerFactory(taskStore, scheduledtaskinfra.RunLoggerConfig{
			MaxLogSegmentBytes: cfg.ScheduledTask.MaxLogSegmentBytes,
			MaxLogRunBytes:     cfg.ScheduledTask.MaxLogRunBytes,
		}),
	}
	worker := scheduledtaskruntime.NewWorker(taskStore, queue, executor, scheduledtaskruntime.WorkerConfig{
		WorkerID: workerID, Role: string(selectedRole), Version: taskdVersion,
		WorkerCount:       cfg.ScheduledTask.WorkerCount,
		PollBlock:         time.Duration(cfg.ScheduledTask.WorkerPollBlockSeconds) * time.Second,
		ClaimIdle:         time.Duration(cfg.ScheduledTask.RecoveryTimeoutSeconds) * time.Second,
		HeartbeatInterval: time.Duration(cfg.ScheduledTask.WorkerHeartbeatSeconds) * time.Second,
		WorkerTTL:         time.Duration(cfg.ScheduledTask.WorkerTTLSeconds) * time.Second,
	})

	logger.Logger.Printf("taskd starting role=%s workerId=%s version=%s", selectedRole, workerID, taskdVersion)
	return runRole(ctx, selectedRole, scheduler, worker)
}

func defaultWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}
