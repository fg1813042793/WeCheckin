package builtin

import (
	"fmt"
	"strings"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// SelectQuestion 单选下拉/单选按钮（type 通用 select）
type SelectQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&SelectQuestion{BaseQuestion: question.BaseQuestion{T: "select", Dis: "下拉选择", Cat: "select"}})
}

func (q *SelectQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案类型错误"}
	}
	if !isValueInOptions(s, sch.PropsMap()["options"]) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "选项值不合法"}
	}
	return nil
}

func (q *SelectQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	if s == "" {
		return ""
	}
	return optionLabel(s, sch.PropsMap()["options"])
}

func (q *SelectQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请选择",
		"options":     []map[string]interface{}{},
	}
}

// RadioQuestion 单选按钮
type RadioQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&RadioQuestion{BaseQuestion: question.BaseQuestion{T: "radio", Dis: "单选", Cat: "select"}})
}

func (q *RadioQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案类型错误"}
	}
	if !isValueInOptions(s, sch.PropsMap()["options"]) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "选项值不合法"}
	}
	return nil
}

func (q *RadioQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return optionLabel(s, sch.PropsMap()["options"])
}

func (q *RadioQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"options": []map[string]interface{}{}}
}

// CheckboxQuestion 多选
type CheckboxQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&CheckboxQuestion{BaseQuestion: question.BaseQuestion{T: "checkbox", Dis: "多选", Cat: "select"}})
}

func (q *CheckboxQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	arr, ok := value.([]interface{})
	if !ok {
		// 也接受 JSON 字符串形式
		if s, ok := value.(string); ok {
			if s == "" {
				return nil
			}
			return nil // 简化为接受，由前端序列化
		}
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "答案必须为数组"}
	}
	if len(arr) == 0 {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请至少选择一项"}
		}
		return nil
	}
	opts := sch.PropsMap()["options"]
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "选项值必须为字符串"}
		}
		if !isValueInOptions(s, opts) {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: fmt.Sprintf("选项 %q 不合法", s)}
		}
	}
	return nil
}

func (q *CheckboxQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	arr, ok := value.([]interface{})
	if !ok {
		if s, ok := value.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", value)
	}
	labels := make([]string, 0, len(arr))
	for _, item := range arr {
		s, _ := item.(string)
		labels = append(labels, optionLabel(s, sch.PropsMap()["options"]))
	}
	return strings.Join(labels, "、")
}

func (q *CheckboxQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"options": []map[string]interface{}{}}
}

// ---- helpers ----

// isValueInOptions 检查 value 是否在 options 数组的 value 字段中。
// options 形如 [{"label":"A","value":"a"},{"label":"B","value":"b"}]
func isValueInOptions(value string, options interface{}) bool {
	opts, ok := options.([]interface{})
	if !ok {
		// 兼容 map 数组 (从 JSON Unmarshal)
		if m, ok := options.([]map[string]interface{}); ok {
			opts = make([]interface{}, len(m))
			for i, v := range m {
				opts[i] = v
			}
		} else {
			return false
		}
	}
	for _, o := range opts {
		if m, ok := o.(map[string]interface{}); ok {
			if v, ok := m["value"].(string); ok && v == value {
				return true
			}
		}
	}
	return false
}

func optionLabel(value string, options interface{}) string {
	opts, _ := options.([]interface{})
	if opts == nil {
		if m, ok := options.([]map[string]interface{}); ok {
			opts = make([]interface{}, len(m))
			for i, v := range m {
				opts[i] = v
			}
		}
	}
	for _, o := range opts {
		if m, ok := o.(map[string]interface{}); ok {
			if v, ok := m["value"].(string); ok && v == value {
				if l, ok := m["label"].(string); ok {
					return l
				}
				return value
			}
		}
	}
	return value
}
