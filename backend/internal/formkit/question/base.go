package question

import (
	"wecheckin-backend/backend/internal/formkit/schema"
)

// BaseQuestion 提供 Question 接口的空实现，子类型只需覆盖需要的方法。
// 例如单行文本继承 BaseQuestion 后只需覆盖 Type/DisplayName/Validate。
type BaseQuestion struct {
	T   string // Type
	Dis string // DisplayName
	Cat string // Category
}

func (b *BaseQuestion) Type() string                       { return b.T }
func (b *BaseQuestion) DisplayName() string                { return b.Dis }
func (b *BaseQuestion) Category() string                   { return b.Cat }
func (b *BaseQuestion) Validate(_ interface{}, _ schema.Question) error { return nil }
func (b *BaseQuestion) ExtractValue(data map[string]interface{}, q schema.Question) interface{} {
	if data == nil {
		return nil
	}
	return data[q.ID]
}
func (b *BaseQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}
func (b *BaseQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{}
}

// validateSchemaRules 跑 schema.Validate 规则（maxLen/minLen/pattern/required）。
// 这部分对所有题型通用，所以 BaseQuestion.Validate 在子类未覆盖时直接调用它。
// 子类可以追加类型特定检查。
func validateSchemaRules(value interface{}, q schema.Question) *ValidationError {
	valStr, _ := value.(string)
	// required
	if q.Required {
		if value == nil {
			return &ValidationError{QuestionID: q.ID, Field: q.ID, Message: "此项为必填"}
		}
		if s, ok := value.(string); ok && s == "" {
			return &ValidationError{QuestionID: q.ID, Field: q.ID, Message: "此项为必填"}
		}
	}
	// skip non-string rules for non-string values
	if valStr == "" {
		return nil
	}
	for _, rule := range q.Validate {
		switch rule.Type {
		case "maxLength":
			if n, ok := rule.Value.(float64); ok {
				if float64(len(valStr)) > n {
					return &ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		case "minLength":
			if n, ok := rule.Value.(float64); ok {
				if float64(len(valStr)) < n {
					return &ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		case "pattern":
			if pat, ok := rule.Value.(string); ok && pat != "" {
				if !matchPattern(pat, valStr) {
					return &ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		}
	}
	return nil
}
