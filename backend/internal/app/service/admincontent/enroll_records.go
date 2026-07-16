package admincontent

import (
	"wecheckin-backend/backend/internal/app/support/dept"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetEnrollUserList(enrollID, keyword string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	queryBuilder := database.DB.Where("`enroll_user_enroll_id` = ?", enrollID)
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_user_mini_openid` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?)", "%"+keyword+"%")
	}
	err := queryBuilder.Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", list[i].MiniOpenID).First(&u)
		list[i].EnrollTitle = u.Name
		list[i].UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var d model.Department
				database.DB.First(&d, ud.DeptID)
				list[i].DeptName = d.Name
				list[i].TopDeptName = dept.TopDeptName(ud.DeptID)
			}
		}
	}
	return list, nil
}

func GetEnrollJoinList(enrollID, keyword string, page, pageSize int) ([]model.EnrollJoin, int64, error) {
	var list []model.EnrollJoin
	var total int64
	queryBuilder := database.DB.Model(&model.EnrollJoin{})
	if enrollID != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_enroll_id` = ?", enrollID)
	}
	if keyword != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_user_id` IN (SELECT `user_mini_openid` FROM `users` WHERE `user_name` LIKE ?) OR `enroll_join_user_id` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	queryBuilder.Count(&total)
	err := queryBuilder.Order("`enroll_join_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	userMap := map[string]model.User{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userMap[u.MiniOpenID] = u
	}
	for i := range list {
		u, ok := userMap[list[i].UserID]
		if ok {
			list[i].EnrollTitle = u.Name
			list[i].UserName = u.Name
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var d model.Department
				database.DB.First(&d, ud.DeptID)
				list[i].DeptName = d.Name
				list[i].TopDeptName = dept.TopDeptName(ud.DeptID)
			}
		}
	}
	return list, total, nil
}

type EnrollStatItem struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	DeptName    string `json:"deptName"`
	TopDeptName string `json:"topDeptName"`
	JoinCnt     int    `json:"joinCnt"`
	DayCnt      int    `json:"dayCnt"`
}

func GetEnrollStats(enrollID, startDay, endDay string) ([]EnrollStatItem, error) {
	var joins []model.EnrollJoin
	queryBuilder := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` <= ?", endDay)
	}
	err := queryBuilder.Find(&joins).Error
	if err != nil {
		return nil, err
	}
	type tmp struct{ cnt, days int }
	agg := map[string]*tmp{}
	daySet := map[string]map[string]bool{}
	for _, j := range joins {
		if _, ok := agg[j.UserID]; !ok {
			agg[j.UserID] = &tmp{}
			daySet[j.UserID] = map[string]bool{}
		}
		agg[j.UserID].cnt++
		daySet[j.UserID][j.Day] = true
	}
	var result []EnrollStatItem
	for uid, t := range agg {
		t.days = len(daySet[uid])
		item := EnrollStatItem{
			UserID:  uid,
			JoinCnt: t.cnt,
			DayCnt:  t.days,
		}
		var u model.User
		database.DB.Where("`user_mini_openid` = ?", uid).First(&u)
		item.UserName = u.Name
		if u.ID > 0 {
			var ud model.UserDept
			database.DB.Where("`user_dept_user_id` = ?", u.ID).First(&ud)
			if ud.DeptID > 0 {
				var d model.Department
				database.DB.First(&d, ud.DeptID)
				item.DeptName = d.Name
				item.TopDeptName = dept.TopDeptName(ud.DeptID)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func DelEnrollJoin(id string) error {
	var join model.EnrollJoin
	err := database.DB.Where("`id` = ?", id).First(&join).Error
	if err != nil {
		return err
	}
	database.DB.Delete(&join)
	var enroll model.Enroll
	database.DB.Where("`id` = ?", join.EnrollID).First(&enroll)
	if enroll.JoinCnt > 0 {
		database.DB.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt-1)
	}
	return nil
}

func DelEnrollJoins(ids []string) error {
	for _, id := range ids {
		if err := DelEnrollJoin(id); err != nil {
			return err
		}
	}
	return nil
}

func RemoveEnrollUser(enrollID, userID string) error {
	database.DB.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Delete(&model.EnrollUser{})
	database.DB.Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", enrollID, userID).Delete(&model.EnrollJoin{})
	var cnt int64
	database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ?", enrollID).Count(&cnt)
	database.DB.Model(&model.Enroll{}).Where("`id` = ?", enrollID).Update("enroll_user_cnt", cnt)
	return nil
}

func RemoveEnrollUsers(enrollID string, userIDs []string) error {
	for _, uid := range userIDs {
		if err := RemoveEnrollUser(enrollID, uid); err != nil {
			return err
		}
	}
	return nil
}

func EditEnrollUserForms(enrollID, userID, forms string) error {
	return database.DB.Model(&model.EnrollUser{}).
		Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).
		Update("enroll_user_forms", forms).Error
}
