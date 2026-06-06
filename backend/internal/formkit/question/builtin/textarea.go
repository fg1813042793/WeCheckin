package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

type TextareaQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&TextareaQuestion{BaseQuestion: question.BaseQuestion{T: "textarea", Dis: "多行文本", Cat: "base"}})
}

func (q *TextareaQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "多行文本题答案必须为字符串"}
	}
	return nil
}

func (q *TextareaQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入", "rows": 4}
}
