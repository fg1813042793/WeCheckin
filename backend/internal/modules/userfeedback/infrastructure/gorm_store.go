package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	notificationmodel "wecheckin/backend/internal/model/notification"
	userfeedbackmodel "wecheckin/backend/internal/model/userfeedback"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/pkg/database"
)

var (
	errFeedbackDatabaseNotInitialized = errors.New("user feedback database is not initialized")
	errFeedbackSequenceOverflow       = errors.New("user feedback daily sequence overflow")
	errFeedbackSequenceUpdate         = errors.New("user feedback daily sequence update failed")
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (store *GormStore) InTransaction(ctx context.Context, fn func(application.TransactionStore) error) error {
	if store == nil || store.db == nil {
		return errFeedbackDatabaseNotInitialized
	}
	if fn == nil {
		return errors.New("user feedback transaction callback is nil")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()

	tx := store.db.WithContext(queryCtx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	txStore := &GormStore{db: tx}
	if callbackErr := fn(txStore); callbackErr != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return errors.Join(callbackErr, rollbackErr)
		}
		return callbackErr
	}
	if commitErr := tx.Commit().Error; commitErr != nil {
		return fmt.Errorf("%w: %w", application.ErrTransactionOutcomeUnknown, commitErr)
	}
	return nil
}

func (store *GormStore) NextFeedbackNumber(ctx context.Context, dateKey string) (uint64, error) {
	sequenceDate, err := parseSequenceDate(dateKey)
	if err != nil {
		return 0, err
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	now := time.Now().UTC().UnixMilli()
	seed := userfeedbackmodel.DailySequence{SequenceDate: sequenceDate, UpdatedAt: now}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sequence_date"}},
		DoNothing: true,
	}).Create(&seed).Error; err != nil {
		return 0, err
	}

	var locked userfeedbackmodel.DailySequence
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("sequence_date = ?", dateKey).
		Take(&locked).Error; err != nil {
		return 0, err
	}
	next, err := incrementSequence(locked.CurrentValue)
	if err != nil {
		return 0, err
	}
	result := db.Model(&userfeedbackmodel.DailySequence{}).
		Where("sequence_date = ?", dateKey).
		Updates(map[string]any{"current_value": next, "updated_at": now})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errFeedbackSequenceUpdate
	}
	return next, nil
}

func parseSequenceDate(dateKey string) (time.Time, error) {
	if len(dateKey) != len("2006-01-02") {
		return time.Time{}, fmt.Errorf("invalid feedback sequence date %q", dateKey)
	}
	value, err := time.Parse("2006-01-02", dateKey)
	if err != nil || value.Format("2006-01-02") != dateKey {
		return time.Time{}, fmt.Errorf("invalid feedback sequence date %q", dateKey)
	}
	return value, nil
}

func incrementSequence(current uint64) (uint64, error) {
	if current == math.MaxUint64 {
		return 0, errFeedbackSequenceOverflow
	}
	return current + 1, nil
}

func (store *GormStore) CountCreatedByUserOnDate(ctx context.Context, userID uint, startMs, endMs int64) (int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	var count int64
	err = db.Model(&userfeedbackmodel.Feedback{}).
		Where("submitter_id = ?", userID).
		Where("created_at >= ?", startMs).
		Where("created_at < ?", endMs).
		Count(&count).Error
	return count, err
}

func (store *GormStore) CreateFeedback(ctx context.Context, record application.FeedbackRecord) (uint64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	row := feedbackModel(record)
	result := db.Clauses(feedbackCreateConflictClause()).Create(&row)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, application.ErrDuplicateRequest
	}
	return row.ID, nil
}

func feedbackCreateConflictClause() clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: "submitter_id"}, {Name: "create_request_id"}},
		DoNothing: true,
	}
}

