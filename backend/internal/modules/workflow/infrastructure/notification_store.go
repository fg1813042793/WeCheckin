package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/pkg/database"
)

var ErrNotificationNotFound = errors.New("通知投递记录不存在")

const (
	workflowNotificationType       = "workflow"
	workflowNotificationSourceType = "workflow_instance"
)

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (repository *GormNotificationRepository) List(ctx context.Context, query application.NotificationQuery) (*application.NotificationList, error) {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	page, pageSize := normalizeNotificationPage(query.Page, query.PageSize)
	statement := db.Model(&workflowmodel.NotificationOutbox{})
	if value := strings.TrimSpace(query.InstanceID); value != "" {
		statement = statement.Where("instance_id = ?", value)
	}
	if value := strings.TrimSpace(query.RecipientUserID); value != "" {
		statement = statement.Where("recipient_user_id = ?", value)
	}
	if value := strings.TrimSpace(query.Kind); value != "" {
		statement = statement.Where("notification_kind = ?", value)
	}
	if value := strings.TrimSpace(query.Channel); value != "" {
		statement = statement.Where("notification_channel = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		statement = statement.Where("notification_status = ?", value)
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []workflowmodel.NotificationOutbox
	if err := statement.Order("add_time DESC").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	records, err := notificationRecordsFromModels(rows)
	if err != nil {
		return nil, err
	}
	return &application.NotificationList{List: records, Total: total, Page: page, PageSize: pageSize}, nil
}

func (repository *GormNotificationRepository) ClaimByIDs(ctx context.Context, ids []string, now, staleBefore int64) ([]application.NotificationRecord, error) {
	cleanIDs := uniqueNonEmptyStrings(ids)
	if len(cleanIDs) == 0 {
		return []application.NotificationRecord{}, nil
	}
	return repository.claim(ctx, now, staleBefore, 0, func(query *gorm.DB) *gorm.DB {
		return query.Where("id IN ?", cleanIDs)
	})
}

func (repository *GormNotificationRepository) ClaimDue(ctx context.Context, now, staleBefore int64, limit int) ([]application.NotificationRecord, error) {
	if limit <= 0 {
		return []application.NotificationRecord{}, nil
	}
	return repository.claim(ctx, now, staleBefore, limit, func(query *gorm.DB) *gorm.DB { return query })
}

func (repository *GormNotificationRepository) claim(
	ctx context.Context,
	now, staleBefore int64,
	limit int,
	filter func(*gorm.DB) *gorm.DB,
) ([]application.NotificationRecord, error) {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var claimed []application.NotificationRecord
	err = db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&workflowmodel.NotificationOutbox{})
		query = filter(query)
		query = applyClaimableNotificationFilter(query, now, staleBefore)
		query = query.Clauses(clause.Locking{Strength: "UPDATE"}).Order("next_retry_at ASC").Order("add_time ASC").Order("id ASC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		var rows []workflowmodel.NotificationOutbox
		if err := query.Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			claimed = []application.NotificationRecord{}
			return nil
		}
		ids := make([]string, 0, len(rows))
		for index := range rows {
			ids = append(ids, rows[index].ID)
			rows[index].Status = workflowmodel.NotificationStatusSending
			rows[index].EditTime = now
		}
		if err := tx.Model(&workflowmodel.NotificationOutbox{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"notification_status": workflowmodel.NotificationStatusSending,
			"edit_time":           now,
		}).Error; err != nil {
			return err
		}
		var err error
		claimed, err = notificationRecordsFromModels(rows)
		return err
	})
	return claimed, err
}

func (repository *GormNotificationRepository) DeliverInApp(ctx context.Context, notification application.NotificationRecord, now int64) error {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var row workflowmodel.NotificationOutbox
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", notification.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotificationNotFound
			}
			return err
		}
		if row.Status == workflowmodel.NotificationStatusSent {
			return nil
		}
		if row.Status != workflowmodel.NotificationStatusSending {
			return fmt.Errorf("通知状态 %q 不可投递", row.Status)
		}
		record, err := notificationRecordFromModel(row)
		if err != nil {
			return err
		}
		if err := tx.Create(workflowInAppNotify(record, now)).Error; err != nil {
			return err
		}
		return tx.Model(&workflowmodel.NotificationOutbox{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
			"notification_status": workflowmodel.NotificationStatusSent,
			"provider_message_id": "",
			"last_error":          "",
			"next_retry_at":       0,
			"sent_at":             now,
			"edit_time":           now,
		}).Error
	})
}

