package service

import (
	"fmt"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func EnrollJoin(enrollID, userID, day, forms, addIP string, status int) error {
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if !checkPublishDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	if !enroll.AllowRepeat {
		var cnt int64
		database.DB.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?", enrollID, userID, day).Count(&cnt)
		if cnt > 0 {
			return fmt.Errorf("已打卡")
		}
	}
	join := model.EnrollJoin{
		EnrollID: enrollID,
		UserID:   userID,
		Day:      day,
		Forms:    forms,
		Status:   status,
		AddTime:  database.Now(),
		AddIP:    addIP,
	}
	if err := database.DB.Create(&join).Error; err != nil {
		return err
	}
	database.DB.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt+1)

	var eu model.EnrollUser
	result := database.DB.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).First(&eu)
	if result.Error != nil {
		eu = model.EnrollUser{
			EnrollID:   enrollID,
			MiniOpenID: userID,
			JoinCnt:    1,
			DayCnt:     1,
			LastDay:    day,
			AddTime:    database.Now(),
		}
		database.DB.Create(&eu)
		database.DB.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1)
	} else {
		updates := map[string]interface{}{
			"enroll_user_join_cnt":  eu.JoinCnt + 1,
			"enroll_user_last_day":  day,
			"enroll_user_edit_time": database.Now(),
		}
		// Check if this is a new day
		if eu.LastDay != day {
			updates["enroll_user_day_cnt"] = eu.DayCnt + 1
		}
		database.DB.Model(&eu).Updates(updates)
	}
	return nil
}

func EnrollUserSubmit(enrollID, userID, forms, addIP string) error {
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if !checkPublishDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	var cnt int64
	database.DB.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("已参与")
	}
	eu := model.EnrollUser{
		EnrollID:   enrollID,
		MiniOpenID: userID,
		Forms:      forms,
		AddTime:    database.Now(),
		AddIP:      addIP,
	}
	if err := database.DB.Create(&eu).Error; err != nil {
		return err
	}
	database.DB.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1)
	return nil
}
