package exam

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"wecheckin/backend/internal/model"
)

func buildRecordGradingDetail(record model.ExamRecord, exam model.Exam, snapshot recordReviewSnapshot, notice string) (*RecordGradingDetail, error) {
	normalizeRecordDeviceMetadata(&record)
	detail := &RecordGradingDetail{Record: record, ExamTitle: exam.Title, PassScore: snapshot.PassScore, Notice: notice, Questions: []GradingQuestion{}}
	result, err := decodeGradingPaperResult(record.Result)
	if err != nil {
		detail.Notice = ErrRecordGradingUnavailable.Error()
		return detail, nil
	}
	var answers map[string]interface{}
	if strings.TrimSpace(record.Answers) != "" && json.Unmarshal([]byte(record.Answers), &answers) != nil {
		detail.Notice = ErrRecordGradingUnavailable.Error()
		return detail, nil
	}
	snapshot, detail.Notice = supplementLegacyUnscoredQuestions(record, exam, snapshot, answers, result.Results, detail.Notice)
	versionData, _ := json.Marshal(struct {
		ID, ExamID                uint
		Status, Score, TotalScore int
		Answers, Result, Snapshot string
		Review                    recordReviewSnapshot
	}{record.ID, record.ExamID, record.Status, record.Score, record.TotalScore, record.Answers, record.Result, record.ReviewSnapshot, snapshot})
	hash := sha256.Sum256(versionData)
	detail.Version = hex.EncodeToString(hash[:])
	byID := make(map[uint]reviewQuestionSnapshot, len(snapshot.Questions))
	for _, question := range snapshot.Questions {
		byID[question.QuestionID] = question
	}
	seen := make(map[uint]bool, len(result.Results))
	manualCount, fullScore := 0, 0
	available := true
	for _, grade := range result.Results {
		if grade.QuestionID == 0 || seen[grade.QuestionID] || grade.FullScore < 0 || grade.GotScore < 0 || grade.GotScore > grade.FullScore || grade.FullScore > math.MaxInt32-fullScore {
			return nil, ErrRecordGradingUnavailable
		}
		seen[grade.QuestionID] = true
		fullScore += grade.FullScore
		question, ok := byID[grade.QuestionID]
		item := GradingQuestion{QuestionID: grade.QuestionID, Title: fmt.Sprintf("题目 #%d", grade.QuestionID), Answer: "未作答", ReferenceAnswer: "-", FullScore: grade.FullScore, GotScore: grade.GotScore, NeedManual: grade.NeedManual}
		if ok {
			var reference string
			var materialAvailable bool
			item.Materials, item.Images, reference, materialAvailable = gradingMaterials(question)
			available = available && materialAvailable
			item.Title, item.Type = stripExamExportHTML(question.Title), question.Type
			if item.Title == "" {
				item.Title = fmt.Sprintf("题目 #%d", grade.QuestionID)
			}
			if value, exists := answers[question.AnswerKey]; exists {
				item.Answer = gradingQuestionAnswer(value, question)
			}
			if strings.TrimSpace(reference) != "" {
				item.ReferenceAnswer = reference
			}
		} else {
			available = false
			item.Answer = "题目资料缺失，无法匹配答案"
		}
		if grade.NeedManual {
			manualCount++
		}
		detail.Questions = append(detail.Questions, item)
	}
	for _, question := range snapshot.Questions {
		if !seen[question.QuestionID] {
			detail.Questions = append(detail.Questions, unscoredGradingQuestion(question, answers))
		}
	}
	sort.SliceStable(detail.Questions, func(i, j int) bool { return detail.Questions[i].QuestionID < detail.Questions[j].QuestionID })
	detail.CanGrade = record.Status == 1 && available && manualCount > 0
	if !available {
		detail.Notice = "部分题目资料已缺失，无法安全批改此答卷"
	}
	if record.Status == 0 {
		detail.Notice = "答卷尚未提交，不能批改"
	}
	if record.Status == 1 && manualCount == 0 {
		detail.Notice = "答卷没有可人工批改的题目"
	}
	return detail, nil
}

func gradingDisplayValue(value interface{}) string {
	if value == nil {
		return "未作答"
	}
	if text, ok := value.(string); ok {
		if strings.TrimSpace(text) == "" {
			return "未作答"
		}
		return text
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "无法展示答案"
	}
	return string(raw)
}

func applyManualGrading(record model.ExamRecord, detail *RecordGradingDetail, request GradeRecordRequest, adminID uint, now int64) (model.ExamRecord, error) {
	if record.Status != 1 {
		return record, ErrRecordNotGradable
	}
	if detail == nil || !detail.CanGrade {
		return record, ErrRecordGradingUnavailable
	}
	if strings.TrimSpace(request.ExpectedVersion) != detail.Version {
		return record, ErrRecordGradingChanged
	}
	comment := strings.TrimSpace(request.Comment)
	if adminID == 0 || utf8.RuneCountInString(comment) > 1000 {
		return record, ErrRecordGradingInvalid
	}
	result, err := decodeGradingPaperResult(record.Result)
	if err != nil {
		return record, err
	}
	pending := make(map[uint]int)
	for i, grade := range result.Results {
		if grade.NeedManual {
			pending[grade.QuestionID] = i
		}
	}
	if len(request.Scores) != len(pending) || len(pending) == 0 {
		return record, ErrRecordGradingInvalid
	}
	seen := make(map[uint]bool, len(request.Scores))
	for _, score := range request.Scores {
		index, exists := pending[score.QuestionID]
		if !exists || seen[score.QuestionID] || score.Score == nil || *score.Score < 0 || *score.Score > result.Results[index].FullScore {
			return record, ErrRecordGradingInvalid
		}
		seen[score.QuestionID] = true
		grade := &result.Results[index]
		grade.GotScore, grade.NeedManual = *score.Score, false
		grade.Correct = grade.FullScore > 0 && grade.GotScore == grade.FullScore
		grade.Reason = fmt.Sprintf("人工评分 %d/%d", grade.GotScore, grade.FullScore)
	}
	result.TotalScore, result.FullScore, result.ManualCount, result.CorrectCnt = 0, 0, 0, 0
	for _, grade := range result.Results {
		if grade.FullScore < 0 || grade.GotScore < 0 || grade.GotScore > grade.FullScore || grade.FullScore > math.MaxInt32-result.FullScore {
			return record, ErrRecordGradingUnavailable
		}
		result.TotalScore += grade.GotScore
		result.FullScore += grade.FullScore
		if grade.Correct {
			result.CorrectCnt++
		}
	}
	var raw []byte
	if strings.HasPrefix(strings.TrimSpace(record.Result), "[") {
		raw, err = json.Marshal(result.Results)
	} else {
		raw, err = json.Marshal(result)
	}
	if err != nil {
		return record, ErrRecordGradingUnavailable
	}
	record.Result, record.Score, record.TotalScore = string(raw), result.TotalScore, result.FullScore
	record.Status, record.Pass = 2, 0
	if record.Score >= detail.PassScore {
		record.Pass = 1
	}
	record.GradedBy, record.GradedAt, record.GradingComment = adminID, now, comment
	return record, nil
}
