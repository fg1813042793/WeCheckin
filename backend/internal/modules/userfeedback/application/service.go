package application

import (
	"context"
	"log"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"wecheckin/backend/internal/modules/userfeedback/domain"
	projectlogger "wecheckin/backend/pkg/logger"
)

const (
	MaxFeedbacksPerUserPerDay = 20
	MaxRequestIDRunes         = 64
	MaxKeywordRunes           = 100
	DefaultPageSize           = 20
	MaxPageSize               = 100
	MaxStatusNoteRunes        = 5000
)

type ReplayReader interface {
	FindCreateReplay(ctx context.Context, key CreateReplayKey) (*FeedbackDetail, bool, error)
	FindMessageReplay(ctx context.Context, key MessageReplayKey) (*FeedbackDetail, bool, error)
}

type Store interface {
	ReplayReader
	InTransaction(ctx context.Context, fn func(TransactionStore) error) error
	GetUserOverview(ctx context.Context, userID uint) (Overview, error)
	ListUserFeedbacks(ctx context.Context, userID uint, query UserListQuery) (FeedbackList, error)
	GetUserFeedback(ctx context.Context, id uint64, userID uint) (*FeedbackDetail, error)
	GetAdminOverview(ctx context.Context, query AdminListQuery) (Overview, error)
	ListAdminFeedbacks(ctx context.Context, query AdminListQuery) (FeedbackList, error)
	GetAdminFeedback(ctx context.Context, id uint64) (*FeedbackDetail, error)
}

type TransactionStore interface {
	ReplayReader
	// NextFeedbackNumber serializes creation for one date key. The following
	// count must observe rows committed by the previous holder of that lock.
	NextFeedbackNumber(ctx context.Context, dateKey string) (uint64, error)
	CountCreatedByUserOnDate(ctx context.Context, userID uint, startMs, endMs int64) (int64, error)
	CreateFeedback(ctx context.Context, record FeedbackRecord) (uint64, error)
	LockFeedback(ctx context.Context, id uint64) (*FeedbackSnapshot, error)
	AppendMessage(ctx context.Context, record MessageRecord) (uint64, error)
	AppendAttachments(ctx context.Context, records []AttachmentRecord) error
	// UpdateSnapshot must return ErrVersionConflict when expectedVersion no longer matches.
	UpdateSnapshot(ctx context.Context, snapshot FeedbackSnapshot, expectedVersion uint64) error
	EnqueueNotification(ctx context.Context, record NotificationOutboxRecord) error
}

type CreateReplayKey struct {
	SubmitterID uint
	RequestID   string
}

type MessageReplayKey struct {
	FeedbackID uint64
	AuthorType domain.AuthorType
	AuthorID   uint
	RequestID  string
}

type FeedbackRecord struct {
	FeedbackNo      string
	SubmitterID     uint
	CreateRequestID string
	Status          domain.Status
	Version         uint64
	LastActivityAt  int64
	ResolvedAt      *int64
	ClosedAt        *int64
	CreatedAt       int64
	UpdatedAt       int64
}

type FeedbackSnapshot struct {
	ID              uint64
	FeedbackNo      string
	SubmitterID     uint
	Status          domain.Status
	HandlerID       *uint
	Version         uint64
	AttachmentCount int64
	LastActivityAt  int64
	ResolvedAt      *int64
	ClosedAt        *int64
	CreatedAt       int64
	UpdatedAt       int64
}

type MessageRecord struct {
	FeedbackID  uint64
	MessageType domain.MessageType
	AuthorType  domain.AuthorType
	AuthorID    uint
	Content     string
	FromStatus  domain.Status
	ToStatus    domain.Status
	RequestID   string
	CreatedAt   int64
}

type AttachmentRecord struct {
	FeedbackID      uint64
	MessageID       uint64
	StorageProvider string
	ObjectKey       string
	OriginalName    string
	ContentType     string
	SizeBytes       uint64
	SortOrder       int
	CreatedAt       int64
}

type NotificationOutboxRecord struct {
	IdempotencyKey   string
	Channel          string
	NotificationType string
	SourceType       string
	SourceID         string
	RecipientUserID  uint
	Title            string
	Content          string
	CreatedAt        int64
}

type Logger interface {
	Printf(format string, values ...interface{})
}

type Service struct {
	store        Store
	imageStorage ImageStorage
	now          func() time.Time
	location     *time.Location
	logger       Logger
}

func NewService(store Store, imageStorage ImageStorage) *Service {
	return newServiceWithClock(store, imageStorage, time.Now, shanghaiLocation(time.LoadLocation))
}

func shanghaiLocation(loader func(string) (*time.Location, error)) *time.Location {
	location, err := loader("Asia/Shanghai")
	if err != nil || location == nil {
		location = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return location
}

func newServiceWithClock(store Store, imageStorage ImageStorage, now func() time.Time, location *time.Location) *Service {
	return newServiceWithClockAndLogger(store, imageStorage, now, location, defaultLogger())
}

func newServiceWithClockAndLogger(store Store, imageStorage ImageStorage, now func() time.Time, location *time.Location, logger Logger) *Service {
	return &Service{store: store, imageStorage: imageStorage, now: now, location: location, logger: logger}
}

func defaultLogger() Logger {
	if projectlogger.Logger != nil {
		return projectlogger.Logger
	}
	return log.Default()
}

func (service *Service) requireStore() error {
	if service == nil || service.store == nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (service *Service) requireWritable() error {
	if err := service.requireStore(); err != nil {
		return err
	}
	if service.imageStorage == nil || service.now == nil || service.location == nil {
		return ErrServiceUnavailable
	}
	return nil
}

func (service *Service) requireClock() error {
	if err := service.requireStore(); err != nil {
		return err
	}
	if service.now == nil {
		return ErrServiceUnavailable
	}
	return nil
}

func normalizeRequestID(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", ErrInvalidArgument
	}
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > MaxRequestIDRunes {
		return "", ErrInvalidArgument
	}
	return value, nil
}

func normalizeKeyword(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", ErrInvalidArgument
	}
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > MaxKeywordRunes {
		return "", ErrInvalidArgument
	}
	return value, nil
}

func normalizePage(page, pageSize int) (int, int, error) {
	if page < 0 || pageSize < 0 || pageSize > MaxPageSize {
		return 0, 0, ErrInvalidArgument
	}
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	return page, pageSize, nil
}

func decorateDetail(detail *FeedbackDetail) *FeedbackDetail {
	if detail != nil {
		detail.AllowsSupplement = detail.Status.AllowsSupplement()
	}
	return detail
}
