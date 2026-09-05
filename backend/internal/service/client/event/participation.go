package event

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/publish"
	"wecheckin/backend/pkg/database"
)

func EventParticipate(eventID, userID, forms, addIP string) error {
	return EventParticipateContext(context.Background(), eventID, userID, forms, addIP)
}

func EventParticipateContext(ctx context.Context, eventID, userID, forms, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	// Check event exists and is active
	var event model.Event
	if err := db.Where("`id` = ? AND `event_status` = 1", eventID).First(&event).Error; err != nil {
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
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if !publish.HasDeptAccess(event.PublishDeptIds, deptIDs) {
			return fmt.Errorf("您不在该项目的发布部门范围内")
		}
	}

	// Check duplicate
	var cnt int64
	db.Model(&model.EventParticipant{}).
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
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&part).Error; err != nil {
			return err
		}
		if err := tx.Model(&event).UpdateColumn("event_join_cnt", event.JoinCnt+1).Error; err != nil {
			return err
		}
		return tx.Model(&event).UpdateColumn("event_user_cnt", event.UserCnt+1).Error
	})
}
