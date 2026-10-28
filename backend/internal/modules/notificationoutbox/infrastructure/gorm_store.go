package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	notificationmodel "wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/pkg/database"
)

var ErrOutboxRecordNotClaimed = errors.New("notification outbox record is not claimed")

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (store *GormStore) Enqueue(ctx context.Context, row notificationmodel.Outbox) (bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	result := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "idempotency_key"}}, DoNothing: true}).Create(&row)
	return result.RowsAffected > 0, result.Error
}

func (store *GormStore) ListWebhookRecords(ctx context.Context, query application.WebhookRecordQuery) ([]notificationmodel.Outbox, int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer cancel()
	base := db.Model(&notificationmodel.Outbox{}).Where("notification_channel = ?", notificationmodel.ChannelWebhook)
	if query.Status != "" {
		base = base.Where("notification_status = ?", query.Status)
	}
	if query.SourceType != "" {
		base = base.Where("source_type = ?", query.SourceType)
	}
	if query.WebhookType != "" {
		base = base.Where("JSON_UNQUOTE(JSON_EXTRACT(recipient_json, '$.type')) = ?", query.WebhookType)
	}
	if query.AddTimeFrom > 0 {
		base = base.Where("add_time >= ?", query.AddTimeFrom)
	}
	if query.AddTimeTo > 0 {
		base = base.Where("add_time <= ?", query.AddTimeTo)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []notificationmodel.Outbox
	offset := (query.Page - 1) * query.PageSize
	if err := base.Order("add_time DESC").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (store *GormStore) GetWebhookRecord(ctx context.Context, id uint64) (*notificationmodel.Outbox, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row notificationmodel.Outbox
	if err := db.Where("id = ? AND notification_channel = ?", id, notificationmodel.ChannelWebhook).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, application.ErrWebhookRecordNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (store *GormStore) DeleteWebhookRecord(ctx context.Context, id uint64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Where("id = ? AND notification_channel = ? AND notification_status <> ?", id, notificationmodel.ChannelWebhook, notificationmodel.StatusSending).Delete(&notificationmodel.Outbox{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		row, findErr := store.GetWebhookRecord(ctx, id)
		if findErr != nil || row.Status == notificationmodel.StatusSending {
			return application.ErrWebhookRecordBusy
		}
		return application.ErrWebhookRecordNotFound
	}
	return nil
}

func (store *GormStore) ClaimDue(ctx context.Context, now, staleBefore int64, limit int) ([]notificationmodel.Outbox, error) {
	if limit <= 0 {
		return []notificationmodel.Outbox{}, nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	claimed := make([]notificationmodel.Outbox, 0)
	err = db.Transaction(func(tx *gorm.DB) error {
		query := applyClaimableFilter(tx.Model(&notificationmodel.Outbox{}), now, staleBefore)
		query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Order("next_retry_at ASC").Order("add_time ASC").Order("id ASC").Limit(limit)
		if err := query.Find(&claimed).Error; err != nil {
			return err
		}
		if len(claimed) == 0 {
			return nil
		}
		ids := make([]uint64, 0, len(claimed))
		for index := range claimed {
			ids = append(ids, claimed[index].ID)
			claimed[index].Status = notificationmodel.StatusSending
			claimed[index].EditTime = now
		}
		return tx.Model(&notificationmodel.Outbox{}).Where("id IN ?", ids).Updates(map[string]any{
			"notification_status":  notificationmodel.StatusSending,
			"response_status_code": 0,
			"response_body":        "",
			"edit_time":            now,
		}).Error
	})
	return claimed, err
}

func (store *GormStore) MarkSent(ctx context.Context, id uint64, now int64, delivery application.DeliveryResult) error {
	return store.updateClaimed(ctx, id, map[string]any{
		"notification_status":  notificationmodel.StatusSent,
		"next_retry_at":        0,
		"last_error":           "",
		"sent_at":              now,
		"response_status_code": delivery.ResponseStatusCode,
		"response_body":        delivery.ResponseBody,
		"edit_time":            now,
	})
}

func (store *GormStore) MarkFailed(ctx context.Context, failure application.Failure) error {
	return store.updateClaimed(ctx, failure.ID, map[string]any{
		"notification_status":  failure.Status,
		"attempts":             failure.Attempts,
		"next_retry_at":        failure.NextRetryAt,
		"last_error":           failure.LastError,
		"response_status_code": failure.ResponseStatusCode,
		"response_body":        failure.ResponseBody,
		"edit_time":            failure.EditTime,
	})
}

func (store *GormStore) updateClaimed(ctx context.Context, id uint64, updates map[string]any) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&notificationmodel.Outbox{}).
		Where("id = ? AND notification_status = ?", id, notificationmodel.StatusSending).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutboxRecordNotClaimed
	}
	return nil
}

func (store *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if store == nil || store.db == nil {
		return nil, func() {}, errors.New("notification outbox database is not initialized")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return store.db.WithContext(queryCtx), cancel, nil
}

func applyClaimableFilter(query *gorm.DB, now, staleBefore int64) *gorm.DB {
	return query.Where(
		"((notification_status = ? AND next_retry_at <= ?) OR (notification_status = ? AND next_retry_at <= ?) OR (notification_status = ? AND edit_time <= ?))",
		notificationmodel.StatusPending, now,
		notificationmodel.StatusFailed, now,
		notificationmodel.StatusSending, staleBefore,
	)
}

var _ application.Store = (*GormStore)(nil)
