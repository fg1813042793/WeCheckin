package event

import (
	"encoding/json"

	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func PostEventDynamic(eventID, userID, title, content, images, videos, addIP string) error {
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
	return database.DB.Create(&dyn).Error
}

func GetEventDynamics(eventID string, page, pageSize int) (map[string]interface{}, error) {
	var list []model.EventDynamic
	var total int64
	query := database.DB.Model(&model.EventDynamic{}).Where("`event_dynamic_event_id` = ?", eventID)
	query.Count(&total)
	err := query.Order("`event_dynamic_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	// Populate user info
	for i := range list {
		var user model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].UserID).First(&user)
		list[i].UserName = user.Name
		list[i].UserAvatar = media.FullURLWithStaticDomain(user.Pic)
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
	return map[string]interface{}{"list": list, "total": total}, nil
}

func EditEventDynamic(id, title, content, images, videos, editIP string) error {
	return database.DB.Model(&model.EventDynamic{}).
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

func DelEventDynamic(id string) error {
	return database.DB.Where("`event_dynamic_id` = ?", id).Delete(&model.EventDynamic{}).Error
}

func DelEventDynamics(ids []string) error {
	return database.DB.Where("`event_dynamic_id` IN ?", ids).Delete(&model.EventDynamic{}).Error
}
