package poststat

import (
	"context"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
)

func sendInternalNotification(surveyID uint, title string, notifyAdmin bool, notifyUserIds string, msg string) {
	sendInternalNotificationContext(context.Background(), surveyID, title, notifyAdmin, notifyUserIds, msg)
}

func sendInternalNotificationContext(ctx context.Context, surveyID uint, title string, notifyAdmin bool, notifyUserIds string, msg string) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	surveyIDStr := strconv.FormatUint(uint64(surveyID), 10)
	if notifyAdmin {
		notify := model.Notify{
			Title:      "问卷统计通知: " + title,
			Content:    msg,
			Type:       "survey_stat",
			SourceID:   surveyIDStr,
			SourceType: "survey",
			UserID:     "",
			AddTime:    time.Now().UnixMilli(),
		}
		if err := db.Create(&notify).Error; err != nil {
			logger.Logger.Printf("[PostStat] create notify error: %v", err)
		}
	}

	for _, id := range strings.Split(notifyUserIds, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		notify := model.Notify{
			Title:      "问卷统计通知: " + title,
			Content:    msg,
			Type:       "survey_stat",
			SourceID:   surveyIDStr,
			SourceType: "survey",
			UserID:     id,
			AddTime:    time.Now().UnixMilli(),
		}
		if err := db.Create(&notify).Error; err != nil {
			logger.Logger.Printf("[PostStat] create notify error: %v", err)
		}
	}
}
