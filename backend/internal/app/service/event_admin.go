package service

import (
	"encoding/json"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ==================== Admin ====================

func GetAdminEventList(keyword, typ, sortStr string, page, pageSize int, adminID uint) ([]model.Event, int64, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)
	var list []model.Event
	var total int64
	query := database.DB.Model(&model.Event{})
	if keyword != "" {
		query = query.Where("`event_title` LIKE ?", "%"+keyword+"%")
	}
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	where, args := BuildDataScopeFilter(&admin, "`event_dept_id`", "`event_create_by`")
	if where != "" {
		query = query.Where(where, args...)
	}
	query.Count(&total)
	orderClause := parseSort(sortStr, map[string]string{
		"title":      "event_title",
		"type":       "event_type",
		"status":     "event_status",
		"order":      "event_order",
		"userCnt":    "event_user_cnt",
		"regStart":   "event_reg_start",
		"regEnd":     "event_reg_end",
		"eventStart": "event_event_start",
		"eventEnd":   "event_event_end",
		"addTime":    "event_add_time",
	})
	if orderClause != "" {
		query = query.Order(orderClause)
	} else {
		query = query.Order("`event_add_time` DESC")
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	list = populateEventFields(list)
	return list, total, nil
}

func GetAdminEventDetail(id string) (*model.Event, error) {
	var event model.Event
	err := database.DB.Where("`id` = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	var obj eventObj
	if event.Obj != "" {
		json.Unmarshal([]byte(event.Obj), &obj)
	}
	if len(obj.Cover) > 0 {
		event.Img = GetFullURL(obj.Cover[0])
	}
	event.Desc = obj.Desc
	event.Rules = obj.Rules
	populateEventTimeFields(&event)
	loadEventRoles(&event)

	// Count participants
	var pCnt int64
	database.DB.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", id).Count(&pCnt)
	event.UserCnt = int(pCnt)

	return &event, nil
}

func InsertEvent(title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID, createBy uint, organizers, assistants, referees []string) error {
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
		AddIP:          addIP,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		return err
	}
	SaveEventRoles(event.ID, organizers, assistants, referees)
	return nil
}

func EditEvent(id, title, cateID, cateName, forms, scoreFields, qr, addIP, publishDeptIds string, typ, status, order int, regStart, regEnd, eventStart, eventEnd int64, obj string, deptID uint, organizers, assistants, referees []string) error {
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
		"event_dept_id":          deptID,
		"event_publish_dept_ids": publishDeptIds,
		"event_qr":               qr,
		"event_edit_time":        database.Now(),
		"event_edit_ip":          addIP,
	}
	if obj != "" {
		updates["event_obj"] = obj
	}
	if err := database.DB.Model(&model.Event{}).Where("`id` = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	eid := parseUint(id)
	SaveEventRoles(uint(eid), organizers, assistants, referees)
	return nil
}

func DelEvent(id string) error {
	eid := parseUint(id)
	tx := database.DB.Begin()
	tx.Where("`event_role_event_id` = ?", eid).Delete(&model.EventRole{})
	tx.Where("`event_part_event_id` = ?", eid).Delete(&model.EventParticipant{})
	tx.Where("`event_dynamic_event_id` = ?", eid).Delete(&model.EventDynamic{})
	tx.Where("`event_score_event_id` = ?", eid).Delete(&model.EventScore{})
	tx.Where("`id` = ?", id).Delete(&model.Event{})
	return tx.Commit().Error
}

func DelEvents(ids []string) error {
	for _, id := range ids {
		if err := DelEvent(id); err != nil {
			return err
		}
	}
	return nil
}

func StatusEvent(id string, status int) error {
	return database.DB.Model(&model.Event{}).Where("`id` = ?", id).
		Update("event_status", status).Error
}

func GetEventParticipantList(eventID string) ([]model.EventParticipant, error) {
	var list []model.EventParticipant
	err := database.DB.Where("`event_part_event_id` = ?", eventID).
		Order("`event_part_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var user model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].MiniOpenID).First(&user)
		list[i].UserName = user.Name
		list[i].UserAvatar = GetFullURL(user.Pic)
		list[i].Mobile = user.Mobile
		if user.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", user.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].DeptName = dept.Name
				list[i].TopDeptName = getTopDeptName(ud.DeptID)
			}
		}
	}
	return list, nil
}

func DelEventParticipant(id string) error {
	return database.DB.Where("`id` = ?", id).Delete(&model.EventParticipant{}).Error
}

func EditEventParticipant(id, forms string) error {
	updates := map[string]interface{}{}
	if forms != "" {
		updates["event_part_forms"] = forms
		updates["event_part_edit_time"] = time.Now().UnixMilli()
	}
	if len(updates) == 0 {
		return nil
	}
	return database.DB.Model(&model.EventParticipant{}).Where("`id` = ?", id).Updates(updates).Error
}

func DelEventParticipants(ids []string) error {
	for _, id := range ids {
		if err := DelEventParticipant(id); err != nil {
			return err
		}
	}
	return nil
}

func AdminEditEventScore(id, score string) error {
	return database.DB.Model(&model.EventScore{}).
		Where("`event_score_id` = ?", id).
		Update("event_score_score", score).Error
}
