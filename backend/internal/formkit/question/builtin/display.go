package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// DividerQuestion 分隔线（type=divider），仅展示无答案
type DividerQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&DividerQuestion{BaseQuestion: question.BaseQuestion{T: "divider", Dis: "分割线", Cat: "layout"}})
}

func (q *DividerQuestion) Validate(_ interface{}, _ schema.Question) error { return nil }
func (q *DividerQuestion) ExtractValue(_ map[string]interface{}, _ schema.Question) interface{} {
	return nil
}
func (q *DividerQuestion) FormatValue(_ interface{}, _ schema.Question) string { return "" }

func (q *DividerQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"color": "#eee", "text": ""}
}

// DescriptionQuestion 说明文字（type=description），仅展示
type DescriptionQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&DescriptionQuestion{BaseQuestion: question.BaseQuestion{T: "description", Dis: "说明文字", Cat: "layout"}})
}

func (q *DescriptionQuestion) Validate(_ interface{}, _ schema.Question) error { return nil }
func (q *DescriptionQuestion) ExtractValue(_ map[string]interface{}, _ schema.Question) interface{} {
	return nil
}
func (q *DescriptionQuestion) FormatValue(_ interface{}, _ schema.Question) string { return "" }

func (q *DescriptionQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"text": ""}
}
