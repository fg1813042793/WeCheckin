package service

import (
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func AdminHome(adminID uint) (map[string]interface{}, error) {
	var admin model.Admin
	database.DB.First(&admin, adminID)

	var userCnt int64
	if admin.Type == 1 || admin.RoleID == 0 {
		database.DB.Model(&model.User{}).Count(&userCnt)
	} else {
		q := database.DB.Model(&model.User{})
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			if role.DataScope == 2 || role.DataScope == 4 {
				var deptIDs []uint
				if role.DataScope == 2 {
					deptIDs = getAdminDeptIDs(admin.ID)
				} else {
					deptIDs = GetRoleDeptIDs(admin.RoleID)
				}
				if len(deptIDs) > 0 {
					q = q.Where("`id` IN (SELECT `user_dept_user_id` FROM `user_depts` WHERE `user_dept_dept_id` IN ?)", deptIDs)
				}
			} else if role.DataScope == 3 {
				q = q.Where("1 = 0")
			}
		}
		q.Count(&userCnt)
	}

	var enrollCnt int64
	q := database.DB.Model(&model.Enroll{})
	where, args := BuildDataScopeFilter(&admin, "`enroll_dept_id`", "`enroll_create_by`")
	if where != "" {
		q = q.Where(where, args...)
	}
	q.Count(&enrollCnt)

	var newsCnt int64
	q2 := database.DB.Model(&model.News{})
	where2, args2 := BuildDataScopeFilter(&admin, "`news_dept_id`", "`news_create_by`")
	if where2 != "" {
		q2 = q2.Where(where2, args2...)
	}
	q2.Count(&newsCnt)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999, now.Location()).UnixMilli()
	var joinCnt int64
	q3 := database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_add_time` BETWEEN ? AND ?", todayStart, todayEnd)
	// Filter join counts by enrolls the admin can see
	if where != "" {
		q3 = q3.Where("`enroll_join_enroll_id` IN (SELECT `id` FROM `enrolls` WHERE "+where+")", args...)
	}
	q3.Count(&joinCnt)

	var eventCnt int64
	q4 := database.DB.Model(&model.Event{})
	where4, args4 := BuildDataScopeFilter(&admin, "`event_dept_id`", "`event_create_by`")
	if where4 != "" {
		q4 = q4.Where(where4, args4...)
	}
	q4.Count(&eventCnt)

	var eventUserCnt int64
	database.DB.Model(&model.EventParticipant{}).Count(&eventUserCnt)

	var mgrCnt int64
	database.DB.Model(&model.Admin{}).Count(&mgrCnt)

	result := map[string]interface{}{
		"userCnt":      userCnt,
		"enrollCnt":    enrollCnt,
		"newsCnt":      newsCnt,
		"joinCnt":      joinCnt,
		"eventCnt":     eventCnt,
		"eventUserCnt": eventUserCnt,
		"mgrCnt":       mgrCnt,
	}
	return result, nil
}

func ClearVouchData() error {
	return database.DB.Model(&model.Enroll{}).Where("1 = 1").Update("enroll_vouch", 0).Error
}
