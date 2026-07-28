package survey

import (
	"context"
	"encoding/json"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

// SaveDraft 暂存草稿
func (r *ResponseService) SaveDraft(surveyID uint, userID uint, answers map[string]interface{}) (*model.SurveyResponse, error) {
	return r.SaveDraftContext(context.Background(), surveyID, userID, answers)
}

func (r *ResponseService) SaveDraftContext(ctx context.Context, surveyID uint, userID uint, answers map[string]interface{}) (*model.SurveyResponse, error) {
	sv, err := r.Survey.GetContext(ctx, surveyID)
	if err != nil {
		return nil, err
	}
	answersJSON, _ := json.Marshal(answers)
	now := time.Now().UnixMilli()
	resp := &model.SurveyResponse{
		SurveyID: surveyID,
		UserID:   userIDToStr(userID, sv.Anonymous == 1),
		Answers:  string(answersJSON),
		Status:   0,
		AddTime:  now,
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Create(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}
