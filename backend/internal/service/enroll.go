package service

import (
	"fmt"
	"time"

	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func GetEnrollList() ([]model.Enroll, error) {
	var enrollList []model.Enroll
	err := database.DB.Where("`ENROLL_STATUS` = 1").Order("`ENROLL_ORDER` ASC, `ENROLL_ADD_TIME` DESC").Find(&enrollList).Error
	if err != nil {
		return nil, err
	}
	return enrollList, nil
}

func ViewEnroll(id string) (*model.Enroll, error) {
	var enroll model.Enroll
	err := database.DB.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	database.DB.Model(&enroll).UpdateColumn("ENROLL_VIEW_CNT", enroll.ViewCnt+1)
	return &enroll, nil
}

func GetEnrollJoinByDay(enrollID, userID, day string) ([]model.EnrollJoin, error) {
	var list []model.EnrollJoin
	query := database.DB.Where("`ENROLL_JOIN_ENROLL_ID` = ? AND `ENROLL_JOIN_USER_ID` = ?", enrollID, userID)
	if day != "" {
		query = query.Where("`ENROLL_JOIN_DAY` = ?", day)
	}
	err := query.Order("`ENROLL_JOIN_DAY` ASC, `ENROLL_JOIN_ADD_TIME` ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetEnrollUserRank(enrollID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`ENROLL_USER_ENROLL_ID` = ?", enrollID).
		Order("`ENROLL_USER_JOIN_CNT` DESC, `ENROLL_USER_DAY_CNT` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetMyEnrollUserList(userID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`ENROLL_USER_MINI_OPENID` = ?", userID).Order("`ENROLL_USER_ADD_TIME` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetMyEnrollJoinList(userID string) ([]model.EnrollJoin, error) {
	var list []model.EnrollJoin
	err := database.DB.Where("`ENROLL_JOIN_USER_ID` = ?", userID).Order("`ENROLL_JOIN_ADD_TIME` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func EnrollJoin(enrollID, userID, day, forms, addIP string, status int) error {
	var cnt int64
	database.DB.Model(&model.EnrollJoin{}).Where("`ENROLL_JOIN_ENROLL_ID` = ? AND `ENROLL_JOIN_USER_ID` = ? AND `ENROLL_JOIN_DAY` = ?", enrollID, userID, day).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("已打卡")
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
	var enroll model.Enroll
	database.DB.Where("`id` = ?", enrollID).First(&enroll)
	database.DB.Model(&enroll).UpdateColumn("ENROLL_JOIN_CNT", enroll.JoinCnt+1)

	var eu model.EnrollUser
	result := database.DB.Where("`ENROLL_USER_ENROLL_ID` = ? AND `ENROLL_USER_MINI_OPENID` = ?", enrollID, userID).First(&eu)
	if result.Error != nil {
		eu = model.EnrollUser{
			EnrollID:   enrollID,
			MiniOpenID: userID,
			JoinCnt:    1,
			DayCnt:     enroll.DayCnt,
			LastDay:    day,
			AddTime:    database.Now(),
		}
		database.DB.Create(&eu)
		database.DB.Model(&enroll).UpdateColumn("ENROLL_USER_CNT", enroll.UserCnt+1)
	} else {
		database.DB.Model(&eu).Updates(map[string]interface{}{
			"ENROLL_USER_JOIN_CNT":  eu.JoinCnt + 1,
			"ENROLL_USER_LAST_DAY":  day,
			"ENROLL_USER_EDIT_TIME": database.Now(),
		})
	}
	return nil
}

func getJoinStatusDesc(status int) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "已通过"
	case 2:
		return "未通过"
	default:
		return "未知"
	}
}

func getTimeShow(t int64) string {
	return time.UnixMilli(t).Format("2006-01-02 15:04:05")
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}
