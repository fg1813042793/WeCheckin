package event

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/query"
	"wecheckin/backend/pkg/database"
)

// ==================== Admin ====================

var adminEventListColumns = []string{
	"id",
	"event_title",
	"event_type",
	"event_status",
	"create_dept_id",
	"event_publish_dept_ids",
	"create_by",
	"update_by",
	"update_dept_id",
	"event_cate_id",
	"event_cate_name",
	"event_reg_start",
	"event_reg_end",
	"event_event_start",
	"event_event_end",
	"event_order",
	"event_vouch",
	"event_is_top",
	"event_obj",
	"event_qr",
	"event_view_cnt",
	"event_join_cnt",
	"event_user_cnt",
	"add_time",
	"edit_time",
}

func GetAdminEventList(keyword, typ, sortStr string, page, pageSize int, adminID uint) ([]model.Event, int64, error) {
	return GetAdminEventListContext(context.Background(), keyword, typ, sortStr, page, pageSize, adminID)
}

func GetAdminEventListContext(ctx context.Context, keyword, typ, sortStr string, page, pageSize int, adminID uint) ([]model.Event, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admin model.Admin
	db.First(&admin, adminID)
	var list []model.Event
	var total int64
	queryBuilder := db.Model(&model.Event{})
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`event_title` LIKE ?", "%"+keyword+"%")
	}
	if typ != "" {
		queryBuilder = queryBuilder.Where("`event_type` = ?", typ)
	}
	where, args := access.DataScopeFilterForResourceWithDBContext(ctx, db, &admin, access.EventAuditFields)
	if where != "" {
		queryBuilder = queryBuilder.Where(where, args...)
	}
	queryBuilder.Count(&total)
	orderClause := query.ParseSort(sortStr, map[string]string{
		"title":      "event_title",
		"type":       "event_type",
		"status":     "event_status",
		"order":      "event_order",
		"userCnt":    "event_user_cnt",
		"regStart":   "event_reg_start",
		"regEnd":     "event_reg_end",
		"eventStart": "event_event_start",
		"eventEnd":   "event_event_end",
		"addTime":    "add_time",
	})
	if orderClause != "" {
		queryBuilder = queryBuilder.Order(orderClause)
	} else {
		queryBuilder = queryBuilder.Order("`add_time` DESC")
	}
	err := queryBuilder.Select(adminEventListColumns).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEventFields(list)
	return list, total, nil
}

func scopedEventQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Event{}, access.EventAuditFields)
}

func ensureEventVisibleContext(ctx context.Context, db *gorm.DB, eventID string, adminID uint) error {
	queryBuilder, err := scopedEventQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return queryBuilder.Where("`id` = ?", eventID).First(&model.Event{}).Error
}

func EnsureEventVisibleForAdminContext(ctx context.Context, eventID string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return ensureEventVisibleContext(ctx, db, eventID, adminID)
}

func GetAdminEventDetail(id string) (*model.Event, error) {
	return GetAdminEventDetailContext(context.Background(), id)
}

