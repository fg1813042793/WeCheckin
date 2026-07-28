package survey

import (
	"context"
	"encoding/json"
	"time"

	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type DailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type DeviceStat struct {
	Mobile int64 `json:"mobile"`
	PC     int64 `json:"pc"`
}

type StatisticResult struct {
	Survey     *model.Survey      `json:"survey"`
	Total      int64              `json:"total"`
	TodayCount int64              `json:"todayCount"`
	Daily      []DailyCount       `json:"daily"`
	DeviceStat DeviceStat         `json:"deviceStat"`
	FieldStats []report.FieldStat `json:"fieldStats"`
	ViewURL    string             `json:"viewUrl"`
}

func (s *SurveyService) StatisticContext(ctx context.Context, surveyID uint, viewURL string) (StatisticResult, error) {
	sv, err := s.GetContext(ctx, surveyID)
	if err != nil {
		return StatisticResult{}, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var total int64
	if err := db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID).Count(&total).Error; err != nil {
		return StatisticResult{}, err
	}
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	var todayCount int64
	if err := db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ?", surveyID, dayStart).Count(&todayCount).Error; err != nil {
		return StatisticResult{}, err
	}
	daily := make([]DailyCount, 7)
	dayCountMap := map[string]int64{}
	sevenDayStart := time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, now.Location())
	tomorrowStart := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	type dailyRow struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}
	var dailyRows []dailyRow
	if err := db.Model(&model.SurveyResponse{}).
		Select("DATE_FORMAT(FROM_UNIXTIME(`survey_resp_add_time` / 1000), '%m-%d') AS date, COUNT(*) AS count").
		Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ? AND `survey_resp_add_time` < ?", surveyID, sevenDayStart.UnixMilli(), tomorrowStart.UnixMilli()).
		Group("date").
		Scan(&dailyRows).Error; err != nil {
		return StatisticResult{}, err
	}
	for _, row := range dailyRows {
		dayCountMap[row.Date] = row.Count
	}
	for i := 0; i < 7; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		date := day.Format("01-02")
		daily[6-i] = DailyCount{Date: date, Count: dayCountMap[date]}
	}
	var deviceStat DeviceStat
	if err := db.Model(&model.SurveyResponse{}).
		Select("COALESCE(SUM(CASE WHEN `survey_resp_device` LIKE ? THEN 1 ELSE 0 END), 0) AS mobile, COALESCE(SUM(CASE WHEN `survey_resp_device` != '' AND `survey_resp_device` NOT LIKE ? THEN 1 ELSE 0 END), 0) AS pc", "%Mobile%", "%Mobile%").
		Where("`survey_resp_survey_id` = ?", surveyID).
		Scan(&deviceStat).Error; err != nil {
		return StatisticResult{}, err
	}
	var answerRows []model.SurveyResponse
	if err := db.Select("survey_resp_answers").Where("`survey_resp_survey_id` = ?", surveyID).Find(&answerRows).Error; err != nil {
		return StatisticResult{}, err
	}
	items := make([]report.AnswerItem, len(answerRows))
	for i, response := range answerRows {
		items[i] = report.AnswerItem{Forms: response.Answers}
	}
	fieldStats := report.FieldStats(sv.Schema, items, surveyStatMode(sv.Settings))
	return StatisticResult{
		Survey:     sv,
		Total:      total,
		TodayCount: todayCount,
		Daily:      daily,
		DeviceStat: deviceStat,
		FieldStats: fieldStats,
		ViewURL:    viewURL,
	}, nil
}

func surveyStatMode(settingsJSON string) string {
	statMode := "count"
	if settingsJSON == "" {
		return statMode
	}
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return statMode
	}
	resultConfig, ok := settings["resultConfig"].(map[string]interface{})
	if !ok {
		return statMode
	}
	if statType, ok := resultConfig["statType"].(string); ok && statType != "" {
		return statType
	}
	return statMode
}
