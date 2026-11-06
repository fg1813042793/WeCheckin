package exam

import (
	"encoding/json"
	"sort"
	"strconv"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

type reviewQuestionSnapshot struct {
	QuestionID      uint            `json:"questionId"`
	AnswerKey       string          `json:"answerKey"`
	Title           string          `json:"title"`
	Type            string          `json:"type"`
	ReferenceAnswer string          `json:"referenceAnswer"`
	Definition      json.RawMessage `json:"definition,omitempty"`
}

type recordReviewSnapshot struct {
	Version   int                      `json:"version"`
	PassScore int                      `json:"passScore"`
	Questions []reviewQuestionSnapshot `json:"questions"`
}

func BuildSchemaReviewSnapshot(plan SchemaGradingPlan, passScore int) string {
	raw, _ := json.Marshal(recordReviewSnapshot{Version: 2, PassScore: passScore, Questions: plan.ReviewQuestions})
	return string(raw)
}

func BuildPaperReviewSnapshot(questions []model.ExamQuestion, passScore int) string {
	keys := make(map[uint]string, len(questions))
	definitions := make(map[uint]json.RawMessage, len(questions))
	items := make([]examPkg.Question, 0, len(questions))
	for _, question := range questions {
		keys[question.ID] = strconv.FormatUint(uint64(question.ID), 10)
		items = append(items, examPkg.Question{ID: question.ID, Title: question.Title, Type: question.Type, Answer: question.Answer})
		var definition map[string]interface{}
		_ = json.Unmarshal([]byte(question.Schema), &definition)
		if definition == nil {
			definition = make(map[string]interface{})
		}
		if definition["title"] == nil {
			definition["title"] = question.Title
		}
		if definition["type"] == nil {
			definition["type"] = question.Type
		}
		if definition["props"] == nil {
			var options interface{}
			if json.Unmarshal([]byte(question.Options), &options) == nil {
				definition["props"] = map[string]interface{}{"options": options}
			}
		}
		definitions[question.ID], _ = json.Marshal(definition)
	}
	return encodeReviewSnapshot(items, keys, passScore, definitions)
}

func encodeReviewSnapshot(questions []examPkg.Question, keys map[uint]string, passScore int, definitions map[uint]json.RawMessage) string {
	snapshot := recordReviewSnapshot{Version: 2, PassScore: passScore, Questions: make([]reviewQuestionSnapshot, 0, len(questions))}
	for _, question := range questions {
		snapshot.Questions = append(snapshot.Questions, reviewQuestionSnapshot{QuestionID: question.ID, AnswerKey: keys[question.ID], Title: stripExamExportHTML(question.Title), Type: question.Type, ReferenceAnswer: question.Answer, Definition: definitions[question.ID]})
	}
	sort.Slice(snapshot.Questions, func(i, j int) bool { return snapshot.Questions[i].QuestionID < snapshot.Questions[j].QuestionID })
	raw, _ := json.Marshal(snapshot)
	return string(raw)
}

func decodeRecordReviewSnapshot(raw string) (recordReviewSnapshot, error) {
	var snapshot recordReviewSnapshot
	if json.Unmarshal([]byte(raw), &snapshot) != nil || (snapshot.Version != 1 && snapshot.Version != 2) || snapshot.PassScore < 0 {
		return snapshot, ErrRecordGradingUnavailable
	}
	seen := make(map[uint]bool, len(snapshot.Questions))
	keys := make(map[string]bool, len(snapshot.Questions))
	for _, question := range snapshot.Questions {
		if question.QuestionID == 0 || question.AnswerKey == "" || seen[question.QuestionID] || keys[question.AnswerKey] {
			return snapshot, ErrRecordGradingUnavailable
		}
		seen[question.QuestionID], keys[question.AnswerKey] = true, true
		if len(question.Definition) > 0 {
			var definition map[string]interface{}
			if json.Unmarshal(question.Definition, &definition) != nil || definition == nil {
				return snapshot, ErrRecordGradingUnavailable
			}
		}
	}
	return snapshot, nil
}