func workflowInAppNotify(notification application.NotificationRecord, now int64) *model.Notify {
	return &model.Notify{
		Title: notification.Payload.Title, Content: notification.Payload.Content, Type: workflowNotificationType,
		SourceID: notification.InstanceID, SourceType: workflowNotificationSourceType,
		UserID: notification.RecipientUserID, IsRead: 0, AddTime: now,
	}
}

func (repository *GormNotificationRepository) MarkSent(ctx context.Context, id, providerMessageID string, now int64) error {
	return repository.updateClaimed(ctx, id, map[string]interface{}{
		"notification_status": workflowmodel.NotificationStatusSent,
		"provider_message_id": strings.TrimSpace(providerMessageID),
		"last_error":          "",
		"next_retry_at":       0,
		"sent_at":             now,
		"edit_time":           now,
	})
}

func (repository *GormNotificationRepository) MarkFailed(ctx context.Context, id string, attempts int, status string, nextRetryAt int64, message string, now int64) error {
	if status != workflowmodel.NotificationStatusFailed && status != workflowmodel.NotificationStatusDead {
		return fmt.Errorf("通知失败状态无效: %s", status)
	}
	return repository.updateClaimed(ctx, id, map[string]interface{}{
		"notification_status": status,
		"attempts":            attempts,
		"next_retry_at":       nextRetryAt,
		"last_error":          message,
		"edit_time":           now,
	})
}

func (repository *GormNotificationRepository) ResetForRetry(ctx context.Context, id string, now int64) error {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&workflowmodel.NotificationOutbox{}).
		Where("id = ? AND notification_status IN ?", strings.TrimSpace(id), []string{workflowmodel.NotificationStatusFailed, workflowmodel.NotificationStatusDead}).
		Updates(map[string]interface{}{
			"notification_status": workflowmodel.NotificationStatusPending,
			"attempts":            0,
			"next_retry_at":       now,
			"last_error":          "",
			"sent_at":             0,
			"edit_time":           now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (repository *GormNotificationRepository) updateClaimed(ctx context.Context, id string, updates map[string]interface{}) error {
	db, cancel, err := repository.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&workflowmodel.NotificationOutbox{}).
		Where("id = ? AND notification_status = ?", strings.TrimSpace(id), workflowmodel.NotificationStatusSending).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (repository *GormNotificationRepository) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if repository == nil || repository.db == nil {
		return nil, func() {}, errors.New("工作流通知数据库未初始化")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return repository.db.WithContext(queryCtx), cancel, nil
}

func applyClaimableNotificationFilter(query *gorm.DB, now, staleBefore int64) *gorm.DB {
	return query.Where(
		"((notification_status = ? AND next_retry_at <= ?) OR (notification_status = ? AND next_retry_at <= ?) OR (notification_status = ? AND edit_time <= ?))",
		workflowmodel.NotificationStatusPending, now,
		workflowmodel.NotificationStatusFailed, now,
		workflowmodel.NotificationStatusSending, staleBefore,
	)
}

func notificationRecordFromModel(row workflowmodel.NotificationOutbox) (application.NotificationRecord, error) {
	var payload application.NotificationPayload
	if err := json.Unmarshal([]byte(row.PayloadJSON), &payload); err != nil {
		return application.NotificationRecord{}, fmt.Errorf("解析通知 %s 负载失败: %w", row.ID, err)
	}
	return application.NotificationRecord{
		ID: row.ID, InstanceID: row.InstanceID, NodeID: row.NodeID, TaskID: row.TaskID,
		RecipientUserID: row.RecipientUserID, Kind: row.Kind, Channel: row.Channel,
		Status: row.Status, Payload: payload, CorpID: row.CorpID,
		ProviderMessageID: row.ProviderMessageID, Attempts: row.Attempts,
		NextRetryAt: row.NextRetryAt, LastError: row.LastError, SentAt: row.SentAt,
		AddTime: row.AddTime, EditTime: row.EditTime,
	}, nil
}

func notificationRecordsFromModels(rows []workflowmodel.NotificationOutbox) ([]application.NotificationRecord, error) {
	records := make([]application.NotificationRecord, 0, len(rows))
	for _, row := range rows {
		record, err := notificationRecordFromModel(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func normalizeNotificationPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
