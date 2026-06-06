package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// runSchemaRules 跑 schema 里的 Validate 规则（maxLength/minLength/pattern/required）。
// 内部暴露给 builtin 包内各题型调用，避免循环 import。
func runSchemaRules(value interface{}, q schema.Question) *question.ValidationError {
	valStr, _ := value.(string)
	if q.Required {
		if value == nil {
			return &question.ValidationError{QuestionID: q.ID, Field: q.ID, Message: "此项为必填"}
		}
		if s, ok := value.(string); ok && s == "" {
			return &question.ValidationError{QuestionID: q.ID, Field: q.ID, Message: "此项为必填"}
		}
	}
	if valStr == "" {
		return nil
	}
	for _, rule := range q.Validate {
		switch rule.Type {
		case "maxLength":
			if n, ok := rule.Value.(float64); ok {
				if float64(len(valStr)) > n {
					return &question.ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		case "minLength":
			if n, ok := rule.Value.(float64); ok {
				if float64(len(valStr)) < n {
					return &question.ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		case "pattern":
			if pat, ok := rule.Value.(string); ok && pat != "" {
				if !matchPattern(pat, valStr) {
					return &question.ValidationError{QuestionID: q.ID, Field: q.ID, Message: rule.Message}
				}
			}
		}
	}
	return nil
}

// matchPattern 用 regexp 校验；定义在本包内供 builtin 题型使用。
func matchPattern(pattern, s string) bool {
	// 延迟编译避免每次调用都重新编译；这里简单起见直接编译
	re, err := compileRegexp(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}
