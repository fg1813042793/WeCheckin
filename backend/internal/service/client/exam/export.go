package exam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
	"wecheckin/backend/pkg/database"

	"gorm.io/gorm"
)

var examExportHTMLTagPattern = regexp.MustCompile(`<[^>]*>`)
var ErrRecordExportUnavailable = errors.New("exam record export is unavailable")

type storedQuestionResult struct {
	QuestionID uint   `json:"questionId"`
	Correct    bool   `json:"correct"`
	GotScore   int    `json:"gotScore"`
	FullScore  int    `json:"fullScore"`
	NeedManual bool   `json:"needManual"`
	Reason     string `json:"reason"`
}

type storedPaperResult struct {
	Results []storedQuestionResult `json:"results"`
}

// ExportRecordForAdminContext renders one visible exam record in the requested format.
func (s *Service) ExportRecordForAdminContext(ctx context.Context, examID, recordID, adminID uint, format workflowsummary.ExportFormat) (*workflowsummary.ExportResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	examQuery, err := scopedExamQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	return exportRecordContext(db, examQuery, examID, recordID, format)
}

// ExportRecordForLinkContext renders the exact record selected by an already validated public grant.
func (s *Service) ExportRecordForLinkContext(ctx context.Context, examID, recordID uint, format workflowsummary.ExportFormat) (*workflowsummary.ExportResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	result, err := exportRecordContext(db, db.Model(&model.Exam{}), examID, recordID, format)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordExportUnavailable
	}
	return result, err
}

func exportRecordContext(db, examQuery *gorm.DB, examID, recordID uint, format workflowsummary.ExportFormat) (*workflowsummary.ExportResult, error) {
	var exam model.Exam
	if err := examQuery.Where("`exam_id` = ?", examID).First(&exam).Error; err != nil {
		return nil, err
	}
	var record model.ExamRecord
	if err := db.Where("`exam_r_id` = ? AND `exam_r_exam_id` = ?", recordID, examID).First(&record).Error; err != nil {
		return nil, err
	}
	paperID := record.PaperID
	if paperID == 0 {
		paperID = exam.PaperID
	}
	if paperID > 0 {
		var paper model.ExamPaper
		if err := db.Where("`exam_p_id` = ?", paperID).First(&paper).Error; err != nil {
			return nil, err
		}
		questions, err := examPaperQuestions(db, paper.QuestionIDs)
		if err != nil {
			return nil, err
		}
		var orderedIDs []uint
		_ = json.Unmarshal([]byte(paper.QuestionIDs), &orderedIDs)
		byID := make(map[uint]model.ExamQuestion, len(questions))
		for _, question := range questions {
			byID[question.ID] = question
		}
		ordered := make([]model.ExamQuestion, 0, len(questions))
		for _, questionID := range orderedIDs {
			if question, ok := byID[questionID]; ok {
				ordered = append(ordered, question)
			}
		}
		exam.Schema = examSchemaFromPaperQuestions(ordered)
	}

	userName := ""
	if userID, parseErr := strconv.ParseUint(strings.TrimSpace(record.UserID), 10, 64); parseErr == nil && userID > 0 {
		var user model.User
		if err := db.Select("id", "user_name").Where("id = ?", uint(userID)).First(&user).Error; err == nil {
			userName = strings.TrimSpace(user.Name)
		}
	}

	document := buildRecordExportDocument(exam, record, userName)
	return workflowsummary.RenderDocuments([]workflowsummary.Document{document}, format)
}