func GetAdminEventDetailContext(ctx context.Context, id string) (*model.Event, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var event model.Event
	err := db.Where("`id` = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	var obj eventObj
	if event.Obj != "" {
		json.Unmarshal([]byte(event.Obj), &obj)
	}
	if len(obj.Cover) > 0 {
		event.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	event.Desc = obj.Desc
	event.Rules = obj.Rules
	populateEventTimeFields(&event)
	loadEventRolesContext(ctx, &event)

	// Count participants
	var pCnt int64
	db.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", id).Count(&pCnt)
	event.UserCnt = int(pCnt)

	return &event, nil
}

func GetAdminEventDetailForAdminContext(ctx context.Context, id string, adminID uint) (*model.Event, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEventQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var event model.Event
	if err := queryBuilder.Where("`id` = ?", id).First(&event).Error; err != nil {
		return nil, err
	}
	var obj eventObj
	if event.Obj != "" {
		json.Unmarshal([]byte(event.Obj), &obj)
	}
	if len(obj.Cover) > 0 {
		event.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	event.Desc = obj.Desc
	event.Rules = obj.Rules
	populateEventTimeFields(&event)
	loadEventRolesContext(ctx, &event)

	var pCnt int64
	db.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", id).Count(&pCnt)
	event.UserCnt = int(pCnt)

	return &event, nil
}

func InsertEvent(title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID, createBy uint, organizers, assistants, referees []string) error {
	return InsertEventContext(context.Background(), title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds, typ, status, order, regStart, regEnd, eventStart, eventEnd, obj, deptID, createBy, organizers, assistants, referees)
}

func InsertEventContext(ctx context.Context, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID, createBy uint, organizers, assistants, referees []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	event := model.Event{
		Title:          title,
		Type:           typ,
		Status:         status,
		CateID:         cateID,
		CateName:       cateName,
		RegStart:       regStart,
		RegEnd:         regEnd,
		EventStart:     eventStart,
		EventEnd:       eventEnd,
		Order:          order,
		Forms:          forms,
		ScoreFields:    scoreFields,
		QR:             qr,
		Obj:            obj,
		DeptID:         deptID,
		PublishDeptIds: publishDeptIds,
		CreateBy:       createBy,
		AddTime:        database.Now(),
		EditTime:       database.Now(),
		UpdateBy:       createBy,
		UpdateDeptID:   deptID,
		AddIP:          addIP,
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return SaveEventRolesTx(tx, event.ID, organizers, assistants, referees)
	})
}

func EditEvent(id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID uint, organizers, assistants, referees []string) error {
	return EditEventContext(context.Background(), id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds, typ, status, order, regStart, regEnd, eventStart, eventEnd, obj, deptID, organizers, assistants, referees)
}

func EditEventContext(ctx context.Context, id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID uint, organizers, assistants, referees []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"event_title":            title,
		"event_type":             typ,
		"event_status":           status,
		"event_cate_id":          cateID,
		"event_cate_name":        cateName,
		"event_reg_start":        regStart,
		"event_reg_end":          regEnd,
		"event_event_start":      eventStart,
		"event_event_end":        eventEnd,
		"event_order":            order,
		"event_forms":            forms,
		"event_score_fields":     scoreFields,
		"create_dept_id":         deptID,
		"update_dept_id":         deptID,
		"event_publish_dept_ids": publishDeptIds,
		"event_qr":               qr,
		"edit_time":              database.Now(),
		"event_edit_ip":          addIP,
	}
	if obj != "" {
		updates["event_obj"] = obj
	}
	eid := parseUint(id)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Event{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return SaveEventRolesTx(tx, uint(eid), organizers, assistants, referees)
	})
}

func EditEventForAdminContext(ctx context.Context, id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID, adminID uint, organizers, assistants, referees []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	updates := map[string]interface{}{
		"event_title":            title,
		"event_type":             typ,
		"event_status":           status,
		"event_cate_id":          cateID,
		"event_cate_name":        cateName,
		"event_reg_start":        regStart,
		"event_reg_end":          regEnd,
		"event_event_start":      eventStart,
		"event_event_end":        eventEnd,
		"event_order":            order,
		"event_forms":            forms,
		"event_score_fields":     scoreFields,
		"create_dept_id":         deptID,
		"update_by":              adminID,
		"update_dept_id":         deptID,
		"event_publish_dept_ids": publishDeptIds,
		"event_qr":               qr,
		"edit_time":              database.Now(),
		"event_edit_ip":          addIP,
	}
	if obj != "" {
		updates["event_obj"] = obj
	}
	eid := parseUint(id)
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedEventQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Updates(updates)); err != nil {
			return err
		}
		return SaveEventRolesTx(tx, uint(eid), organizers, assistants, referees)
	})
}

func DelEvent(id string) error {
	return DelEventContext(context.Background(), id)
}

func DelEventContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	eid := parseUint(id)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`event_role_event_id` = ?", eid).Delete(&model.EventRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_part_event_id` = ?", eid).Delete(&model.EventParticipant{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_dynamic_event_id` = ?", eid).Delete(&model.EventDynamic{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_score_event_id` = ?", eid).Delete(&model.EventScore{}).Error; err != nil {
			return err
		}
		return tx.Where("`id` = ?", id).Delete(&model.Event{}).Error
	})
}

func DelEventForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedEventQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		var event model.Event
		if err := queryBuilder.Where("`id` = ?", id).First(&event).Error; err != nil {
			return err
		}
		eid := event.ID
		if err := tx.Where("`event_role_event_id` = ?", eid).Delete(&model.EventRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_part_event_id` = ?", eid).Delete(&model.EventParticipant{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_dynamic_event_id` = ?", eid).Delete(&model.EventDynamic{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`event_score_event_id` = ?", eid).Delete(&model.EventScore{}).Error; err != nil {
			return err
		}
		return access.RequireRowsAffected(tx.Where("`id` = ?", eid).Delete(&model.Event{}))
	})
}

func DelEvents(ids []string) error {
	return DelEventsContext(context.Background(), ids)
}

func DelEventsContext(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := DelEventContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func DelEventsForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelEventForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func StatusEvent(id string, status int) error {
	return StatusEventContext(context.Background(), id, status)
}

func StatusEventContext(ctx context.Context, id string, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Event{}).Where("`id` = ?", id).
		Update("event_status", status).Error
}

func StatusEventForAdminContext(ctx context.Context, id string, status int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEventQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("event_status", status))
}

