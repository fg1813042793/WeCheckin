package poststat

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wecheckin/backend/internal/formkit/report"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

// Process handles postStat rules after a survey response is submitted.
func Process(surveyID uint, userID uint, nickname string, currentAnswers string) {
	if err := ProcessContext(context.Background(), surveyID, userID, nickname, currentAnswers); err != nil {
		logger.Logger.Printf("[PostStat] process error surveyId=%d: %v", surveyID, err)
	}
}

func ProcessContext(ctx context.Context, surveyID uint, userID uint, nickname string, currentAnswers string) error {
	return ProcessResponseContext(ctx, surveyID, 0, userID, nickname, currentAnswers)
}

func ProcessResponseContext(ctx context.Context, surveyID, responseID uint, userID uint, nickname string, currentAnswers string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var sv model.Survey
	if err := db.Select("survey_id", "survey_title", "survey_schema", "survey_settings").
		Where("`survey_id` = ?", surveyID).
		First(&sv).Error; err != nil {
		return fmt.Errorf("load post-submit survey %d: %w", surveyID, err)
	}
	if sv.Settings == "" {
		return nil
	}
	rules := parseRules(sv.Settings)
	if len(rules) == 0 {
		return nil
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
			return fmt.Errorf("load post-submit responses: %w", err)
		}
		totalCount = int64(len(answerRows))
		items = make([]report.AnswerItem, len(answerRows))
		for i, row := range answerRows {
			items[i] = report.AnswerItem{Forms: row.Answers}
		}
	} else if err := db.Model(&model.SurveyResponse{}).
		Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).
		Count(&totalCount).Error; err != nil {
		return fmt.Errorf("count post-submit responses: %w", err)
	}

	submitter := resolveSubmitterContext(ctx, sv.Schema, currentAnswers, nickname, userID)
	now := time.Now()
	total := int(totalCount)

	for index, rule := range rules {
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

		msg := renderNotificationMessage(rule.MessageTemplate, sv.Title, submitter, now, total, fieldStats, statMode)
		if err := dispatchRuleNotifications(ctx, configuredNotificationDispatcher(), ruleNotificationInput{
			SurveyID: surveyID, ResponseID: responseID, RuleIndex: index,
			SurveyTitle: sv.Title, Message: msg, Rule: rule,
		}); err != nil {
			return fmt.Errorf("enqueue post-submit notification rule %d: %w", index, err)
		}
	}
	return nil
}

func renderNotificationMessage(template, title, submitter string, now time.Time, total int, fieldStats []report.FieldStat, statMode string) string {
	if template == "" {
		template = "📋 问卷「{title}」收到新答卷\n提交人：{submitter}　时间：{date}\n共 {total} 份提交\n\n{result}"
	}
	questionCount := 0
	for _, field := range fieldStats {
		if field.Type != "divider" && field.Type != "description" {
			questionCount++
		}
	}
	message := strings.ReplaceAll(template, "{title}", title)
	message = strings.ReplaceAll(message, "{questionCount}", strconv.Itoa(questionCount))
	message = strings.ReplaceAll(message, "{total}", strconv.Itoa(total))
	message = strings.ReplaceAll(message, "{submitter}", submitter)
	message = strings.ReplaceAll(message, "{date}", now.Format("2006-01-02 15:04:05"))
	return strings.ReplaceAll(message, "{result}", buildResultText(fieldStats, statMode))
}

func postStatNeedsAggregateResponses(rules []Rule) bool {
	for _, rule := range rules {
		if rule.StatScope != "single" {
			return true
		}
	}
	return false
}
