package formkitadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wecheckin/backend/internal/formkit/report"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ReportResult struct {
	Schema string             `json:"schema"`
	Table  report.Table       `json:"table"`
	Stats  []report.FieldStat `json:"stats"`
	Count  int                `json:"count"`
	Title  string             `json:"title"`
}

func EventReportContext(ctx context.Context, eventID string) (ReportResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var event model.Event
	if err := db.Where("`id` = ?", eventID).First(&event).Error; err != nil {
		return ReportResult{}, err
	}
	var participants []model.EventParticipant
	if err := db.Where("`event_part_event_id` = ?", eventID).
		Order("`event_part_add_time` DESC").
		Find(&participants).Error; err != nil {
		return ReportResult{}, err
	}
	items := make([]report.AnswerItem, 0, len(participants))
	for _, participant := range participants {
		items = append(items, report.AnswerItem{
			UserID:  participant.MiniOpenID,
			AddTime: time.UnixMilli(participant.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   participant.Forms,
		})
	}
	table, err := report.RenderAnswers(event.Forms, items)
	if err != nil {
		return ReportResult{}, err
	}
	return ReportResult{
		Schema: event.Forms,
		Table:  table,
		Stats:  report.FieldStats(event.Forms, items, "count"),
		Count:  len(participants),
		Title:  event.Title,
	}, nil
}

func EventReportCSVContext(ctx context.Context, eventID string) ([]byte, string, error) {
	result, err := EventReportContext(ctx, eventID)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("event_%s_%d.csv", report.SanitizeFilename(result.Title), time.Now().Unix())
	return report.ToCSV(result.Table), filename, nil
}

func EnrollReportContext(ctx context.Context, enrollID string) (ReportResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var enroll model.Enroll
	if err := db.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return ReportResult{}, err
	}
	var joins []model.EnrollJoin
	if err := db.Where("`enroll_join_enroll_id` = ?", enrollID).
		Order("`enroll_join_add_time` DESC").
		Find(&joins).Error; err != nil {
		return ReportResult{}, err
	}
	items := make([]report.AnswerItem, 0, len(joins))
	for _, join := range joins {
		items = append(items, report.AnswerItem{
			UserID:  join.UserID,
			AddTime: time.UnixMilli(join.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   join.Forms,
		})
	}
	table, err := report.RenderAnswers(enroll.Forms, items)
	if err != nil {
		return ReportResult{}, err
	}
	return ReportResult{
		Schema: enroll.Forms,
		Table:  table,
		Stats:  report.FieldStats(enroll.Forms, items, "count"),
		Count:  len(joins),
		Title:  enroll.Title,
	}, nil
}

func EnrollReportCSVContext(ctx context.Context, enrollID string) ([]byte, string, error) {
	result, err := EnrollReportContext(ctx, enrollID)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("enroll_%s_%d.csv", report.SanitizeFilename(result.Title), time.Now().Unix())
	return report.ToCSV(result.Table), filename, nil
}

func SurveyReportContext(ctx context.Context, surveyID uint) (ReportResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var survey model.Survey
	if err := db.Where("`survey_id` = ?", surveyID).First(&survey).Error; err != nil {
		return ReportResult{}, err
	}
	var responses []model.SurveyResponse
	if err := db.Where("`survey_resp_survey_id` = ?", surveyID).
		Order("`survey_resp_id` DESC").
		Find(&responses).Error; err != nil {
		return ReportResult{}, err
	}
	items := make([]report.AnswerItem, 0, len(responses))
	for _, response := range responses {
		items = append(items, report.AnswerItem{
			UserID:  response.UserID,
			AddTime: time.UnixMilli(response.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   response.Answers,
		})
	}
	table, err := report.RenderAnswers(survey.Schema, items)
	if err != nil {
		return ReportResult{}, err
	}
	return ReportResult{
		Schema: survey.Schema,
		Table:  table,
		Stats:  report.FieldStats(survey.Schema, items, reportStatMode(survey.Settings)),
		Count:  len(responses),
		Title:  survey.Title,
	}, nil
}

func SurveyReportCSVContext(ctx context.Context, surveyID uint) ([]byte, string, error) {
	result, err := SurveyReportContext(ctx, surveyID)
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("survey_%s_%d.csv", report.SanitizeFilename(result.Title), time.Now().Unix())
	return report.ToCSV(result.Table), filename, nil
}

func reportStatMode(settingsJSON string) string {
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
