package event

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type EventScoreListResponse struct {
	List  []model.EventScore `json:"list"`
	Total int64              `json:"total"`
}

func SaveEventScore(eventID, participantID, score, judgeID string) error {
	return SaveEventScoreContext(context.Background(), eventID, participantID, score, judgeID)
}

func SaveEventScoreContext(ctx context.Context, eventID, participantID, score, judgeID string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	// Upsert: find existing or create new
	var existing model.EventScore
	result := db.Where("`event_score_event_id` = ? AND `event_score_participant_id` = ?", eventID, participantID).First(&existing)
	if result.Error == nil {
		return db.Model(&existing).Updates(map[string]interface{}{
			"event_score_score":     score,
			"event_score_judge_id":  judgeID,
			"event_score_edit_time": database.Now(),
		}).Error
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	es := model.EventScore{
		EventID:       uint(parseUint(eventID)),
		ParticipantID: participantID,
		Score:         score,
		JudgeID:       judgeID,
		AddTime:       database.Now(),
	}
	return db.Create(&es).Error
}

func SaveEventScoreForAdminContext(ctx context.Context, eventID, participantID, score, judgeID string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEventVisibleContext(ctx, db, eventID, adminID); err != nil {
		return err
	}
	return SaveEventScoreContext(ctx, eventID, participantID, score, judgeID)
}

func GetEventScores(eventID string, page, pageSize int) (EventScoreListResponse, error) {
	return GetEventScoresContext(context.Background(), eventID, page, pageSize)
}

func GetEventScoresContext(ctx context.Context, eventID string, page, pageSize int) (EventScoreListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.EventScore
	var total int64
	query := db.Model(&model.EventScore{}).Where("`event_score_event_id` = ?", eventID)
	if err := query.Count(&total).Error; err != nil {
		return EventScoreListResponse{}, err
	}
	err := query.Order("`event_score_add_time` ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return EventScoreListResponse{}, err
	}
	list = enrichEventScoresWithUserInfoContext(ctx, db, list)
	return EventScoreListResponse{List: list, Total: total}, nil
}

func GetEventScoresForAdminContext(ctx context.Context, eventID string, page, pageSize int, adminID uint) (EventScoreListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureEventVisibleContext(ctx, db, eventID, adminID); err != nil {
		return EventScoreListResponse{}, err
	}
	return GetEventScoresContext(ctx, eventID, page, pageSize)
}

func AdminEditEventScoreForAdminContext(ctx context.Context, id, score string, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var eventScore model.EventScore
	if err := db.Where("`event_score_id` = ?", id).First(&eventScore).Error; err != nil {
		return err
	}
	if err := ensureEventVisibleContext(ctx, db, strconv.Itoa(int(eventScore.EventID)), adminID); err != nil {
		return err
	}
	return access.RequireRowsAffected(db.Model(&model.EventScore{}).
		Where("`event_score_id` = ?", id).
		Update("event_score_score", score))
}
