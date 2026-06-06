package builtin

import (
	"strconv"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// SwitchQuestion 开关题（type=switch），值 "1" 或 "0"
type SwitchQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&SwitchQuestion{BaseQuestion: question.BaseQuestion{T: "switch", Dis: "开关", Cat: "base"}})
}

func (q *SwitchQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	switch v := value.(type) {
	case string:
		if v == "0" || v == "1" {
			return nil
		}
	case bool:
		return nil
	}
	return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "开关题答案必须为 0/1 或布尔"}
}

func (q *SwitchQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		if v == "1" {
			return "是"
		}
		return "否"
	case bool:
		if v {
			return "是"
		}
		return "否"
	}
	return ""
}

func (q *SwitchQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"onText": "是", "offText": "否"}
}

// RatingQuestion 评分题（type=rating），值 1~max
type RatingQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&RatingQuestion{BaseQuestion: question.BaseQuestion{T: "rating", Dis: "评分", Cat: "base"}})
}

func (q *RatingQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请评分"}
		}
		return nil
	}
	props := sch.PropsMap()
	maxN := 5
	if m, ok := props["max"].(float64); ok && m > 0 {
		maxN = int(m)
	}
	var n int
	switch v := value.(type) {
	case float64:
		n = int(v)
	case int:
		n = v
	case int64:
		n = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "评分必须为整数"}
		}
		n = parsed
	default:
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "评分必须为数字"}
	}
	if n < 1 || n > maxN {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "评分超出范围"}
	}
	return nil
}

func (q *RatingQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case float64:
		return strconv.Itoa(int(v))
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	}
	return ""
}

func (q *RatingQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"max": 5, "icon": "star"}
}
