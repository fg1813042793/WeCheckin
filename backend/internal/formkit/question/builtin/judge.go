package builtin

import (
	"fmt"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// JudgeQuestion 判断题（对/错，本质是固定选项的单选）
type JudgeQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&JudgeQuestion{BaseQuestion: question.BaseQuestion{T: "judge", Dis: "判断题", Cat: "select"}})
}

func (q *JudgeQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案类型错误"}
	}
	if s != "true" && s != "false" {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案不合法"}
	}
	return nil
}

func (q *JudgeQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	switch s {
	case "true":
		return "对"
	case "false":
		return "错"
	}
	return s
}

func (q *JudgeQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"options": []map[string]interface{}{
			{"label": "对", "value": "true"},
			{"label": "错", "value": "false"},
		},
	}
}

// NpsQuestion NPS 评分（0-10 分）
type NpsQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&NpsQuestion{BaseQuestion: question.BaseQuestion{T: "nps", Dis: "NPS评分", Cat: "base"}})
}

func (q *NpsQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	n, ok := toFloat64(value)
	if !ok || n < 0 || n > 10 {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "评分必须在 0-10 之间"}
	}
	return nil
}

func (q *NpsQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func (q *NpsQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"min":  0,
		"max":  10,
		"step": 1,
	}
}

// ---- helpers ----
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		var f float64
		if _, err := fmt.Sscanf(n, "%f", &f); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}
