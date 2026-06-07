package builtin

import (
	"fmt"
	"strings"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

type CascadeQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&CascadeQuestion{BaseQuestion: question.BaseQuestion{T: "cascade", Dis: "级联选择", Cat: "select"}})
}

func (q *CascadeQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" && sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
	case []interface{}:
		if len(v) == 0 && sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请至少选择一项"}
		}
	default:
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案类型错误"}
	}
	return nil
}

func (q *CascadeQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			s, _ := item.(string)
			if s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " / ")
	}
	return fmt.Sprintf("%v", value)
}

func (q *CascadeQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请选择",
		"options":     []map[string]interface{}{},
	}
}