func buildRecordExportDocument(exam model.Exam, record model.ExamRecord, userName string) workflowsummary.Document {
	displayUser := strings.TrimSpace(userName)
	if displayUser == "" {
		displayUser = firstExamExportText(record.UserID, "匿名用户")
	}
	document := workflowsummary.Document{
		Layout:   workflowsummary.DocumentLayoutAnswerSheet,
		Title:    firstExamExportText(exam.Title, "考试答卷"),
		FileStem: fmt.Sprintf("exam-%d-record-%d", exam.ID, record.ID),
		Metadata: [][]string{
			{"考生", displayUser},
			{"总成绩", fmt.Sprintf("%d/%d", record.Score, record.TotalScore)},
			{"答卷状态", examRecordStatusLabel(record.Status)},
			{"是否通过", examRecordPassLabel(record.Pass)},
			{"考试用时", formatExamExportDuration(record.TimeSpent)},
			{"提交时间", formatExamExportTime(record.SubmitTime)},
			{"开始时间", formatExamExportTime(record.StartTime)},
			{"答卷编号", strconv.FormatUint(uint64(record.ID), 10)},
			{"考试编号", strconv.FormatUint(uint64(exam.ID), 10)},
			{"用户 ID", firstExamExportText(record.UserID, "匿名")},
		},
	}

	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	_ = json.Unmarshal([]byte(exam.Schema), &schema)
	answers := make(map[string]interface{})
	_ = json.Unmarshal([]byte(record.Answers), &answers)
	results, legacyResults := decodeStoredExamResults(record.Result)
	gradeIDs := make(map[string]int)
	snapshot, snapshotErr := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	if snapshotErr == nil {
		for _, question := range snapshot.Questions {
			gradeIDs[question.AnswerKey] = int(question.QuestionID)
		}
	}
	paperMode := record.PaperID > 0 || exam.PaperID > 0

	sectionTitle := "答题明细"
	rows := make([][]string, 0, len(schema.Questions))
	flushRows := func() {
		if len(rows) == 0 {
			return
		}
		blocks := make([]workflowsummary.AnswerBlock, 0, len(rows))
		for _, row := range rows {
			blocks = append(blocks, workflowsummary.AnswerBlock{Title: row[0], Answer: row[1], ReferenceAnswer: row[2], Score: row[3], Result: row[4]})
		}
		document.Sections = append(document.Sections, workflowsummary.Section{
			Title:        sectionTitle,
			Headers:      []string{"题目", "用户答案", "参考答案", "得分", "判题结果"},
			Rows:         rows,
			AnswerBlocks: blocks,
		})
		rows = nil
	}

	questionNumber := 0
	for schemaIndex, question := range schema.Questions {
		questionType := strings.TrimSpace(fmt.Sprint(question["type"]))
		if questionType == "divider" {
			flushRows()
			sectionTitle = firstExamExportText(stripExamExportHTML(fmt.Sprint(question["title"])), "答题明细")
			continue
		}
		if isExamExportLayoutType(questionType) {
			continue
		}
		questionNumber++
		questionID := strings.TrimSpace(fmt.Sprint(question["id"]))
		gradeKey := schemaIndex + 1
		if snapshotErr == nil {
			gradeKey = gradeIDs[questionID]
		} else if paperMode {
			gradeKey, _ = strconv.Atoi(questionID)
		}
		grade, hasGrade := results[gradeKey]
		if !hasGrade {
			if correct, ok := legacyResults[questionID]; ok {
				grade = storedQuestionResult{QuestionID: uint(questionNumber), Correct: correct, FullScore: intValue(question["examScore"])}
				if correct {
					grade.GotScore = grade.FullScore
				}
				hasGrade = true
			}
		}
		fullScore := intValue(question["examScore"])
		if hasGrade {
			fullScore = grade.FullScore
		}
		unscored := !hasGrade && schemaStringValue(question["examScoreMode"]) == examPkg.ScoreModeNone
		scoreText := "-"
		if unscored {
			scoreText = "不计分"
		} else if hasGrade && grade.NeedManual {
			scoreText = "待评分"
			if fullScore > 0 {
				scoreText = fmt.Sprintf("待评分 / %d", fullScore)
			}
		} else if hasGrade && fullScore == 0 {
			scoreText = "不计分"
		} else if hasGrade {
			scoreText = fmt.Sprintf("%d/%d", grade.GotScore, fullScore)
		} else if fullScore > 0 {
			scoreText = fmt.Sprintf("-/%d", fullScore)
		}
		referenceText, gradeText := formatExamReferenceValue(question, question["examCorrectAnswer"]), formatExamGrade(grade, hasGrade)
		if unscored {
			referenceText, gradeText = "不适用（不计分）", "不计分"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d. %s", questionNumber, firstExamExportText(stripExamExportHTML(fmt.Sprint(question["title"])), "未命名题目")),
			formatExamQuestionValue(question, answers[questionID]),
			referenceText,
			scoreText,
			gradeText,
		})
	}
	flushRows()
	if len(document.Sections) == 0 {
		document.Sections = append(document.Sections, workflowsummary.Section{Title: "答题明细", Headers: []string{"内容"}, Rows: [][]string{{"暂无可导出的答题内容"}}})
	}
	return document
}

func examSchemaFromPaperQuestions(questions []model.ExamQuestion) string {
	items := make([]map[string]interface{}, 0, len(questions))
	for _, question := range questions {
		var options interface{}
		if strings.TrimSpace(question.Options) != "" {
			_ = json.Unmarshal([]byte(question.Options), &options)
		}
		items = append(items, map[string]interface{}{
			"id":                strconv.FormatUint(uint64(question.ID), 10),
			"type":              question.Type,
			"title":             question.Title,
			"examCorrectAnswer": question.Answer,
			"examScore":         question.Score,
			"examScoreMode":     paperQuestionScoreMode(question),
			"props":             map[string]interface{}{"options": options},
		})
	}
	data, _ := json.Marshal(map[string]interface{}{"questions": items})
	return string(data)
}

