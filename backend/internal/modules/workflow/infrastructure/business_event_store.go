package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/pkg/database"
)

type GormBusinessEventRepository struct {
	db *gorm.DB
}

func NewGormBusinessEventRepository(db *gorm.DB) *GormBusinessEventRepository {
	return &GormBusinessEventRepository{db: db}
}

func (store *GormStore) CreateBusinessEvent(ctx context.Context, event application.WorkflowBusinessEvent) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	row, err := businessEventModel(event, database.Now())
	if err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (repository *GormBusinessEventRepository) ClaimDue(
	ctx context.Context,
	now, staleBefore int64,
	limit int,
) ([]application.BusinessEventRecord, error) {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	if limit <= 0 {
		return []application.BusinessEventRecord{}, nil
	}

	var records []application.BusinessEventRecord
	err = db.Transaction(func(tx *gorm.DB) error {
		var rows []workflowmodel.BusinessEventOutbox
		if err := claimableBusinessEventQuery(tx, now, staleBefore).Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			records = []application.BusinessEventRecord{}
			return nil
		}
		ids := make([]string, 0, len(rows))
		for index := range rows {
			ids = append(ids, rows[index].ID)
			rows[index].Status = workflowmodel.BusinessEventStatusSending
			rows[index].EditTime = now
		}
		if err := tx.Model(&workflowmodel.BusinessEventOutbox{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"event_status": workflowmodel.BusinessEventStatusSending,
			"edit_time":    now,
		}).Error; err != nil {
			return err
		}
		records = make([]application.BusinessEventRecord, 0, len(rows))
		for _, row := range rows {
			record, err := businessEventRecordFromModel(row)
			if err != nil {
				return err
			}
			records = append(records, record)
		}
		return nil
	})
	return records, err
}

func claimableBusinessEventQuery(db *gorm.DB, now, staleBefore int64) *gorm.DB {
	return db.Model(&workflowmodel.BusinessEventOutbox{}).
		Where(
			"((event_status = ? AND next_retry_at <= ?) OR (event_status = ? AND next_retry_at <= ?) OR (event_status = ? AND edit_time <= ?))",
			workflowmodel.BusinessEventStatusPending, now,
			workflowmodel.BusinessEventStatusFailed, now,
			workflowmodel.BusinessEventStatusSending, staleBefore,
		).
		Order("next_retry_at ASC").Order("add_time ASC").Order("id ASC").
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
}

func (repository *GormBusinessEventRepository) MarkSent(ctx context.Context, id string, now int64) error {
	return repository.updateClaimed(ctx, id, map[string]interface{}{
		"event_status":  workflowmodel.BusinessEventStatusSent,
		"last_error":    "",
		"next_retry_at": 0,
		"sent_at":       now,
		"edit_time":     now,
	})
}

func (repository *GormBusinessEventRepository) MarkFailed(
	ctx context.Context,
	id string,
	attempts int,
	status string,
	nextRetryAt int64,
	message string,
	now int64,
) error {
	if status != workflowmodel.BusinessEventStatusFailed && status != workflowmodel.BusinessEventStatusDead {
		return fmt.Errorf("业务事件失败状态无效: %s", status)
	}
	return repository.updateClaimed(ctx, id, map[string]interface{}{
		"event_status":  status,
		"attempts":      attempts,
		"next_retry_at": nextRetryAt,
		"last_error":    message,
		"edit_time":     now,
	})
}

func (repository *GormBusinessEventRepository) updateClaimed(ctx context.Context, id string, updates map[string]interface{}) error {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&workflowmodel.BusinessEventOutbox{}).
		Where("id = ? AND event_status = ?", strings.TrimSpace(id), workflowmodel.BusinessEventStatusSending).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("工作流业务事件不存在或未被领取")
	}
	return nil
}

func (repository *GormBusinessEventRepository) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if repository == nil || repository.db == nil {
		return nil, func() {}, errors.New("工作流业务事件数据库未初始化")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return repository.db.WithContext(queryCtx), cancel, nil
}

func businessEventModel(event application.WorkflowBusinessEvent, now int64) (workflowmodel.BusinessEventOutbox, error) {
	if strings.TrimSpace(event.ID) == "" || strings.TrimSpace(event.DedupeKey) == "" {
		return workflowmodel.BusinessEventOutbox{}, errors.New("工作流业务事件标识不能为空")
	}
	payload, err := json.Marshal(event.Event)
	if err != nil {
		return workflowmodel.BusinessEventOutbox{}, fmt.Errorf("序列化工作流业务事件失败: %w", err)
	}
	return workflowmodel.BusinessEventOutbox{
		ID: strings.TrimSpace(event.ID), EventType: string(event.Event.Type),
		AggregateID: event.Event.InstanceID, BusinessType: event.Event.BusinessType,
		BusinessKey: event.Event.BusinessKey, PayloadJSON: string(payload),
		Status: workflowmodel.BusinessEventStatusPending, DedupeKey: strings.TrimSpace(event.DedupeKey),
		NextRetryAt: now, AddTime: now, EditTime: now,
	}, nil
}

func businessEventRecordFromModel(row workflowmodel.BusinessEventOutbox) (application.BusinessEventRecord, error) {
	var event application.LifecycleEvent
	if err := json.Unmarshal([]byte(row.PayloadJSON), &event); err != nil {
		return application.BusinessEventRecord{}, fmt.Errorf("解析工作流业务事件 %s 失败: %w", row.ID, err)
	}
	return application.BusinessEventRecord{ID: row.ID, Event: event, Attempts: row.Attempts}, nil
}

var _ application.BusinessEventRepository = (*GormBusinessEventRepository)(nil)
