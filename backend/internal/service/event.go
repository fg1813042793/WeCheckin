package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

type eventObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
	Rules string   `json:"rules"`
}

// ==================== Client ====================

func GetEventList(page, pageSize int, userID, keyword, typ string) (map[string]interface{}, error) {
	var list []model.Event
	var total int64
	query := database.DB.Model(&model.Event{}).Where("`event_status` = 1")
	if keyword != "" {
		query = query.Where("`event_title` LIKE ?", "%"+keyword+"%")
	}
	if typ != "" {
		query = query.Where("`event_type` = ?", typ)
	}
	if userID != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if len(deptIDs) > 0 {
			query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL OR " +
				buildDeptOverlap("event_publish_dept_ids", deptIDs) + ")")
		} else {
			query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)")
		}
	} else {
		query = query.Where("(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)")
	}
	query.Count(&total)
	err := query.Order("`event_order` ASC, `event_add_time` DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, err
	}
	list = populateEventFields(list)

	if userID != "" {
		participatedIDs := map[string]bool{}
		var parts []model.EventParticipant
		database.DB.Where("`event_part_mini_openid` = ?", userID).Find(&parts)
		for _, p := range parts {
			participatedIDs[strconv.Itoa(int(p.EventID))] = true
		}
		for i := range list {
			idStr := strconv.Itoa(int(list[i].ID))
			list[i].IsJoin = participatedIDs[idStr]
		}
	}

	return map[string]interface{}{"list": list, "total": total}, nil
}

func ViewEvent(id, userID string) (*model.Event, error) {
	var event model.Event
	err := database.DB.Where("`id` = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	database.DB.Model(&event).UpdateColumn("event_view_cnt", event.ViewCnt+1)

	if userID != "" {
		var cnt int64
		database.DB.Model(&model.EventParticipant{}).
			Where("`event_part_event_id` = ? AND `event_part_mini_openid` = ?", id, userID).Count(&cnt)
		if cnt > 0 {
			event.IsJoin = true
		}
	}

	populateEventTimeFields(&event)
	loadEventRoles(&event)

	// Parse obj for desc and img
	if event.Obj != "" {
		var obj eventObj
		json.Unmarshal([]byte(event.Obj), &obj)
		if obj.Desc != "" {
	event.Desc = obj.Desc
	event.Rules = obj.Rules
		}
		if len(obj.Cover) > 0 {
			event.Img = GetFullURL(obj.Cover[0])
		}
		if obj.Rules != "" {
			event.Rules = obj.Rules
		}
	}

	// Fall back to QR as cover image
	if event.Img == "" && event.QR != "" {
		event.Img = event.QR
	}

	// Count participants
	var pCnt int64
	database.DB.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", id).Count(&pCnt)
	event.UserCnt = int(pCnt)

	return &event, nil
}

func EventParticipate(eventID, userID, forms, addIP string) error {
	// Check event exists and is active
	var event model.Event
	if err := database.DB.Where("`id` = ? AND `event_status` = 1", eventID).First(&event).Error; err != nil {
		return fmt.Errorf("项目不存在或已停用")
	}

	// Check registration time
	now := time.Now().UnixMilli()
	if event.RegStart > 0 && now < event.RegStart {
		return fmt.Errorf("报名尚未开始")
	}
	if event.RegEnd > 0 && now > event.RegEnd {
		return fmt.Errorf("报名已结束")
	}

	// Check publish department
	if event.PublishDeptIds != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if !checkPublishDeptAccess(event.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该项目的发布部门范围内")
		}
	}

	// Check duplicate
	var cnt int64
	database.DB.Model(&model.EventParticipant{}).
		Where("`event_part_event_id` = ? AND `event_part_mini_openid` = ?", eventID, userID).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("已参与")
	}

	part := model.EventParticipant{
		EventID:    uint(parseUint(eventID)),
		MiniOpenID: userID,
		Forms:      forms,
		Status:     1,
		AddTime:    database.Now(),
		AddIP:      addIP,
	}
	if err := database.DB.Create(&part).Error; err != nil {
		return err
	}
	database.DB.Model(&event).UpdateColumn("event_join_cnt", event.JoinCnt+1)
	database.DB.Model(&event).UpdateColumn("event_user_cnt", event.UserCnt+1)
	return nil
}

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
		"hasOrganizer":  len(orgIDs) > 0,
		"hasAssistant":  len(astIDs) > 0,
		"hasReferee":    len(refIDs) > 0,
		"organizerIDs":  orgIDs,
		"assistantIDs":  astIDs,
		"refereeIDs":    refIDs,
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
		list[i].UserAvatar = GetFullURL(user.Pic)
		if list[i].Images != "" {
			json.Unmarshal([]byte(list[i].Images), &list[i].ImageList)
			for j := range list[i].ImageList {
				list[i].ImageList[j] = GetFullURL(list[i].ImageList[j])
			}
		}
		if list[i].Videos != "" {
			json.Unmarshal([]byte(list[i].Videos), &list[i].VideoList)
			for j := range list[i].VideoList {
				list[i].VideoList[j] = GetFullURL(list[i].VideoList[j])
			}
		}
	}
	return map[string]interface{}{"list": list, "total": total}, nil
}

