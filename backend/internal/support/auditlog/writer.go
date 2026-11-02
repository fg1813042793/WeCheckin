package auditlog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm/clause"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

type Writer interface {
	Write(context.Context, *model.Log) error
}

type DatabaseWriter struct{}

// Write is synchronous and bounded. Reusing an event ID only finalizes the same request.
func (DatabaseWriter) Write(ctx context.Context, event *model.Log) error {
	if ctx == nil {
		ctx = context.Background()
	}
	writeCtx, timeout := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer timeout()
	db, cancel := database.WithContext(writeCtx)
	defer cancel()
	if db == nil {
		return errors.New("audit database unavailable")
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "log_event_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"log_type", "log_content", "create_by", "update_by", "log_admin_name", "log_admin_desc",
			"log_category", "log_source", "log_module", "log_action", "log_outcome", "log_reason_code", "log_reason",
			"log_actor_type", "log_attempted_account", "log_target_type", "log_target_id", "log_target_name",
			"log_target_ids", "log_requested_count",
			"log_affected_count", "log_request_id", "log_method", "log_path", "log_duration_ms", "log_user_agent",
			"log_changes", "log_related_type", "log_related_id", "edit_time",
			"log_snapshot_status", "log_snapshot_reason",
			"log_browser", "log_device_type", "log_platform_type",
		}),
	}).Create(event).Error
}

func WriteFallback(event *model.Log, stage string) {
	// Events are already allowlisted/redacted; never include the failed SQL or raw request body.
	data, _ := json.Marshal(event)
	logger.Logger.Printf("[AUDIT_WRITE_FAILED] stage=%s event=%s", stage, data)
}
