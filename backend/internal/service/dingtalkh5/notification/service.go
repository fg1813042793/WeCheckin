package notification

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

var (
	ErrUnauthenticated      = errors.New("h5app user is not authenticated")
	ErrNotificationNotFound = errors.New("notification not found")
)

type Repository interface {
	List(ctx context.Context, userID string, page, pageSize int) ([]model.Notify, int64, error)
	UnreadCount(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, userID string, id uint) (bool, error)
	MarkAllRead(ctx context.Context, userID string) error
}

type Notification struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	IsRead     int    `json:"isRead"`
	AddTime    int64  `json:"addTime"`
}

type ListResult struct {
	List     []Notification `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type Service struct {
	repository Repository
}

func NewService(db *gorm.DB) *Service {
	return NewServiceWithRepository(NewGormRepository(db))
}

func NewServiceWithRepository(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, currentUserID uint, page, pageSize int) (ListResult, error) {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return ListResult{}, err
	}
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := s.repository.List(ctx, userID, page, pageSize)
	if err != nil {
		return ListResult{}, err
	}
	result := ListResult{
		List:     make([]Notification, 0, len(items)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, item := range items {
		result.List = append(result.List, Notification{
			ID:         item.ID,
			Title:      item.Title,
			Content:    item.Content,
			Type:       item.Type,
			SourceType: item.SourceType,
			SourceID:   item.SourceID,
			IsRead:     item.IsRead,
			AddTime:    item.AddTime,
		})
	}
	return result, nil
}

func (s *Service) UnreadCount(ctx context.Context, currentUserID uint) (int64, error) {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return 0, err
	}
	return s.repository.UnreadCount(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, currentUserID, notificationID uint) error {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return err
	}
	if notificationID == 0 {
		return ErrNotificationNotFound
	}
	updated, err := s.repository.MarkRead(ctx, userID, notificationID)
	if err != nil {
		return err
	}
	if !updated {
		return ErrNotificationNotFound
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, currentUserID uint) error {
	userID, err := requiredUserID(currentUserID)
	if err != nil {
		return err
	}
	return s.repository.MarkAllRead(ctx, userID)
}

func requiredUserID(currentUserID uint) (string, error) {
	if currentUserID == 0 {
		return "", ErrUnauthenticated
	}
	return strconv.FormatUint(uint64(currentUserID), 10), nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) List(ctx context.Context, userID string, page, pageSize int) ([]model.Notify, int64, error) {
	db, cancel := r.withContext(ctx)
	defer cancel()
	query := db.Model(&model.Notify{}).Where("`notify_user_id` = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.Notify, 0)
	err := query.Order("`notify_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *GormRepository) UnreadCount(ctx context.Context, userID string) (int64, error) {
	db, cancel := r.withContext(ctx)
	defer cancel()
	var count int64
	err := db.Model(&model.Notify{}).
		Where("`notify_user_id` = ? AND `notify_is_read` = 0", userID).
		Count(&count).Error
	return count, err
}

func (r *GormRepository) MarkRead(ctx context.Context, userID string, id uint) (bool, error) {
	db, cancel := r.withContext(ctx)
	defer cancel()
	var item model.Notify
	err := db.
		Select("notify_id").
		Where("`notify_id` = ? AND `notify_user_id` = ?", id, userID).
		Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = db.Model(&model.Notify{}).
		Where("`notify_id` = ? AND `notify_user_id` = ?", id, userID).
		UpdateColumn("notify_is_read", 1).Error
	return err == nil, err
}

func (r *GormRepository) MarkAllRead(ctx context.Context, userID string) error {
	db, cancel := r.withContext(ctx)
	defer cancel()
	return db.Model(&model.Notify{}).
		Where("`notify_user_id` = ? AND `notify_is_read` = 0", userID).
		UpdateColumn("notify_is_read", 1).Error
}

func (r *GormRepository) withContext(parent context.Context) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := database.QueryContext(parent)
	return r.db.WithContext(ctx), cancel
}