func EditEventDynamic(id, title, content, images, videos, editIP string) error {
	return database.DB.Model(&model.EventDynamic{}).
		Where("`event_dynamic_id` = ?", id).
		Updates(map[string]interface{}{
			"event_dynamic_title":      title,
			"event_dynamic_content":    content,
			"event_dynamic_images":     images,
			"event_dynamic_videos":     videos,
			"event_dynamic_edit_time":  database.Now(),
			"event_dynamic_edit_ip":    editIP,
		}).Error
}

func DelEventDynamic(id string) error {
	return database.DB.Where("`event_dynamic_id` = ?", id).Delete(&model.EventDynamic{}).Error
}

func DelEventDynamics(ids []string) error {
	return database.DB.Where("`event_dynamic_id` IN ?", ids).Delete(&model.EventDynamic{}).Error
}

func SaveEventScore(eventID, participantID, score, judgeID string) error {
	// Upsert: find existing or create new
	var existing model.EventScore
	result := database.DB.Where("`event_score_event_id` = ? AND `event_score_participant_id` = ?", eventID, participantID).First(&existing)
	if result.Error == nil {
		return database.DB.Model(&existing).Updates(map[string]interface{}{
			"event_score_score":    score,
			"event_score_judge_id": judgeID,
			"event_score_edit_time": database.Now(),
		}).Error
	}
	es := model.EventScore{
		EventID:       uint(parseUint(eventID)),
		ParticipantID: participantID,
		Score:         score,
		JudgeID:       judgeID,
		AddTime:       database.Now(),
	}
	return database.DB.Create(&es).Error
}

func AdminEditEventScore(id, score string) error {
	return database.DB.Model(&model.EventScore{}).
		Where("`event_score_id` = ?", id).
		Update("event_score_score", score).Error
}

func GetEventScores(eventID string, page, pageSize int) (map[string]interface{}, error) {
	var list []model.EventScore
	var total int64
	query := database.DB.Model(&model.EventScore{}).Where("`event_score_event_id` = ?", eventID)
	query.Count(&total)
	err := query.Order("`event_score_add_time` ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var user model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].ParticipantID).First(&user)
		list[i].ParticipantName = user.Name
		list[i].ParticipantAvatar = GetFullURL(user.Pic)
		if user.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", user.ID).First(&ud)
			if ud.DeptID > 0 {
				var dept model.Department
				database.DB.First(&dept, ud.DeptID)
				list[i].ParticipantDept = dept.Name
				list[i].ParticipantTopDept = getTopDeptName(ud.DeptID)
			}
		}
	}
	return map[string]interface{}{"list": list, "total": total}, nil
}

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
		"title":     "event_title",
		"type":      "event_type",
		"status":    "event_status",
		"order":     "event_order",
		"userCnt":   "event_user_cnt",
		"regStart":  "event_reg_start",
		"regEnd":    "event_reg_end",
		"eventStart": "event_event_start",
		"eventEnd":  "event_event_end",
		"addTime":   "event_add_time",
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
		Title:           title,
		Type:            typ,
		Status:          status,
		CateID:          cateID,
		CateName:        cateName,
		RegStart:        regStart,
		RegEnd:          regEnd,
		EventStart:      eventStart,
		EventEnd:        eventEnd,
		Order:           order,
		Forms:           forms,
		ScoreFields:     scoreFields,
		QR:              qr,
		Obj:             obj,
		DeptID:          deptID,
		PublishDeptIds:  publishDeptIds,
		CreateBy:        createBy,
		AddTime:         database.Now(),
		AddIP:           addIP,
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

