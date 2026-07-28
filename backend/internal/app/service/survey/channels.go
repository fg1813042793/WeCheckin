package survey

import (
	"context"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func (s *SurveyService) ChannelListContext(ctx context.Context, surveyID uint) ([]model.SurveyChannel, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.SurveyChannel
	err := db.Where("`survey_ch_survey_id` = ?", surveyID).Order("`survey_ch_id` DESC").Find(&list).Error
	return list, err
}

func (s *SurveyService) ChannelListForAdminContext(ctx context.Context, surveyID uint, adminID uint) ([]model.SurveyChannel, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureSurveyVisibleContext(ctx, db, surveyID, adminID); err != nil {
		return nil, err
	}
	return s.ChannelListContext(ctx, surveyID)
}

func (s *SurveyService) ChannelCreateContext(ctx context.Context, channel *model.SurveyChannel) error {
	if channel.AddTime == 0 {
		channel.AddTime = time.Now().UnixMilli()
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Create(channel).Error
}

func (s *SurveyService) ChannelCreateForAdminContext(ctx context.Context, channel *model.SurveyChannel, adminID uint) error {
	if channel.AddTime == 0 {
		channel.AddTime = time.Now().UnixMilli()
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureSurveyVisibleContext(ctx, db, channel.SurveyID, adminID); err != nil {
		return err
	}
	return db.Create(channel).Error
}

func (s *SurveyService) ChannelDeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`survey_ch_id` = ?", id).Delete(&model.SurveyChannel{}).Error
}

func (s *SurveyService) ChannelDeleteForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var channel model.SurveyChannel
	if err := db.Where("`survey_ch_id` = ?", id).First(&channel).Error; err != nil {
		return err
	}
	if err := ensureSurveyVisibleContext(ctx, db, channel.SurveyID, adminID); err != nil {
		return err
	}
	return db.Where("`survey_ch_id` = ?", id).Delete(&model.SurveyChannel{}).Error
}
