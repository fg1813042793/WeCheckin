package event

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

// ==================== Internal Helpers ====================

type eventObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
	Rules string   `json:"rules"`
}

type DeptUser struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	OpenID string `json:"openid"`
}

type eventUserInfo struct {
	User        model.User
	Avatar      string
	DeptName    string
	TopDeptName string
}

func decodeEventObj(raw string) eventObj {
	var obj eventObj
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &obj)
	}
	return obj
}

func SaveEventRoles(eventID uint, organizers, assistants, referees []string) {
	_ = SaveEventRolesContext(context.Background(), eventID, organizers, assistants, referees)
}

func SaveEventRolesContext(ctx context.Context, eventID uint, organizers, assistants, referees []string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return SaveEventRolesTx(db, eventID, organizers, assistants, referees)
}

func SaveEventRolesTx(tx *gorm.DB, eventID uint, organizers, assistants, referees []string) error {
	if err := tx.Where("`event_role_event_id` = ?", eventID).Delete(&model.EventRole{}).Error; err != nil {
		return err
	}
	insertRoles := func(users []string, role string) error {
		for _, uid := range users {
			if uid == "" {
				continue
			}
			if err := tx.Create(&model.EventRole{
				EventID: eventID,
				UserID:  uid,
				Role:    role,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err := insertRoles(organizers, "organizer"); err != nil {
		return err
	}
	if err := insertRoles(assistants, "assistant"); err != nil {
		return err
	}
	if err := insertRoles(referees, "referee"); err != nil {
		return err
	}
	return nil
}

func GetDeptUsers(deptIDs []uint) ([]DeptUser, error) {
	return GetDeptUsersContext(context.Background(), deptIDs)
}

func GetDeptUsersContext(ctx context.Context, deptIDs []uint) ([]DeptUser, error) {
	if len(deptIDs) == 0 {
		return nil, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var users []model.User
	if err := db.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]DeptUser, 0, len(users))
	for _, u := range users {
		result = append(result, DeptUser{
			ID:     u.ID,
			Name:   u.Name,
			Avatar: media.FullURLWithStaticDomain(u.Pic),
			OpenID: u.MiniOpenID,
		})
	}
	return result, nil
}

func loadEventRolesForList(list []model.Event, userID string) {
	loadEventRolesForListContext(context.Background(), list, userID)
}

func loadEventRolesForListContext(ctx context.Context, list []model.Event, userID string) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if len(list) == 0 || userID == "" {
		return
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

func enrichEventParticipantsWithUserInfoContext(ctx context.Context, db *gorm.DB, list []model.EventParticipant) []model.EventParticipant {
	openIDs := make([]string, 0, len(list))
	for _, item := range list {
		openIDs = append(openIDs, item.MiniOpenID)
	}
	infoByOpenID, err := loadEventUserInfoByOpenIDContext(ctx, db, openIDs)
	if err != nil {
		return list
	}
	for i := range list {
		info := infoByOpenID[list[i].MiniOpenID]
		list[i].UserName = info.User.Name
		list[i].UserAvatar = info.Avatar
		list[i].Mobile = info.User.Mobile
		list[i].DeptName = info.DeptName
		list[i].TopDeptName = info.TopDeptName
	}
	return list
}

func enrichEventScoresWithUserInfoContext(ctx context.Context, db *gorm.DB, list []model.EventScore) []model.EventScore {
	openIDs := make([]string, 0, len(list))
	for _, item := range list {
		openIDs = append(openIDs, item.ParticipantID)
	}
	infoByOpenID, err := loadEventUserInfoByOpenIDContext(ctx, db, openIDs)
	if err != nil {
		return list
	}
	for i := range list {
		info := infoByOpenID[list[i].ParticipantID]
		list[i].ParticipantName = info.User.Name
		list[i].ParticipantAvatar = info.Avatar
		list[i].ParticipantDept = info.DeptName
		list[i].ParticipantTopDept = info.TopDeptName
	}
	return list
}

func loadEventUserInfoByOpenIDContext(ctx context.Context, db *gorm.DB, openIDs []string) (map[string]eventUserInfo, error) {
	result := make(map[string]eventUserInfo)
	uniqueOpenIDs := uniqueNonEmptyEventOpenIDs(openIDs)
	if len(uniqueOpenIDs) == 0 {
		return result, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return result, err
		}
	}

	var users []model.User
	if err := db.Select("id", "user_mini_openid", "user_name", "user_mobile", "user_pic").
		Where("`user_mini_openid` IN ?", uniqueOpenIDs).
		Find(&users).Error; err != nil {
		return result, err
	}
	if len(users) == 0 {
		return result, nil
	}

	userIDToOpenID := make(map[uint]string, len(users))
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		result[user.MiniOpenID] = eventUserInfo{
			User:   user,
			Avatar: media.FullURLWithStaticDomain(user.Pic),
		}
		userIDToOpenID[user.ID] = user.MiniOpenID
		userIDs = append(userIDs, user.ID)
	}

	var userDepts []model.UserDept
	if err := db.Select("id", "user_dept_user_id", "user_dept_dept_id").
		Where("`user_dept_user_id` IN ?", userIDs).
		Order("`id` ASC").
		Find(&userDepts).Error; err != nil {
		return result, err
	}
	if len(userDepts) == 0 {
		return result, nil
	}

	deptIDByUserID := make(map[uint]uint, len(userDepts))
	for _, userDept := range userDepts {
		if deptIDByUserID[userDept.UserID] == 0 {
			deptIDByUserID[userDept.UserID] = userDept.DeptID
		}
	}

	var departments []model.Department
	if err := db.Select("id", "dept_name", "dept_parent_id").Find(&departments).Error; err != nil {
		return result, err
	}
	deptByID := make(map[uint]model.Department, len(departments))
	for _, department := range departments {
		deptByID[department.ID] = department
	}

	for userID, deptID := range deptIDByUserID {
		openID := userIDToOpenID[userID]
		if openID == "" {
			continue
		}
		info := result[openID]
		if department, ok := deptByID[deptID]; ok {
			info.DeptName = department.Name
		}
		info.TopDeptName = topEventDeptNameFromDepartmentMap(deptID, deptByID)
		result[openID] = info
	}
	return result, nil
}

func topEventDeptNameFromDepartmentMap(deptID uint, deptByID map[uint]model.Department) string {
	visited := make(map[uint]struct{})
	for deptID > 0 {
		if _, ok := visited[deptID]; ok {
			return ""
		}
		visited[deptID] = struct{}{}

		department, ok := deptByID[deptID]
		if !ok {
			return ""
		}
		if department.ParentID == 0 {
			return department.Name
		}
		deptID = department.ParentID
	}
	return ""
}

func uniqueNonEmptyEventOpenIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func populateEventFields(list []model.Event) []model.Event {
	for i := range list {
		populateEventTimeFields(&list[i])
		obj := decodeEventObj(list[i].Obj)
		if len(obj.Cover) > 0 {
			list[i].Img = media.FullURLWithStaticDomain(obj.Cover[0])
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
	loadEventRolesContext(context.Background(), e)
}

func loadEventRolesContext(ctx context.Context, e *model.Event) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var roles []model.EventRole
	if err := db.Where("`event_role_event_id` = ?", e.ID).Find(&roles).Error; err != nil || len(roles) == 0 {
		return
	}
	userIDs := make([]string, 0, len(roles))
	for _, r := range roles {
		if r.UserID != "" {
			userIDs = append(userIDs, r.UserID)
		}
	}
	userByOpenID := map[string]model.User{}
	if len(userIDs) > 0 {
		var users []model.User
		if err := db.Where("`user_mini_openid` IN ?", userIDs).Find(&users).Error; err == nil {
			for _, userItem := range users {
				userByOpenID[userItem.MiniOpenID] = userItem
			}
		}
	}
	for _, r := range roles {
		userItem := userByOpenID[r.UserID]
		entry := map[string]string{
			"userId": r.UserID,
			"name":   userItem.Name,
			"avatar": media.FullURLWithStaticDomain(userItem.Pic),
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
	return getTopDeptNameContext(context.Background(), deptID)
}

func getTopDeptNameContext(ctx context.Context, deptID uint) string {
	return dept.TopDeptNameContext(ctx, deptID)
}
