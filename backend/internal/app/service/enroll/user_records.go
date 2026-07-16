package enroll

import (
	"time"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetEnrollUserRank(enrollID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`enroll_user_enroll_id` = ?", enrollID).
		Order("`enroll_user_join_cnt` DESC, `enroll_user_day_cnt` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetMyEnrollUserList(userID string) ([]model.EnrollUser, error) {
	var list []model.EnrollUser
	err := database.DB.Where("`enroll_user_mini_openid` = ?", userID).Order("`enroll_user_add_time` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	// Populate title, daily limit, today's check-in status, and recalculate dayCnt
	today := time.Now().Format("2006-01-02")
	for i := range list {
		var enroll model.Enroll
		if err := database.DB.Where("`id` = ?", list[i].EnrollID).First(&enroll).Error; err == nil {
			list[i].EnrollTitle = enroll.Title
			list[i].DailyLimit = enroll.DailyLimit
		}
		// Recalculate dayCnt from actual unique check-in days
		var uniqueDays int64
		database.DB.Model(&model.EnrollJoin{}).
			Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ?", list[i].EnrollID, userID).
			Select("COUNT(DISTINCT `enroll_join_day`)").Scan(&uniqueDays)
		list[i].DayCnt = int(uniqueDays)
		var todayCnt int64
		database.DB.Model(&model.EnrollJoin{}).
			Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?",
				list[i].EnrollID, userID, today).Count(&todayCnt)
		list[i].CheckedInToday = todayCnt > 0
		list[i].TodayJoinCnt = int(todayCnt)
	}
	return list, nil
}
