package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// PickerQuestion 选择器（type=picker），从 options 数组中单选（同 select 语义）。
// SurveyKing 用 picker 区分 "下拉" 与 "平铺"，这里兼容两者，校验/格式化复用 select 逻辑。
type PickerQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&PickerQuestion{BaseQuestion: question.BaseQuestion{T: "picker", Dis: "选择器（picker 别名）", Cat: "select"}})
}

func (q *PickerQuestion) Validate(value interface{}, sch schema.Question) error {
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

func (q *PickerQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return optionLabel(s, sch.PropsMap()["options"])
}

func (q *PickerQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请选择", "options": []map[string]interface{}{}}
}

// SignatureQuestion 签名题，值是 base64 字符串或 URL
type SignatureQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&SignatureQuestion{BaseQuestion: question.BaseQuestion{T: "signature", Dis: "签名", Cat: "media"}})
}

func (q *SignatureQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请签名"}
		}
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "签名题答案必须为字符串"}
	}
	if s == "" && sch.Required {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请签名"}
	}
	return nil
}

func (q *SignatureQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *SignatureQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请签名"}
}
