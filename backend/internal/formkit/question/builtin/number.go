package builtin

import (
	"fmt"
	"strconv"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

type NumberQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&NumberQuestion{BaseQuestion: question.BaseQuestion{T: "number", Dis: "数字", Cat: "base"}})
}

func (q *NumberQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	// 接受 number/string(可解析为数字) 两种
	var f float64
	switch v := value.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case string:
		if v == "" {
			return nil
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入有效的数字"}
		}
		f = parsed
	default:
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请输入数字"}
	}
	// 范围检查（来自 props.min/max）
	props := sch.PropsMap()
	if mn, ok := props["min"].(float64); ok && f < mn {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: fmt.Sprintf("不能小于 %v", mn)}
	}
	if mx, ok := props["max"].(float64); ok && f > mx {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: fmt.Sprintf("不能大于 %v", mx)}
	}
	return nil
}

func (q *NumberQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case string:
		return v
	}
	return fmt.Sprintf("%v", value)
}

func (q *NumberQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "请输入数字", "min": 0, "max": 999999}
}
