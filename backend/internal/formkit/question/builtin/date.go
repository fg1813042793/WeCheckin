package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// DateQuestion 日期题
type DateQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&DateQuestion{BaseQuestion: question.BaseQuestion{T: "date", Dis: "日期", Cat: "base"}})
}

func (q *DateQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案必须为字符串"}
	}
	// 接受 YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS
	if !looksLikeDate(s) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "日期格式不合法"}
	}
	return nil
}

func (q *DateQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *DateQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请选择日期", "format": "YYYY-MM-DD"}
}

func looksLikeDate(s string) bool {
	if len(s) < 10 {
		return false
	}
	// 简化校验：YYYY-MM-DD 前 10 字符
	if s[4] != '-' || s[7] != '-' {
		return false
	}
	return true
}
