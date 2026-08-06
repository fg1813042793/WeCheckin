package enroll

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/publish"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func EnrollJoin(enrollID, userID, day, forms, addIP string, status int) error {
	return EnrollJoinContext(context.Background(), enrollID, userID, day, forms, addIP, status)
}

func EnrollJoinContext(ctx context.Context, enrollID, userID, day, forms, addIP string, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var enroll model.Enroll
	if err := db.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if !publish.HasDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	if !enroll.AllowRepeat {
		var cnt int64
		db.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?", enrollID, userID, day).Count(&cnt)
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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&join).Error; err != nil {
			return err
		}
		if err := tx.Model(&enroll).UpdateColumn("enroll_join_cnt", enroll.JoinCnt+1).Error; err != nil {
			return err
		}
		var eu model.EnrollUser
		result := tx.Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).First(&eu)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return result.Error
			}
			eu = model.EnrollUser{
				EnrollID:   enrollID,
				MiniOpenID: userID,
				JoinCnt:    1,
				DayCnt:     1,
				LastDay:    day,
				AddTime:    database.Now(),
			}
			if err := tx.Create(&eu).Error; err != nil {
				return err
			}
			return tx.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1).Error
		}
		updates := map[string]interface{}{
			"enroll_user_join_cnt":  eu.JoinCnt + 1,
			"enroll_user_last_day":  day,
			"enroll_user_edit_time": database.Now(),
		}
		if eu.LastDay != day {
			updates["enroll_user_day_cnt"] = eu.DayCnt + 1
		}
		return tx.Model(&eu).Updates(updates).Error
	})
}

func EnrollUserSubmit(enrollID, userID, forms, addIP string) error {
	return EnrollUserSubmitContext(context.Background(), enrollID, userID, forms, addIP)
}

func EnrollUserSubmitContext(ctx context.Context, enrollID, userID, forms, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var enroll model.Enroll
	if err := db.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	if enroll.PublishDeptIds != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if !publish.HasDeptAccess(enroll.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该打卡项目的发布部门范围内")
		}
	}
	var cnt int64
	db.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", enrollID, userID).Count(&cnt)
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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&eu).Error; err != nil {
			return err
		}
		return tx.Model(&enroll).UpdateColumn("enroll_user_cnt", enroll.UserCnt+1).Error
	})
}
