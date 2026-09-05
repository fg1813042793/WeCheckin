package event

import (
	"context"
	"encoding/json"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
)

func ViewEvent(id, userID string) (*model.Event, error) {
	return ViewEventContext(context.Background(), id, userID)
}

func ViewEventContext(ctx context.Context, id, userID string) (*model.Event, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var event model.Event
	err := db.Where("`id` = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	db.Model(&event).UpdateColumn("event_view_cnt", event.ViewCnt+1)

	if userID != "" {
		var cnt int64
		db.Model(&model.EventParticipant{}).
			Where("`event_part_event_id` = ? AND `event_part_mini_openid` = ?", id, userID).Count(&cnt)
		if cnt > 0 {
			event.IsJoin = true
		}
	}

	populateEventTimeFields(&event)
	loadEventRolesContext(ctx, &event)

	// Parse obj for desc and img
	if event.Obj != "" {
		var obj eventObj
		json.Unmarshal([]byte(event.Obj), &obj)
		if obj.Desc != "" {
			event.Desc = obj.Desc
			event.Rules = obj.Rules
		}
		if len(obj.Cover) > 0 {
			event.Img = media.FullURLWithStaticDomain(obj.Cover[0])
		}
		if obj.Rules != "" {
			event.Rules = obj.Rules
		}
	}

	// Fall back to QR as cover image
	if event.Img == "" && event.QR != "" {
		event.Img = event.QR
	}

	// Count participants
	var pCnt int64
	db.Model(&model.EventParticipant{}).Where("`event_part_event_id` = ?", id).Count(&pCnt)
	event.UserCnt = int(pCnt)

	return &event, nil
}