func (store *GormStore) LockFeedback(ctx context.Context, id uint64) (*application.FeedbackSnapshot, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row feedbackSnapshotRow
	err = db.Model(&userfeedbackmodel.Feedback{}).
		Select("user_feedbacks.*, (SELECT COUNT(*) FROM user_feedback_attachments WHERE user_feedback_attachments.feedback_id = user_feedbacks.id) AS attachment_count").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_feedbacks.id = ?", id).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, application.ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	snapshot := feedbackSnapshot(row)
	return &snapshot, nil
}

func (store *GormStore) AppendMessage(ctx context.Context, record application.MessageRecord) (uint64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	row := messageModel(record)
	result := db.Clauses(messageCreateConflictClause()).Create(&row)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, application.ErrDuplicateRequest
	}
	return row.ID, nil
}

func messageCreateConflictClause() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{
			{Name: "feedback_id"},
			{Name: "author_type"},
			{Name: "author_id"},
			{Name: "request_id"},
		},
		DoNothing: true,
	}
}

func (store *GormStore) AppendAttachments(ctx context.Context, records []application.AttachmentRecord) error {
	if store == nil || store.db == nil {
		return errFeedbackDatabaseNotInitialized
	}
	if len(records) == 0 {
		return nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	rows := attachmentModels(records)
	return db.Create(&rows).Error
}

func (store *GormStore) UpdateSnapshot(ctx context.Context, snapshot application.FeedbackSnapshot, expectedVersion uint64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	result := db.Model(&userfeedbackmodel.Feedback{}).
		Where("id = ? AND version = ?", snapshot.ID, expectedVersion).
		Updates(map[string]any{
			"feedback_no":      snapshot.FeedbackNo,
			"submitter_id":     snapshot.SubmitterID,
			"feedback_status":  string(snapshot.Status),
			"handler_id":       snapshot.HandlerID,
			"version":          snapshot.Version,
			"last_activity_at": snapshot.LastActivityAt,
			"resolved_at":      snapshot.ResolvedAt,
			"closed_at":        snapshot.ClosedAt,
			"created_at":       snapshot.CreatedAt,
			"updated_at":       snapshot.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrVersionConflict
	}
	return nil
}

func (store *GormStore) EnqueueNotification(ctx context.Context, record application.NotificationOutboxRecord) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	row, err := notificationOutboxModel(record)
	if err != nil {
		return err
	}
	return db.Clauses(notificationCreateConflictClause()).Create(&row).Error
}

func notificationCreateConflictClause() clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}
}

type feedbackNotificationPayload struct {
	notificationoutboxapp.MessagePayload
	NotificationType string `json:"notificationType"`
}

func notificationOutboxModel(record application.NotificationOutboxRecord) (notificationmodel.Outbox, error) {
	recipientJSON, err := json.Marshal(notificationoutboxapp.InternalRecipient{
		UserIDs: []uint{record.RecipientUserID},
	})
	if err != nil {
		return notificationmodel.Outbox{}, err
	}
	payloadJSON, err := json.Marshal(feedbackNotificationPayload{
		MessagePayload: notificationoutboxapp.MessagePayload{
			Title:      record.Title,
			Content:    record.Content,
			SourceType: record.SourceType,
			SourceID:   record.SourceID,
		},
		NotificationType: record.NotificationType,
	})
	if err != nil {
		return notificationmodel.Outbox{}, err
	}
	return notificationmodel.Outbox{
		IdempotencyKey: record.IdempotencyKey,
		SourceType:     record.SourceType,
		SourceID:       record.SourceID,
		Channel:        record.Channel,
		RecipientJSON:  string(recipientJSON),
		PayloadJSON:    string(payloadJSON),
		Status:         notificationmodel.StatusPending,
		Attempts:       0,
		NextRetryAt:    0,
		AddTime:        record.CreatedAt,
		EditTime:       record.CreatedAt,
	}, nil
}

func (store *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if store == nil || store.db == nil {
		return nil, func() {}, errFeedbackDatabaseNotInitialized
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return store.db.WithContext(queryCtx), cancel, nil
}

var (
	_ application.Store            = (*GormStore)(nil)
	_ application.TransactionStore = (*GormStore)(nil)
)
