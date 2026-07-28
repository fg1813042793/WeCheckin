package poststat

import (
	"context"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
)

// Process handles postStat rules after a survey response is submitted.
func Process(surveyID uint, userID uint, nickname string, currentAnswers string) {
	ProcessContext(context.Background(), surveyID, userID, nickname, currentAnswers)
}

func ProcessContext(ctx context.Context, surveyID uint, userID uint, nickname string, currentAnswers string) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var sv model.Survey
	if err := db.Select("survey_id", "survey_title", "survey_schema", "survey_settings").
		Where("`survey_id` = ?", surveyID).
		First(&sv).Error; err != nil {
		logger.Logger.Printf("[PostStat] survey not found: %d", surveyID)
		return
	}
	if sv.Settings == "" {
		return
	}
	rules := parseRules(sv.Settings)
	if len(rules) == 0 {
		return
	}

	var items []report.AnswerItem
	var totalCount int64
	if postStatNeedsAggregateResponses(rules) {
		var answerRows []struct {
			Answers string `gorm:"column:survey_resp_answers"`
		}
		if err := db.Model(&model.SurveyResponse{}).
			Select("survey_resp_answers").
			Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).
			Find(&answerRows).Error; err != nil {
			logger.Logger.Printf("[PostStat] response query error: %v", err)
			return
		}
		totalCount = int64(len(answerRows))
		items = make([]report.AnswerItem, len(answerRows))
		for i, row := range answerRows {
			items[i] = report.AnswerItem{Forms: row.Answers}
		}
	} else if err := db.Model(&model.SurveyResponse{}).
		Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).
		Count(&totalCount).Error; err != nil {
		logger.Logger.Printf("[PostStat] response count error: %v", err)
		return
	}

	submitter := resolveSubmitterContext(ctx, sv.Schema, currentAnswers, nickname, userID)
	now := time.Now()
	dateStr := now.Format("2006-01-02 15:04:05")
	total := int(totalCount)

	for _, rule := range rules {
		statMode := "value"
		if rule.StatField == "label" {
			statMode = "count"
		}
		var fieldStats []report.FieldStat
		if rule.StatScope == "single" {
			singleItems := []report.AnswerItem{{Forms: currentAnswers}}
			fieldStats = report.FieldStats(sv.Schema, singleItems, statMode)
		} else {
			fieldStats = report.FieldStats(sv.Schema, items, statMode)
		}

		msg := rule.MessageTemplate
		if msg == "" {
			msg = "📋 问卷「{title}」收到新答卷\n提交人：{submitter}　时间：{date}\n共 {total} 份提交\n\n{result}"
		}
		questionCount := 0
		for _, fs := range fieldStats {
			if fs.Type != "divider" && fs.Type != "description" {
				questionCount++
			}
		}
		msg = strings.ReplaceAll(msg, "{title}", sv.Title)
		msg = strings.ReplaceAll(msg, "{questionCount}", strconv.Itoa(questionCount))
		msg = strings.ReplaceAll(msg, "{total}", strconv.Itoa(total))
		msg = strings.ReplaceAll(msg, "{submitter}", submitter)
		msg = strings.ReplaceAll(msg, "{date}", dateStr)
		msg = strings.ReplaceAll(msg, "{result}", buildResultText(fieldStats, statMode))

		if rule.NotifyChannel == "webhook" || rule.NotifyChannel == "both" {
			if rule.WebhookURL != "" {
				go sendWebhook(rule.WebhookType, rule.WebhookURL, sv.Title, msg)
			}
		}
		if rule.NotifyChannel == "internal" || rule.NotifyChannel == "both" {
			go sendInternalNotificationContext(context.Background(), surveyID, sv.Title, rule.NotifyAdmin, rule.NotifyUserIds, msg)
		}
	}
}

func postStatNeedsAggregateResponses(rules []Rule) bool {
	for _, rule := range rules {
		if rule.StatScope != "single" {
			return true
		}
	}
	return false
}
