package event

import (
	"context"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type EventListResponse struct {
	List  []model.Event `json:"list"`
	Total int64         `json:"total"`
}

type EventRolesResponse struct {
	HasOrganizer bool   `json:"hasOrganizer"`
	HasAssistant bool   `json:"hasAssistant"`
	HasReferee   bool   `json:"hasReferee"`
	OrganizerIDs []uint `json:"organizerIDs"`
	AssistantIDs []uint `json:"assistantIDs"`
	RefereeIDs   []uint `json:"refereeIDs"`
}

func GetMyEventList(userID, typ, status string, page, pageSize int) (EventListResponse, error) {
	return GetMyEventListContext(context.Background(), userID, typ, status, page, pageSize)
}

func GetMyEventListContext(ctx context.Context, userID, typ, status string, page, pageSize int) (EventListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var partIDs []uint
	if err := db.Model(&model.EventParticipant{}).
		Where("`event_part_mini_openid` = ?", userID).
		Pluck("`event_part_event_id`", &partIDs).Error; err != nil {
		return EventListResponse{}, err
	}
	if len(partIDs) == 0 {
		return EventListResponse{List: []model.Event{}, Total: 0}, nil
	}
	query := db.Model(&model.Event{}).Where("`id` IN ?", partIDs)
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	if status != "" {
		query = query.Where("`event_status` = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return EventListResponse{}, err
	}
	var list []model.Event
	if err := query.Order("`add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return EventListResponse{}, err
	}
	list = populateEventFields(list)
	for i := range list {
		list[i].IsJoin = true
	}
	return EventListResponse{List: list, Total: total}, nil
}

func GetMyEventRoles(userID string) (EventRolesResponse, error) {
	return GetMyEventRolesContext(context.Background(), userID)
}

func GetMyEventRolesContext(ctx context.Context, userID string) (EventRolesResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var roles []model.EventRole
	if err := db.Where("`event_role_user_id` = ?", userID).Find(&roles).Error; err != nil {
		return EventRolesResponse{}, err
	}

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

	return EventRolesResponse{
		HasOrganizer: len(orgIDs) > 0,
		HasAssistant: len(astIDs) > 0,
		HasReferee:   len(refIDs) > 0,
		OrganizerIDs: orgIDs,
		AssistantIDs: astIDs,
		RefereeIDs:   refIDs,
	}, nil
}

func GetMyManagedList(userID, typ, status, keyword string, page, pageSize int) (EventListResponse, error) {
	return GetMyManagedListContext(context.Background(), userID, typ, status, keyword, page, pageSize)
}

func GetMyManagedListContext(ctx context.Context, userID, typ, status, keyword string, page, pageSize int) (EventListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var eventIDs []uint
	if err := db.Model(&model.EventRole{}).
		Where("`event_role_user_id` = ?", userID).
		Pluck("`event_role_event_id`", &eventIDs).Error; err != nil {
		return EventListResponse{}, err
	}
	if len(eventIDs) == 0 {
		return EventListResponse{List: []model.Event{}, Total: 0}, nil
	}
	query := db.Model(&model.Event{}).Where("`id` IN ?", eventIDs)
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
	if err := query.Count(&total).Error; err != nil {
		return EventListResponse{}, err
	}
	var list []model.Event
	if err := query.Order("`add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return EventListResponse{}, err
	}
	list = populateEventFields(list)

	loadEventRolesForListWithDB(ctx, db, list, userID)

	return EventListResponse{List: list, Total: total}, nil
}

func loadEventRolesForListWithDB(ctx context.Context, db *gorm.DB, list []model.Event, userID string) {
	if len(list) == 0 || userID == "" {
		return
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return
		}
	}

	eventIDs := make([]uint, 0, len(list))
	for _, item := range list {
		eventIDs = append(eventIDs, item.ID)
	}
	var roles []model.EventRole
	if err := db.Where("`event_role_event_id` IN ? AND `event_role_user_id` = ?", eventIDs, userID).Find(&roles).Error; err != nil {
		return
	}
	roleByEventID := make(map[uint]string, len(roles))
	for _, role := range roles {
		roleByEventID[role.EventID] = role.Role
	}

	for i := range list {
		roleName := roleByEventID[list[i].ID]
		if roleName == "" {
			continue
		}
		switch roleName {
		case "organizer":
			list[i].RoleName = "工作人员:主办人"
		case "assistant":
			list[i].RoleName = "工作人员:主办人助理"
		case "referee":
			list[i].RoleName = "工作人员:裁判"
		}
	}
}
