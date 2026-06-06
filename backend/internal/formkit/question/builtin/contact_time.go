package builtin

import (
	"regexp"
	"strconv"
	"strings"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// PhoneQuestion 手机号（type=phone）
type PhoneQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&PhoneQuestion{BaseQuestion: question.BaseQuestion{T: "phone", Dis: "手机号", Cat: "base"}})
}

var phoneRE = regexp.MustCompile(`^1[3-9]\d{9}$`)

func (q *PhoneQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入手机号"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "手机号必须为字符串"}
	}
	if !phoneRE.MatchString(s) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "手机号格式不合法"}
	}
	return nil
}

func (q *PhoneQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *PhoneQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入 11 位手机号"}
}

// EmailQuestion 邮箱（type=email）
type EmailQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&EmailQuestion{BaseQuestion: question.BaseQuestion{T: "email", Dis: "邮箱", Cat: "base"}})
}

var emailRE = regexp.MustCompile(`^[\w.+-]+@[\w-]+(\.[\w-]+)+$`)

func (q *EmailQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入邮箱"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "邮箱必须为字符串"}
	}
	if !emailRE.MatchString(s) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "邮箱格式不合法"}
	}
	return nil
}

func (q *EmailQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *EmailQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入邮箱"}
}

// IDCardQuestion 身份证号（type=idCard）
type IDCardQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&IDCardQuestion{BaseQuestion: question.BaseQuestion{T: "idCard", Dis: "身份证号", Cat: "base"}})
}

var idCardRE = regexp.MustCompile(`^\d{17}[\dXx]$`)

func (q *IDCardQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入身份证号"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "身份证号必须为字符串"}
	}
	if !idCardRE.MatchString(s) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "身份证号格式不合法"}
	}
	return nil
}

func (q *IDCardQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *IDCardQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入 18 位身份证号"}
}

// PasswordQuestion 密码（type=password）
type PasswordQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&PasswordQuestion{BaseQuestion: question.BaseQuestion{T: "password", Dis: "密码", Cat: "base"}})
}

func (q *PasswordQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入密码"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "密码必须为字符串"}
	}
	if len(s) < 6 {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "密码至少 6 位"}
	}
	return nil
}

func (q *PasswordQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	return "******"
}

func (q *PasswordQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入密码"}
}

// TimeQuestion 时间 HH:MM 或 HH:MM:SS
type TimeQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&TimeQuestion{BaseQuestion: question.BaseQuestion{T: "time", Dis: "时间", Cat: "base"}})
}

func (q *TimeQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请选择时间"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "时间必须为字符串"}
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "时间格式不合法"}
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "时间格式不合法"}
		}
	}
	return nil
}

func (q *TimeQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *TimeQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请选择时间", "format": "HH:mm"}
}

// DateRangeQuestion 日期范围，存 ["YYYY-MM-DD","YYYY-MM-DD"]
type DateRangeQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&DateRangeQuestion{BaseQuestion: question.BaseQuestion{T: "dateRange", Dis: "日期范围", Cat: "base"}})
}

func (q *DateRangeQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请选择日期范围"}
		}
		return nil
	}
	arr, ok := value.([]interface{})
	if !ok || len(arr) != 2 {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "日期范围必须为 2 个元素"}
	}
	for i, item := range arr {
		s, ok := item.(string)
		if !ok || !looksLikeDate(s) {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "日期格式不合法 (第 " + strconv.Itoa(i+1) + " 项)"}
		}
	}
	return nil
}

func (q *DateRangeQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	arr, ok := value.([]interface{})
	if !ok || len(arr) != 2 {
		if s, ok := value.(string); ok {
			return s
		}
		return ""
	}
	a, _ := arr[0].(string)
	b, _ := arr[1].(string)
	return a + " ~ " + b
}

func (q *DateRangeQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请选择起止日期"}
}