func decodeStoredExamResults(raw string) (map[int]storedQuestionResult, map[string]bool) {
	results := make(map[int]storedQuestionResult)
	legacy := make(map[string]bool)
	if strings.TrimSpace(raw) == "" {
		return results, legacy
	}
	var paper storedPaperResult
	if json.Unmarshal([]byte(raw), &paper) == nil && len(paper.Results) > 0 {
		for _, item := range paper.Results {
			results[int(item.QuestionID)] = item
		}
		return results, legacy
	}
	var items []storedQuestionResult
	if json.Unmarshal([]byte(raw), &items) == nil && len(items) > 0 {
		for _, item := range items {
			results[int(item.QuestionID)] = item
		}
		return results, legacy
	}
	_ = json.Unmarshal([]byte(raw), &legacy)
	return results, legacy
}

func formatExamQuestionValue(question map[string]interface{}, value interface{}) string {
	if value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" {
		return "未作答"
	}
	labels := examQuestionOptionLabels(question)
	formatOne := func(item interface{}) string {
		text := strings.TrimSpace(fmt.Sprint(item))
		if label := labels[text]; label != "" {
			return label
		}
		return text
	}
	switch typed := value.(type) {
	case []interface{}:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, formatOne(item))
		}
		return strings.Join(parts, "、")
	case string:
		trimmed := strings.TrimSpace(typed)
		if json.Valid([]byte(trimmed)) {
			var decoded interface{}
			if json.Unmarshal([]byte(trimmed), &decoded) == nil {
				if decodedText, ok := decoded.(string); !ok || decodedText != typed {
					return formatExamQuestionValue(question, decoded)
				}
			}
		}
		var array []interface{}
		if strings.HasPrefix(trimmed, "[") && json.Unmarshal([]byte(typed), &array) == nil {
			return formatExamQuestionValue(question, array)
		}
		if len(labels) > 0 && strings.Contains(typed, ",") {
			parts := strings.Split(typed, ",")
			for index := range parts {
				parts[index] = formatOne(parts[index])
			}
			return strings.Join(parts, "、")
		}
		return formatOne(typed)
	default:
		if data, err := json.Marshal(value); err == nil && (strings.HasPrefix(string(data), "{") || strings.HasPrefix(string(data), "[")) {
			return string(data)
		}
		return formatOne(value)
	}
}

func formatExamReferenceValue(question map[string]interface{}, value interface{}) string {
	if emptyExamReferenceValue(value) {
		return "未设置参考答案"
	}
	return formatExamQuestionValue(question, value)
}

func emptyExamReferenceValue(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return true
		}
		var decoded interface{}
		if json.Unmarshal([]byte(trimmed), &decoded) == nil {
			return emptyExamReferenceValue(decoded)
		}
	case []interface{}:
		return len(typed) == 0
	case map[string]interface{}:
		return len(typed) == 0
	}
	return false
}

func examQuestionOptionLabels(question map[string]interface{}) map[string]string {
	result := make(map[string]string)
	props, _ := question["props"].(map[string]interface{})
	options, _ := props["options"].([]interface{})
	if len(options) == 0 {
		options, _ = question["options"].([]interface{})
	}
	for _, raw := range options {
		option, _ := raw.(map[string]interface{})
		value := strings.TrimSpace(fmt.Sprint(option["value"]))
		label := strings.TrimSpace(fmt.Sprint(option["label"]))
		if value != "" && label != "" {
			result[value] = stripExamExportHTML(label)
		}
	}
	return result
}

func formatExamGrade(result storedQuestionResult, exists bool) string {
	if !exists {
		return "未判定"
	}
	if result.NeedManual {
		reason := strings.TrimSpace(result.Reason)
		if reason == "" || reason == "主观题，需人工判分" {
			return "待人工评分"
		}
		return "待人工评分：" + reason
	}
	if result.Correct {
		return "正确"
	}
	return firstExamExportText(result.Reason, "错误")
}

func isExamExportLayoutType(questionType string) bool {
	switch questionType {
	case "description", "pagination", "questionSet":
		return true
	default:
		return false
	}
}

func stripExamExportHTML(value string) string {
	value = examExportHTMLTagPattern.ReplaceAllString(value, "")
	return strings.TrimSpace(html.UnescapeString(value))
}

func intValue(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	case json.Number:
		result, _ := typed.Int64()
		return int(result)
	default:
		result, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
		return result
	}
}

func examRecordPassLabel(pass int) string {
	if pass == 1 {
		return "通过"
	}
	return "未通过"
}

func examRecordStatusLabel(status int) string {
	switch status {
	case 1:
		return "已提交"
	case 2:
		return "已批改"
	default:
		return "进行中"
	}
}

func formatExamExportDuration(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	minutes := seconds / 60
	remaining := seconds % 60
	if minutes == 0 {
		return fmt.Sprintf("%d 秒", remaining)
	}
	return fmt.Sprintf("%d 分 %d 秒", minutes, remaining)
}

func formatExamExportTime(milliseconds int64) string {
	if milliseconds <= 0 {
		return "-"
	}
	return time.UnixMilli(milliseconds).Format("2006-01-02 15:04:05")
}

func firstExamExportText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
