package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm/clause"

	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type RunLoggerConfig struct {
	MaxLogSegmentBytes int
	MaxLogRunBytes     int
	Now                func() time.Time
}

type GormRunLoggerFactory struct {
	store  *GormStore
	config RunLoggerConfig
}

func NewGormRunLoggerFactory(store *GormStore, cfg RunLoggerConfig) *GormRunLoggerFactory {
	if cfg.MaxLogSegmentBytes <= 0 {
		cfg.MaxLogSegmentBytes = 16 * 1024
	}
	if cfg.MaxLogRunBytes <= 0 {
		cfg.MaxLogRunBytes = 1024 * 1024
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &GormRunLoggerFactory{store: store, config: cfg}
}

func (factory *GormRunLoggerFactory) ForRun(runID string) application.RunLogger {
	return &gormRunLogger{factory: factory, runID: strings.TrimSpace(runID)}
}

type gormRunLogger struct {
	factory *GormRunLoggerFactory
	runID   string
}

func (logger *gormRunLogger) Log(ctx context.Context, level, stage, content string) error {
	if logger == nil || logger.factory == nil || logger.factory.store == nil || logger.runID == "" {
		return errors.New("scheduled task run logger is not initialized")
	}
	level = normalizeLogLevel(level)
	stage = truncateUTF8(strings.TrimSpace(stage), 40)
	content = truncateUTF8(content, logger.factory.config.MaxLogSegmentBytes)
	if content == "" {
		return nil
	}
	now := logger.factory.config.Now().UnixMilli()
	return logger.factory.store.InTransaction(ctx, func(tx *GormStore) error {
		db, cancel, err := tx.contextDB(ctx)
		if err != nil {
			return err
		}
		defer cancel()
		var run scheduledtaskmodel.Run
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&run, "id = ?", logger.runID).Error; err != nil {
			return err
		}
		var aggregate struct {
			Sequence   int   `gorm:"column:sequence"`
			TotalBytes int64 `gorm:"column:total_bytes"`
		}
		if err := db.Model(&scheduledtaskmodel.RunLog{}).
			Select("COALESCE(MAX(log_sequence), 0) AS sequence, COALESCE(SUM(OCTET_LENGTH(log_content)), 0) AS total_bytes").
			Where("run_id = ?", logger.runID).Scan(&aggregate).Error; err != nil {
			return err
		}
		remaining := int64(logger.factory.config.MaxLogRunBytes) - aggregate.TotalBytes
		if remaining <= 0 {
			return nil
		}
		if int64(len(content)) > remaining {
			content = truncateUTF8(content, int(remaining))
		}
		if content == "" {
			return nil
		}
		return db.Create(&scheduledtaskmodel.RunLog{
			RunID: logger.runID, Sequence: aggregate.Sequence + 1,
			Level: level, Stage: stage, Content: content, LogTime: now, AddTime: now,
		}).Error
	})
}

type LoggingExecutor struct {
	Registry *application.HandlerRegistry
	Loggers  *GormRunLoggerFactory
}

func (executor LoggingExecutor) Execute(ctx context.Context, run *scheduledtaskmodel.Run) (application.HandlerResult, error) {
	if executor.Registry == nil {
		return application.HandlerResult{}, errors.New("scheduled task handler registry is not initialized")
	}
	var logger application.RunLogger
	if executor.Loggers != nil && run != nil {
		logger = executor.Loggers.ForRun(run.ID)
		_ = logger.Log(ctx, "info", "execute", "execution started")
	}
	result, err := executor.Registry.ExecuteWithLogger(ctx, run, logger)
	if logger != nil {
		if err != nil {
			_ = logger.Log(ctx, "error", "complete", "execution failed")
		} else {
			_ = logger.Log(ctx, "info", "complete", "execution completed")
		}
	}
	return result, err
}

func normalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "info", "warn", "error":
		return strings.ToLower(strings.TrimSpace(level))
	default:
		return "info"
	}
}

func truncateUTF8(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len(value) <= limit {
		return value
	}
	value = value[:limit]
	for value != "" && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

var _ application.RunLogger = (*gormRunLogger)(nil)
