package survey

import (
	"context"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type NotificationQuery struct {
	SourceType string
	SourceID   string
	UserID     string
	Page       int
	PageSize   int
}

type NotificationListResult struct {
	List  []model.Notify `json:"list"`
	Total int64          `json:"total"`
}

type NotificationReadInput struct {
	ID     uint
	All    bool
	UserID string
}

func (s *SurveyService) NotificationListContext(ctx context.Context, query NotificationQuery) (NotificationListResult, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Notify{})
	if query.SourceType != "" {
		q = q.Where("`notify_source_type` = ?", query.SourceType)
	}
	if query.SourceID != "" {
		q = q.Where("`notify_source_id` = ?", query.SourceID)
	}
	if query.UserID != "" {
		q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", query.UserID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return NotificationListResult{}, err
	}
	var list []model.Notify
	err := q.Order("`notify_id` DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&list).Error
	return NotificationListResult{List: list, Total: total}, err
}

func (s *SurveyService) MarkNotificationsReadContext(ctx context.Context, input NotificationReadInput) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if input.All {
		q := db.Model(&model.Notify{}).Where("`notify_is_read` = 0")
		if input.UserID != "" {
			q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", input.UserID)
		}
		return q.UpdateColumn("notify_is_read", 1).Error
	}
	if input.ID > 0 {
		return db.Model(&model.Notify{}).Where("`notify_id` = ?", input.ID).UpdateColumn("notify_is_read", 1).Error
	}
	return nil
}

func (s *SurveyService) NotificationUnreadCountContext(ctx context.Context, userID string) (int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Notify{}).Where("`notify_is_read` = 0")
	if userID != "" {
		q = q.Where("`notify_user_id` = ? OR `notify_user_id` = ''", userID)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}
