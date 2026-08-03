package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"wecheckin/backend/internal/app/formkit/question"
	"wecheckin/backend/internal/app/formkit/schema"
)

// MultiInputQuestion 多项填空
type MultiInputQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&MultiInputQuestion{BaseQuestion: question.BaseQuestion{T: "multiInput", Dis: "多项填空", Cat: "fill"}})
}

func (q *MultiInputQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	return validateMultiFields(value, sch)
}

func (q *MultiInputQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请输入",
		"fields":      []map[string]interface{}{},
	}
}

// HInputQuestion 横向填空
type HInputQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&HInputQuestion{BaseQuestion: question.BaseQuestion{T: "hInput", Dis: "横向填空", Cat: "fill"}})
}

func (q *HInputQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	return validateMultiFields(value, sch)
}

func (q *HInputQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"placeholder": "请输入",
		"fields":      []map[string]interface{}{},
	}
}

// validateMultiFields 校验多项填空的答案。
// value 接受 []interface{}（数组下标索引）或 map[string]interface{}（key 索引）。
func validateMultiFields(value interface{}, sch schema.Question) error {
	// 收集字段值
	values := collectFieldValues(value)
	if values == nil {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "多项填空答案必须为数组或对象"}
	}
	// 读取字段定义
	fields := getFieldDefs(sch)
	if len(fields) == 0 {
		// 没有字段定义时只检查每项是否为字符串
		for i, v := range values {
			if _, ok := v.(string); !ok {
				return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: fmt.Sprintf("第 %d 项必须为字符串", i+1)}
			}
		}
		return nil
	}
	for i, f := range fields {
		if i >= len(values) {
			if f.required {
				return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: f.label + "为必填"}
			}
			continue
		}
		fv := values[i]
		if fv == nil || fmt.Sprintf("%v", fv) == "" {
			if f.required {
				return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: f.label + "为必填"}
			}
			continue
		}
		s, ok := fv.(string)
		if !ok {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: f.label + "必须为字符串"}
		}
		if err := validateFieldByType(s, f.dataType, f.decimalPlaces); err != nil {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: f.label + ": " + err.Error()}
		}
	}
	return nil
}

// fieldDef 字段定义
type fieldDef struct {
	label         string
	dataType      string
	decimalPlaces int
	required      bool
}

// getFieldDefs 从 schema props 中提取字段定义列表
func getFieldDefs(sch schema.Question) []fieldDef {
	props := sch.PropsMap()
	raw, ok := props["fields"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]fieldDef, 0, len(raw))
	for _, r := range raw {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		f := fieldDef{}
		if v, _ := m["label"].(string); v != "" {
			f.label = v
		} else {
			f.label = fmt.Sprintf("第 %d 项", len(out)+1)
		}
		if v, _ := m["dataType"].(string); v != "" {
			f.dataType = v
		}
		if v, ok := m["decimalPlaces"].(float64); ok {
			f.decimalPlaces = int(v)
		}
		if v, ok := m["required"].(bool); ok {
			f.required = v
		}
		out = append(out, f)
	}
	return out
}

// collectFieldValues 将答案统一为 []interface{}
func collectFieldValues(value interface{}) []interface{} {
	switch v := value.(type) {
	case []interface{}:
		return v
	case map[string]interface{}:
		out := make([]interface{}, 0, len(v))
		for i := 0; i < len(v); i++ {
			key := strconv.Itoa(i)
			if val, ok := v[key]; ok {
				out = append(out, val)
			} else {
				// 尝试字符串 key 匹配
				break
			}
		}
		if len(out) == 0 {
			// 用 label/key 遍历
			for _, val := range v {
				out = append(out, val)
			}
		}
		return out
	}
	return nil
}

// validateFieldByType 按 dataType 校验单个字段值
func validateFieldByType(s, dataType string, decimalPlaces int) error {
	switch dataType {
	case "mobile":
		if !phoneRE.MatchString(s) {
			return fmt.Errorf("手机号格式不合法")
		}
	case "email":
		if !emailRE.MatchString(s) {
			return fmt.Errorf("邮箱格式不合法")
		}
	case "idCard":
		if !idCardRE.MatchString(s) {
			return fmt.Errorf("身份证号格式不合法")
		}
	case "number":
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return fmt.Errorf("必须为数字")
		}
		if decimalPlaces > 0 {
			parts := strings.Split(s, ".")
			if len(parts) == 2 && len(parts[1]) > decimalPlaces {
				return fmt.Errorf("小数位数不能超过 %d 位", decimalPlaces)
			}
		}
	}
	return nil
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
