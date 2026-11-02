package exam

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

// SchemaGradingPlan contains only questions that participate in grading while
// retaining their original schema ordinal for result-to-question mapping.
type SchemaGradingPlan struct {
	Questions   []examPkg.Question
	IDMap       map[string]uint
	Definitions map[uint]json.RawMessage
}

var schemaNonGradingTypes = map[string]struct{}{
	"description": {}, "divider": {}, "pagination": {}, "questionSet": {},
	"name": {}, "phone": {}, "email": {}, "studentId": {}, "employeeId": {}, "class": {}, "idCard": {}, "password": {},
}

// ShouldGradePaperQuestion keeps collection-only and zero-score paper fields out of grading.
func ShouldGradePaperQuestion(questionType string, score int) bool {
	if _, excluded := schemaNonGradingTypes[strings.TrimSpace(questionType)]; excluded {
		return false
	}
	return score > 0
}

// BuildSchemaGradingPlan converts designer schema questions into the grading engine contract.
func BuildSchemaGradingPlan(schemaJSON string) SchemaGradingPlan {
	plan := SchemaGradingPlan{Questions: make([]examPkg.Question, 0), IDMap: make(map[string]uint), Definitions: make(map[uint]json.RawMessage)}
	var schema struct {
		Questions []map[string]interface{} `json:"questions"`
	}
	if json.Unmarshal([]byte(schemaJSON), &schema) != nil {
		return plan
	}
	for index, raw := range schema.Questions {
		id := schemaStringValue(raw["id"])
		typ := schemaStringValue(raw["type"])
		if id == "" || typ == "" {
			continue
		}
		if _, excluded := schemaNonGradingTypes[typ]; excluded {
			continue
		}
		scoreMode := schemaStringValue(raw["examScoreMode"])
		score := schemaScoreValue(raw["examScore"])
		fields := schemaAnswerFields(raw)
		if scoreMode == "" && score <= 0 && !hasScoredSchemaFields(fields) {
			continue
		}
		ordinal := uint(index + 1)
		plan.IDMap[id] = ordinal
		plan.Definitions[ordinal], _ = json.Marshal(raw)
		plan.Questions = append(plan.Questions, examPkg.Question{
			ID:         ordinal,
			Type:       typ,
			Title:      schemaStringValue(raw["title"]),
			Answer:     schemaStringValue(raw["examCorrectAnswer"]),
			Score:      score,
			ScoreMode:  scoreMode,
			Fields:     fields,
			NeedManual: examPkg.QWithType(typ),
		})
	}
	return plan
}

// ApplySchemaSubmissionOutcome stores score, grading state and pass result.
func ApplySchemaSubmissionOutcome(record *model.ExamRecord, result examPkg.PaperResult, passScore int) {
	if record == nil {
		return
	}
	record.Score = result.TotalScore
	record.TotalScore = result.FullScore
	record.Pass = 0
	if result.ManualCount > 0 {
		record.Status = 1
		return
	}
	record.Status = 2
	if result.TotalScore >= passScore {
		record.Pass = 1
	}
}

func schemaAnswerFields(raw map[string]interface{}) []examPkg.AnswerField {
	props, _ := raw["props"].(map[string]interface{})
	rawFields, _ := props["fields"].([]interface{})
	fields := make([]examPkg.AnswerField, 0, len(rawFields))
	for _, item := range rawFields {
		field, _ := item.(map[string]interface{})
		answer := schemaStringValue(field["examCorrectAnswer"])
		score := schemaScoreValue(field["examScore"])
		required, _ := field["examCorrect"].(bool)
		fields = append(fields, examPkg.AnswerField{Answer: answer, Score: score, Required: required})
	}
	return fields
}

func hasScoredSchemaFields(fields []examPkg.AnswerField) bool {
	for _, field := range fields {
		if field.Score > 0 || field.Required || strings.TrimSpace(field.Answer) != "" {
			return true
		}
	}
	return false
}

func schemaScoreValue(value interface{}) int {
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

func schemaStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}
