package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
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
	errFeedbackSequenceSeedInsert     = errors.New("user feedback daily sequence seed insert affected no rows")
	errFeedbackSequenceUpdate         = errors.New("user feedback daily sequence update failed")
	errFeedbackInsert                 = errors.New("user feedback insert affected no rows")
	errFeedbackMessageInsert          = errors.New("user feedback message insert affected no rows")
	errFeedbackAttachmentInsert       = errors.New("user feedback attachment insert affected no rows")
	errFeedbackNotificationInsert     = errors.New("user feedback notification insert affected no rows")
)

const (
	mysqlDuplicateEntryNumber          = uint16(1062)
	sequencePrimaryIndex               = "primary"
	feedbackRequestIndex               = "uk_user_feedbacks_submitter_request"
	feedbackMessageRequestIndex        = "uk_user_feedback_messages_request"
	notificationOutboxIdempotencyIndex = "uk_notification_outbox_idempotency"
)

var (
	mysqlDuplicateKeySuffixPattern = regexp.MustCompile(`(?i)(?:^|[[:space:]])for[[:space:]]+key[[:space:]]+(.+?)[[:space:]]*$`)
	mysqlIndexNamePattern          = regexp.MustCompile(`^[[:alnum:]_$-]+$`)
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
	seedResult := db.Create(&seed)
	if seedResult.Error != nil {
		if !isMySQLDuplicateIndex(seedResult.Error, sequencePrimaryIndex) {
			return 0, seedResult.Error
		}
	} else if seedResult.RowsAffected == 0 {
		return 0, errFeedbackSequenceSeedInsert
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
	result := db.Create(&row)
	if result.Error != nil {
		if isMySQLDuplicateIndex(result.Error, feedbackRequestIndex) {
			return 0, application.ErrDuplicateRequest
		}
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errFeedbackInsert
	}
	return row.ID, nil
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
	result := db.Create(&row)
	if result.Error != nil {
		if isMySQLDuplicateIndex(result.Error, feedbackMessageRequestIndex) {
			return 0, application.ErrDuplicateRequest
		}
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errFeedbackMessageInsert
	}
	return row.ID, nil
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
	result := db.Create(&rows)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errFeedbackAttachmentInsert
	}
	return nil
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
	result := db.Create(&row)
	if result.Error != nil {
		if isMySQLDuplicateIndex(result.Error, notificationOutboxIdempotencyIndex) {
			return nil
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errFeedbackNotificationInsert
	}
	return nil
}

func notificationOutboxModel(record application.NotificationOutboxRecord) (notificationmodel.Outbox, error) {
	recipientJSON, err := json.Marshal(notificationoutboxapp.InternalRecipient{
		UserIDs: []uint{record.RecipientUserID},
	})
	if err != nil {
		return notificationmodel.Outbox{}, err
	}
	payloadJSON, err := json.Marshal(notificationoutboxapp.MessagePayload{
		Title:            record.Title,
		Content:          record.Content,
		SourceType:       record.SourceType,
		SourceID:         record.SourceID,
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

func isMySQLDuplicateIndex(err error, expectedIndex string) bool {
	index, ok := mysqlDuplicateIndexName(err)
	return ok && index == strings.ToLower(expectedIndex)
}

func mysqlDuplicateIndexName(err error) (string, bool) {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != mysqlDuplicateEntryNumber {
		return "", false
	}
	match := mysqlDuplicateKeySuffixPattern.FindStringSubmatch(mysqlErr.Message)
	if len(match) != 2 {
		return "", false
	}
	qualifiedName := strings.Trim(strings.TrimSpace(match[1]), "'`\"")
	parts := strings.Split(qualifiedName, ".")
	index := strings.ToLower(strings.Trim(strings.TrimSpace(parts[len(parts)-1]), "'`\""))
	if !mysqlIndexNamePattern.MatchString(index) {
		return "", false
	}
	return index, true
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