func GetEventParticipantList(eventID string) ([]model.EventParticipant, error) {
	return GetEventParticipantListContext(context.Background(), eventID)
}

func GetEventParticipantListContext(ctx context.Context, eventID string) ([]model.EventParticipant, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EventParticipant
	err := db.Where("`event_part_event_id` = ?", eventID).
		Order("`event_part_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return enrichEventParticipantsWithUserInfoContext(ctx, db, list), nil
}

func GetEventParticipantListForAdminContext(ctx context.Context, eventID string, adminID uint) ([]model.EventParticipant, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEventVisibleContext(ctx, db, eventID, adminID); err != nil {
		return nil, err
	}
	return GetEventParticipantListContext(ctx, eventID)
}

func DelEventParticipant(id string) error {
	return DelEventParticipantContext(context.Background(), id)
}

func DelEventParticipantContext(ctx context.Context, id string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var participant model.EventParticipant
		if err := tx.Where("`id` = ?", id).First(&participant).Error; err != nil {
			return err
		}
		if err := tx.Where("`id` = ?", id).Delete(&model.EventParticipant{}).Error; err != nil {
			return err
		}
		var userCnt int64
		if err := tx.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", participant.EventID).Count(&userCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Event{}).Where("`id` = ?", participant.EventID).Update("event_user_cnt", userCnt).Error
	})
}

func DelEventParticipantForAdminContext(ctx context.Context, id string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var participant model.EventParticipant
		if err := tx.Where("`id` = ?", id).First(&participant).Error; err != nil {
			return err
		}
		if err := ensureEventVisibleContext(ctx, tx, strconv.Itoa(int(participant.EventID)), adminID); err != nil {
			return err
		}
		if err := tx.Where("`id` = ?", id).Delete(&model.EventParticipant{}).Error; err != nil {
			return err
		}
		var userCnt int64
		if err := tx.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", participant.EventID).Count(&userCnt).Error; err != nil {
			return err
		}
		return tx.Model(&model.Event{}).Where("`id` = ?", participant.EventID).Update("event_user_cnt", userCnt).Error
	})
}

func EditEventParticipant(id, forms string) error {
	return EditEventParticipantContext(context.Background(), id, forms)
}

func EditEventParticipantContext(ctx context.Context, id, forms string) error {
	updates := map[string]interface{}{}
	if forms != "" {
		updates["event_part_forms"] = forms
		updates["event_part_edit_time"] = time.Now().UnixMilli()
	}
	if len(updates) == 0 {
		return nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.EventParticipant{}).Where("`id` = ?", id).Updates(updates).Error
}

func EditEventParticipantForAdminContext(ctx context.Context, id, forms string, adminID uint) error {
	updates := map[string]interface{}{}
	if forms != "" {
		updates["event_part_forms"] = forms
		updates["event_part_edit_time"] = time.Now().UnixMilli()
	}
	if len(updates) == 0 {
		return nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var participant model.EventParticipant
		if err := tx.Where("`id` = ?", id).First(&participant).Error; err != nil {
			return err
		}
		if err := ensureEventVisibleContext(ctx, tx, strconv.Itoa(int(participant.EventID)), adminID); err != nil {
			return err
		}
		return access.RequireRowsAffected(tx.Model(&model.EventParticipant{}).Where("`id` = ?", id).Updates(updates))
	})
}

func DelEventParticipants(ids []string) error {
	return DelEventParticipantsContext(context.Background(), ids)
}

func DelEventParticipantsContext(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := DelEventParticipantContext(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func DelEventParticipantsForAdminContext(ctx context.Context, ids []string, adminID uint) error {
	for _, id := range ids {
		if err := DelEventParticipantForAdminContext(ctx, id, adminID); err != nil {
			return err
		}
	}
	return nil
}

func AdminEditEventScore(id, score string) error {
	return AdminEditEventScoreContext(context.Background(), id, score)
}

func AdminEditEventScoreContext(ctx context.Context, id, score string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.EventScore{}).
		Where("`event_score_id` = ?", id).
		Update("event_score_score", score).Error
}

func VouchEvent(id string, vouch int) error {
	return VouchEventContext(context.Background(), id, vouch)
}

func VouchEventContext(ctx context.Context, id string, vouch int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Event{}).Where("`id` = ?", id).Update("event_vouch", vouch).Error
}

func VouchEventForAdminContext(ctx context.Context, id string, vouch int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEventQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("event_vouch", vouch))
}

func TopEvent(id string, top int) error {
	return TopEventContext(context.Background(), id, top)
}

func TopEventContext(ctx context.Context, id string, top int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Event{}).Where("`id` = ?", id).Update("event_is_top", top).Error
}

func TopEventForAdminContext(ctx context.Context, id string, top int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedEventQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`id` = ?", id).Update("event_is_top", top))
}
