package poststat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wecheckin/backend/internal/model"
	notificationmodel "wecheckin/backend/internal/model/notification"
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/pkg/database"
)

type examRuleNotificationInput struct {
	ExamID    uint
	RecordID  uint
	RuleIndex int
	ExamTitle string
	Message   string
	Rule      Rule
}

// ProcessExamSubmissionContext handles postStat rules after an exam record is persisted.
func ProcessExamSubmissionContext(ctx context.Context, examID, recordID, userID uint, score, fullScore int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var exam model.Exam
	if err := db.Select("exam_id", "exam_title", "exam_settings").Where("`exam_id` = ?", examID).First(&exam).Error; err != nil {
		return fmt.Errorf("load post-submit exam %d: %w", examID, err)
	}
	rules := parseRules(exam.Settings)
	if len(rules) == 0 {
		return nil
	}
	var total int64
	if err := db.Model(&model.ExamRecord{}).Where("`exam_r_exam_id` = ? AND `exam_r_status` >= 1", examID).Count(&total).Error; err != nil {
		return fmt.Errorf("count post-submit exam records: %w", err)
	}
	submitter := resolveExamSubmitterContext(ctx, userID)
	now := time.Now()
	for index, rule := range rules {
		message := renderExamNotificationMessage(rule.MessageTemplate, exam.Title, submitter, now, int(total), score, fullScore, rule.CalculateTotalScore)
		for _, request := range examNotificationRequests(examRuleNotificationInput{
			ExamID: examID, RecordID: recordID, RuleIndex: index, ExamTitle: exam.Title, Message: message, Rule: rule,
		}) {
			dispatcher := configuredNotificationDispatcher()
			if dispatcher == nil {
				return ErrNotificationDispatcherUnavailable
			}
			if err := dispatcher.Enqueue(ctx, request); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderExamNotificationMessage(template, title, submitter string, now time.Time, total, score, fullScore int, calculateTotalScore bool) string {
	if template == "" {
		template = "📝 考试「{title}」收到新答卷\n提交人：{submitter}　时间：{date}\n共 {total} 份提交\n\n{result}"
	}
	result, scoreText, fullScoreText := "", "", ""
	if calculateTotalScore {
		result = fmt.Sprintf("得分 %d / %d", score, fullScore)
		scoreText, fullScoreText = strconv.Itoa(score), strconv.Itoa(fullScore)
	}
	message := strings.ReplaceAll(template, "{title}", title)
	message = strings.ReplaceAll(message, "{submitter}", submitter)
	message = strings.ReplaceAll(message, "{date}", now.Format("2006-01-02 15:04:05"))
	message = strings.ReplaceAll(message, "{total}", strconv.Itoa(total))
	message = strings.ReplaceAll(message, "{score}", scoreText)
	message = strings.ReplaceAll(message, "{fullScore}", fullScoreText)
	return strings.ReplaceAll(message, "{result}", result)
}

func examNotificationRequests(input examRuleNotificationInput) []NotificationRequest {
	sourceID := strconv.FormatUint(uint64(input.RecordID), 10)
	requests := make([]NotificationRequest, 0, 2)
	if (input.Rule.NotifyChannel == "webhook" || input.Rule.NotifyChannel == "both") && strings.TrimSpace(input.Rule.WebhookURL) != "" {
		requests = append(requests, NotificationRequest{
			IdempotencyKey: examPostStatIdempotencyKey(input, notificationmodel.ChannelWebhook),
			SourceType:     "exam_record", SourceID: sourceID, Channel: notificationmodel.ChannelWebhook,
			Recipient: notificationoutboxapp.WebhookRecipient{Type: input.Rule.WebhookType, URL: strings.TrimSpace(input.Rule.WebhookURL)},
			Payload:   notificationoutboxapp.MessagePayload{Title: input.ExamTitle, Content: input.Message},
		})
	}
	userIDs := parseNotificationUserIDs(input.Rule.NotifyUserIds)
	if (input.Rule.NotifyChannel == "internal" || input.Rule.NotifyChannel == "both") && (input.Rule.NotifyAdmin || len(userIDs) > 0) {
		requests = append(requests, NotificationRequest{
			IdempotencyKey: examPostStatIdempotencyKey(input, notificationmodel.ChannelInternal),
			SourceType:     "exam_record", SourceID: sourceID, Channel: notificationmodel.ChannelInternal,
			Recipient: notificationoutboxapp.InternalRecipient{NotifyAdmin: input.Rule.NotifyAdmin, UserIDs: userIDs},
			Payload: notificationoutboxapp.MessagePayload{
				Title: "考试统计通知: " + input.ExamTitle, Content: input.Message,
				SourceType: "exam", SourceID: strconv.FormatUint(uint64(input.ExamID), 10),
			},
		})
	}
	return requests
}

func examPostStatIdempotencyKey(input examRuleNotificationInput, channel string) string {
	ruleID := strings.TrimSpace(input.Rule.ID)
	if ruleID == "" {
		ruleID = fmt.Sprintf("index-%d", input.RuleIndex)
	}
	raw := fmt.Sprintf("exam:%d:record:%d:rule:%s:channel:%s", input.ExamID, input.RecordID, ruleID, channel)
	sum := sha256.Sum256([]byte(raw))
	return "poststat:" + hex.EncodeToString(sum[:])
}

func resolveExamSubmitterContext(ctx context.Context, userID uint) string {
	if userID == 0 {
		return "匿名用户"
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.User
	if db.Select("id", "name").Where("`id` = ?", userID).First(&user).Error == nil && user.Name != "" {
		return user.Name
	}
	return fmt.Sprintf("用户%d", userID)
}
