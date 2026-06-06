package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

type InputQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&InputQuestion{BaseQuestion: question.BaseQuestion{T: "input", Dis: "单行文本", Cat: "base"}})
}

func (q *InputQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "文本题答案必须为字符串"}
	}
	return nil
}

func (q *InputQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入"}
}

type TextQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&TextQuestion{BaseQuestion: question.BaseQuestion{T: "text", Dis: "文本（input 别名）", Cat: "base"}})
}

func (q *TextQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	return nil
}

func (q *TextQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入"}
}
