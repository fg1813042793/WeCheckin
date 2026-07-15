package service

import (
	"encoding/json"
	"strconv"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// ==================== Internal Helpers ====================

type eventObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
	Rules string   `json:"rules"`
}

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
