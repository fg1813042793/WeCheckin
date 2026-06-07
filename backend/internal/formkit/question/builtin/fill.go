package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// MultiInputQuestion 多项填空
type MultiInputQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&MultiInputQuestion{BaseQuestion: question.BaseQuestion{T: "multiInput", Dis: "多项填空", Cat: "fill"}})
}

func (q *MultiInputQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案必须为字符串"}
	}
	return nil
}

func (q *MultiInputQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请输入",
		"fields": []map[string]interface{}{},
	}
}

// HInputQuestion 横向填空
type HInputQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&HInputQuestion{BaseQuestion: question.BaseQuestion{T: "hInput", Dis: "横向填空", Cat: "fill"}})
}

func (q *HInputQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案必须为字符串"}
	}
	return nil
}

func (q *HInputQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请输入",
		"fields":      []map[string]interface{}{},
	}
}

// ScanCodeQuestion 扫码题
type ScanCodeQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&ScanCodeQuestion{BaseQuestion: question.BaseQuestion{T: "scanCode", Dis: "扫码", Cat: "fill"}})
}

func (q *ScanCodeQuestion) Validate(value interface{}, sch schema.Question) error {
	if err := runSchemaRules(value, sch); err != nil {
		return err
	}
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案必须为字符串"}
	}
	return nil
}

func (q *ScanCodeQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "扫码"}
}