// ==================== Internal Helpers ====================

func SaveEventRoles(eventID uint, organizers, assistants, referees []string) {
	database.DB.Where("`event_role_event_id` = ?", eventID).Delete(&model.EventRole{})
	insertRoles := func(users []string, role string) {
		for _, uid := range users {
			if uid == "" {
				continue
			}
			database.DB.Create(&model.EventRole{
				EventID: eventID,
				UserID:  uid,
				Role:    role,
			})
		}
	}
	insertRoles(organizers, "organizer")
	insertRoles(assistants, "assistant")
	insertRoles(referees, "referee")
}

func GetDeptUsers(deptIDs []uint) ([]map[string]interface{}, error) {
	if len(deptIDs) == 0 {
		return nil, nil
	}
	var users []model.User
	database.DB.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs).
		Find(&users)
	var result []map[string]interface{}
	for _, u := range users {
		result = append(result, map[string]interface{}{
			"id":     u.ID,
			"name":   u.Name,
			"avatar": GetFullURL(u.Pic),
			"openid": u.MiniOpenID,
		})
	}
	return result, nil
}

func loadEventRolesForList(list []model.Event, userID string) {
	for i := range list {
		var role model.EventRole
		database.DB.Where("`event_role_event_id` = ? AND `event_role_user_id` = ?", list[i].ID, userID).First(&role)
		if role.ID > 0 {
			switch role.Role {
			case "organizer":
				list[i].RoleName = "工作人员:主办人"
			case "assistant":
				list[i].RoleName = "工作人员:主办人助理"
			case "referee":
				list[i].RoleName = "工作人员:裁判"
			}
		}
	}
}

func populateEventFields(list []model.Event) []model.Event {
	for i := range list {
		populateEventTimeFields(&list[i])
		var obj eventObj
		if list[i].Obj != "" {
			json.Unmarshal([]byte(list[i].Obj), &obj)
		}
		if len(obj.Cover) > 0 {
			list[i].Img = GetFullURL(obj.Cover[0])
		}
		list[i].Desc = obj.Desc
		list[i].Rules = obj.Rules
	}
	return list
}

func populateEventTimeFields(e *model.Event) {
	if e.RegStart > 0 {
		e.RegStartStr = time.UnixMilli(e.RegStart).Format("2006-01-02 15:04")
	}
	if e.RegEnd > 0 {
		e.RegEndStr = time.UnixMilli(e.RegEnd).Format("2006-01-02 15:04")
	}
	if e.EventStart > 0 {
		e.EventStartStr = time.UnixMilli(e.EventStart).Format("2006-01-02 15:04")
	}
	if e.EventEnd > 0 {
		e.EventEndStr = time.UnixMilli(e.EventEnd).Format("2006-01-02 15:04")
	}
	now := time.Now().UnixMilli()
	if e.Status == 0 {
		e.StatusDesc = "已停用"
	} else if e.RegStart > 0 && now < e.RegStart {
		e.StatusDesc = "未开始"
	} else if e.RegEnd > 0 && now > e.RegEnd {
		e.StatusDesc = "报名结束"
	} else if e.EventEnd > 0 && now > e.EventEnd {
		e.StatusDesc = "已结束"
	} else {
		e.StatusDesc = "进行中"
	}
}

func loadEventRoles(e *model.Event) {
	var roles []model.EventRole
	database.DB.Where("`event_role_event_id` = ?", e.ID).Find(&roles)
	for _, r := range roles {
		var user model.User
		database.DB.Where("`user_mini_openid` = ?", r.UserID).First(&user)
		entry := map[string]string{
			"userId": r.UserID,
			"name":   user.Name,
			"avatar": GetFullURL(user.Pic),
		}
		switch r.Role {
		case "organizer":
			e.Organizers = append(e.Organizers, entry)
		case "assistant":
			e.Assistants = append(e.Assistants, entry)
		case "referee":
			e.Referees = append(e.Referees, entry)
		}
	}
}

func parseUint(s string) uint64 {
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

func getTopDeptName(deptID uint) string {
	visited := map[uint]bool{}
	for deptID > 0 {
		if visited[deptID] {
			break
		}
		visited[deptID] = true
		var dept model.Department
		if err := database.DB.First(&dept, deptID).Error; err != nil {
			break
		}
		if dept.ParentID == 0 {
			return dept.Name
		}
		deptID = dept.ParentID
	}
	return ""
}
