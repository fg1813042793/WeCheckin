package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/app/formkit/schema"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
)

type PostStatRule struct {
	ID              string `json:"id"`
	Action          string `json:"action"`
	ConditionType   string `json:"conditionType"`
	StatField       string `json:"statField"`
	StatScope       string `json:"statScope"`
	NotifyChannel   string `json:"notifyChannel"`
	WebhookType     string `json:"webhookType"`
	WebhookURL      string `json:"webhookUrl"`
	NotifyAdmin     bool   `json:"notifyAdmin"`
	NotifyUserIds   string `json:"notifyUserIds"`
	MessageTemplate string `json:"messageTemplate"`
}

// ProcessPostStat 在问卷提交后处理 postStat 规则
func ProcessPostStat(surveyID uint, userID uint, nickname string, currentAnswers string) {
	sv, err := NewSurveyService().Get(surveyID)
	if err != nil {
		logger.Logger.Printf("[PostStat] survey not found: %d", surveyID)
		return
	}
	if sv.Settings == "" {
		return
	}
	rules := parsePostStatRules(sv.Settings)
	if len(rules) == 0 {
		return
	}

	var allResp []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ? AND `survey_resp_status` = 1", surveyID).Find(&allResp)

	items := make([]report.AnswerItem, len(allResp))
	for i, r := range allResp {
		items[i] = report.AnswerItem{Forms: r.Answers}
	}

	submitter := resolveSubmitter(sv.Schema, currentAnswers, nickname, userID)

	now := time.Now()
	dateStr := now.Format("2006-01-02 15:04:05")
	total := len(allResp)

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
			go sendInternalNotification(surveyID, sv.Title, rule.NotifyAdmin, rule.NotifyUserIds, msg)
		}
	}
}

// resolveSubmitter 从答卷中提取提交人：依次从个人信息题取值，其次用传入的 nickname，最后查用户表
func resolveSubmitter(schemaJSON, currentAnswers, nickname string, userID uint) string {
	sch, err := schema.Parse(schemaJSON)
	if err == nil && currentAnswers != "" {
		var ans map[string]interface{}
		if json.Unmarshal([]byte(currentAnswers), &ans) == nil {
			personalTypes := []string{"name", "phone", "email", "studentId", "employeeId"}
			for _, pt := range personalTypes {
				for _, q := range sch.Questions {
					if q.Type == pt {
						if val, ok := ans[q.ID]; ok && val != nil {
							if s, ok := val.(string); ok && s != "" {
								return s
							}
						}
					}
				}
			}
		}
	}
	if nickname != "" {
		return nickname
	}
	if userID > 0 {
		var u model.User
		if database.DB.Where("`id` = ?", userID).First(&u).Error == nil && u.Name != "" {
			return u.Name
		}
	}
	return "匿名用户"
}

func parsePostStatRules(settings string) []PostStatRule {
	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &settingsMap); err != nil {
		return nil
	}
	raw, ok := settingsMap["logicRules"]
	if !ok {
		return nil
	}
	var rawBytes []byte
	switch v := raw.(type) {
	case string:
		rawBytes = []byte(v)
	default:
		rawBytes, _ = json.Marshal(v)
	}
	var allRules []PostStatRule
	if err := json.Unmarshal(rawBytes, &allRules); err != nil {
		logger.Logger.Printf("[PostStat] parse logicRules error: %v", err)
		return nil
	}
	var out []PostStatRule
	for _, r := range allRules {
		if r.Action == "postStat" {
			out = append(out, r)
		}
	}
	return out
}

func buildResultText(stats []report.FieldStat, statMode string) string {
	modeLabel := "选项标签"
	if statMode == "value" {
		modeLabel = "选项值"
	}
	// 聚合所有题目的选项值分布
	type kv struct{ k string; v int }
	agg := map[string]int{}
	for _, s := range stats {
		if s.Type == "divider" || s.Type == "description" {
			continue
		}
		for k, v := range s.Dist {
			agg[k] += v
		}
	}
	if len(agg) == 0 {
		return ""
	}
	var sorted []kv
	total := 0
	for k, v := range agg {
		sorted = append(sorted, kv{k, v})
		total += v
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	lines := []string{fmt.Sprintf("统计类型（%s）", modeLabel)}
	for _, p := range sorted {
		pct := float64(p.v) / float64(total) * 100
		lines = append(lines, fmt.Sprintf("%s 个数 %d 比重 %.1f%%", p.k, p.v, pct))
	}
	return strings.Join(lines, "\n")
}

func sendWebhook(webhookType, webhookURL, title, msg string) {
	// 各平台的 markdown 渲染对换行要求不同，统一用双换行保证视觉换行
	msg = strings.ReplaceAll(msg, "\n", "\n\n")
	client := &http.Client{Timeout: 10 * time.Second}
	var body []byte

	switch webhookType {
	case "dingtalk":
		payload := map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": title,
				"text":  msg,
			},
		}
		body, _ = json.Marshal(payload)
	case "wecom":
		payload := map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": msg,
			},
		}
		body, _ = json.Marshal(payload)
	case "lark":
		payload := map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"header": map[string]interface{}{
					"title": map[string]string{"tag": "plain_text", "content": title},
					"template": "blue",
				},
				"elements": []map[string]interface{}{
					{"tag": "markdown", "content": msg},
				},
			},
		}
		body, _ = json.Marshal(payload)
	default:
		payload := map[string]interface{}{
			"title":   title,
			"content": msg,
			"type":    "survey_stat",
		}
		body, _ = json.Marshal(payload)
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Logger.Printf("[PostStat] webhook send error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		logger.Logger.Printf("[PostStat] webhook status %d for %s", resp.StatusCode, webhookURL)
	}
}

func sendInternalNotification(surveyID uint, title string, notifyAdmin bool, notifyUserIds string, msg string) {
	surveyIdStr := strconv.FormatUint(uint64(surveyID), 10)
	if notifyAdmin {
		notify := model.Notify{
			Title:      "问卷统计通知: " + title,
			Content:    msg,
			Type:       "survey_stat",
			SourceID:   surveyIdStr,
			SourceType: "survey",
			UserID:     "",
			AddTime:    time.Now().UnixMilli(),
		}
		if err := database.DB.Create(&notify).Error; err != nil {
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
			SourceID:   surveyIdStr,
			SourceType: "survey",
			UserID:     id,
			AddTime:    time.Now().UnixMilli(),
		}
		if err := database.DB.Create(&notify).Error; err != nil {
			logger.Logger.Printf("[PostStat] create notify error: %v", err)
		}
	}
}
