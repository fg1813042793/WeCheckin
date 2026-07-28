package event

import (
	"context"
	"encoding/json"
	"strconv"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/app/support/access"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type EventDynamicListResponse struct {
	List  []model.EventDynamic `json:"list"`
	Total int64                `json:"total"`
}

func PostEventDynamic(eventID, userID, title, content, images, videos, addIP string) error {
	return PostEventDynamicContext(context.Background(), eventID, userID, title, content, images, videos, addIP)
}

func PostEventDynamicContext(ctx context.Context, eventID, userID, title, content, images, videos, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	dyn := model.EventDynamic{
		EventID: uint(parseUint(eventID)),
		UserID:  userID,
		Title:   title,
		Content: content,
		Images:  images,
		Videos:  videos,
		AddTime: database.Now(),
		AddIP:   addIP,
	}
	return db.Create(&dyn).Error
}

func PostEventDynamicForAdminContext(ctx context.Context, eventID, userID, title, content, images, videos, addIP string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEventVisibleContext(ctx, db, eventID, adminID); err != nil {
		return err
	}
	dyn := model.EventDynamic{
		EventID: uint(parseUint(eventID)),
		UserID:  userID,
		Title:   title,
		Content: content,
		Images:  images,
		Videos:  videos,
		AddTime: database.Now(),
		AddIP:   addIP,
	}
	return db.Create(&dyn).Error
}

func GetEventDynamics(eventID string, page, pageSize int) (EventDynamicListResponse, error) {
	return GetEventDynamicsContext(context.Background(), eventID, page, pageSize)
}

func GetEventDynamicsContext(ctx context.Context, eventID string, page, pageSize int) (EventDynamicListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EventDynamic
	var total int64
	query := db.Model(&model.EventDynamic{}).Where("`event_dynamic_event_id` = ?", eventID)
	if err := query.Count(&total).Error; err != nil {
		return EventDynamicListResponse{}, err
	}
	err := query.Order("`event_dynamic_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return EventDynamicListResponse{}, err
	}
	userIDs := make([]string, 0, len(list))
	for _, item := range list {
		userIDs = append(userIDs, item.UserID)
	}
	userInfoByOpenID, _ := loadEventUserInfoByOpenIDContext(ctx, db, userIDs)
	for i := range list {
		userInfo := userInfoByOpenID[list[i].UserID]
		list[i].UserName = userInfo.User.Name
		list[i].UserAvatar = userInfo.Avatar
		if list[i].Images != "" {
			json.Unmarshal([]byte(list[i].Images), &list[i].ImageList)
			for j := range list[i].ImageList {
				list[i].ImageList[j] = media.FullURLWithStaticDomain(list[i].ImageList[j])
			}
		}
		if list[i].Videos != "" {
			json.Unmarshal([]byte(list[i].Videos), &list[i].VideoList)
			for j := range list[i].VideoList {
				list[i].VideoList[j] = media.FullURLWithStaticDomain(list[i].VideoList[j])
			}
		}
	}
	return EventDynamicListResponse{List: list, Total: total}, nil
}

func GetEventDynamicsForAdminContext(ctx context.Context, eventID string, page, pageSize int, adminID uint) (EventDynamicListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEventVisibleContext(ctx, db, eventID, adminID); err != nil {
		return EventDynamicListResponse{}, err
	}
	return GetEventDynamicsContext(ctx, eventID, page, pageSize)
}

func EditEventDynamic(id, title, content, images, videos, editIP string) error {
	return EditEventDynamicContext(context.Background(), id, title, content, images, videos, editIP)
}

func EditEventDynamicContext(ctx context.Context, id, title, content, images, videos, editIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.EventDynamic{}).
		Where("`event_dynamic_id` = ?", id).
		Updates(map[string]interface{}{
			"event_dynamic_title":     title,
			"event_dynamic_content":   content,
			"event_dynamic_images":    images,
			"event_dynamic_videos":    videos,
			"event_dynamic_edit_time": database.Now(),
			"event_dynamic_edit_ip":   editIP,
		}).Error
}

func EditEventDynamicForAdminContext(ctx context.Context, id, title, content, images, videos, editIP string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var dyn model.EventDynamic
		if err := tx.Where("`event_dynamic_id` = ?", id).First(&dyn).Error; err != nil {
			return err
		}
		if err := ensureEventVisibleContext(ctx, tx, strconv.Itoa(int(dyn.EventID)), adminID); err != nil {
			return err
		}
		return access.RequireRowsAffected(tx.Model(&model.EventDynamic{}).
			Where("`event_dynamic_id` = ?", id).
			Updates(map[string]interface{}{
				"event_dynamic_title":     title,
				"event_dynamic_content":   content,
				"event_dynamic_images":    images,
				"event_dynamic_videos":    videos,
				"event_dynamic_edit_time": database.Now(),
				"event_dynamic_edit_ip":   editIP,
			}))
	})
}

func DelEventDynamic(id string) error {
	return DelEventDynamicContext(context.Background(), id)
}

func DelEventDynamicContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`event_dynamic_id` = ?", id).Delete(&model.EventDynamic{}).Error
}

func DelEventDynamicForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var dyn model.EventDynamic
		if err := tx.Where("`event_dynamic_id` = ?", id).First(&dyn).Error; err != nil {
			return err
		}
		if err := ensureEventVisibleContext(ctx, tx, strconv.Itoa(int(dyn.EventID)), adminID); err != nil {
			return err
		}
		return access.RequireRowsAffected(tx.Where("`event_dynamic_id` = ?", id).Delete(&model.EventDynamic{}))
	})
}

func DelEventDynamics(ids []string) error {
	return DelEventDynamicsContext(context.Background(), ids)
}

func DelEventDynamicsContext(ctx context.Context, ids []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`event_dynamic_id` IN ?", ids).Delete(&model.EventDynamic{}).Error
}

func DelEventDynamicsForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelEventDynamicForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}
