package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

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

func SaveEventScore(eventID, participantID, score, judgeID string) error {
	// Upsert: find existing or create new
	var existing model.EventScore
	result := database.DB.Where("`event_score_event_id` = ? AND `event_score_participant_id` = ?", eventID, participantID).First(&existing)
	if result.Error == nil {
		return database.DB.Model(&existing).Updates(map[string]interface{}{
			"event_score_score":     score,
			"event_score_judge_id":  judgeID,
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
