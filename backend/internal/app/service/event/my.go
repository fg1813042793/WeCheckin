package event

import (
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetMyEventList(userID, typ, status string, page, pageSize int) (map[string]interface{}, error) {
	var partIDs []uint
	database.DB.Model(&model.EventParticipant{}).
		Where("`event_part_mini_openid` = ?", userID).
		Pluck("`event_part_event_id`", &partIDs)
	if len(partIDs) == 0 {
		return map[string]interface{}{"list": []model.Event{}, "total": 0}, nil
	}
	query := database.DB.Model(&model.Event{}).Where("`id` IN ?", partIDs)
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	if status != "" {
		query = query.Where("`event_status` = ?", status)
	}
	var total int64
	query.Count(&total)
	var list []model.Event
	query.Order("`event_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	list = populateEventFields(list)
	for i := range list {
		list[i].IsJoin = true
	}
	return map[string]interface{}{"list": list, "total": total}, nil
}

func GetMyEventRoles(userID string) (map[string]interface{}, error) {
	var roles []model.EventRole
	database.DB.Where("`event_role_user_id` = ?", userID).Find(&roles)

	orgIDs := []uint{}
	astIDs := []uint{}
	refIDs := []uint{}
	for _, r := range roles {
		switch r.Role {
		case "organizer":
			orgIDs = append(orgIDs, r.EventID)
		case "assistant":
			astIDs = append(astIDs, r.EventID)
		case "referee":
			refIDs = append(refIDs, r.EventID)
		}
	}

	result := map[string]interface{}{
		"hasOrganizer": len(orgIDs) > 0,
		"hasAssistant": len(astIDs) > 0,
		"hasReferee":   len(refIDs) > 0,
		"organizerIDs": orgIDs,
		"assistantIDs": astIDs,
		"refereeIDs":   refIDs,
	}
	return result, nil
}

func GetMyManagedList(userID, typ, status, keyword string, page, pageSize int) (map[string]interface{}, error) {
	var eventIDs []uint
	database.DB.Model(&model.EventRole{}).
		Where("`event_role_user_id` = ?", userID).
		Pluck("`event_role_event_id`", &eventIDs)
	if len(eventIDs) == 0 {
		return map[string]interface{}{"list": []model.Event{}, "total": 0}, nil
	}
	query := database.DB.Model(&model.Event{}).Where("`id` IN ?", eventIDs)
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	if status != "" {
		query = query.Where("`event_status` = ?", status)
	}
	if keyword != "" {
		query = query.Where("`event_title` LIKE ?", "%"+keyword+"%")
	}
	var total int64
	query.Count(&total)
	var list []model.Event
	query.Order("`event_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	list = populateEventFields(list)

	// Attach role name
	loadEventRolesForList(list, userID)

	return map[string]interface{}{"list": list, "total": total}, nil
}
