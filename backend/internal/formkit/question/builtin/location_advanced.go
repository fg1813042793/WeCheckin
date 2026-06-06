package builtin

import (
	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

// LocationQuestion 位置/打卡（type=location）
// 值是 {"address":"...","lat":"...","lng":"..."}
type LocationQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&LocationQuestion{BaseQuestion: question.BaseQuestion{T: "location", Dis: "位置/打卡", Cat: "media"}})
}

func (q *LocationQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请获取位置"}
		}
		return nil
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "位置题答案必须为对象"}
	}
	addr, _ := m["address"].(string)
	if addr == "" {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请获取位置（地址为空）"}
	}
	return nil
}

func (q *LocationQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		if s, ok := value.(string); ok {
			return s
		}
		return ""
	}
	addr, _ := m["address"].(string)
	lat, _ := m["lat"].(string)
	lng, _ := m["lng"].(string)
	if lat == "" && lng == "" {
		return addr
	}
	return addr + " (" + lat + "," + lng + ")"
}

func (q *LocationQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"placeholder": "点击获取位置"}
}

// AutoPopQuestion 自动填充（type=autopop），从用户档案自动填入
// 值是 string，校验逻辑是 "只要用户提供了来源字段" 即可
type AutoPopQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&AutoPopQuestion{BaseQuestion: question.BaseQuestion{T: "autopop", Dis: "自动填充（来自用户）", Cat: "advanced"}})
}

func (q *AutoPopQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "此项为必填"}
		}
		return nil
	}
	if _, ok := value.(string); !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "自动填充题答案必须为字符串"}
	}
	return nil
}

func (q *AutoPopQuestion) FormatValue(value interface{}, _ schema.Question) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}

func (q *AutoPopQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{"source": "user.name", "placeholder": "自动填充"}
}

// MatrixRadioQuestion 矩阵单选（type=matrixRadio）
// props.rows + props.columns；值是 {"row1":"colA","row2":"colB"}
type MatrixRadioQuestion struct{ question.BaseQuestion }

func init() {
	question.Register(&MatrixRadioQuestion{BaseQuestion: question.BaseQuestion{T: "matrixRadio", Dis: "矩阵单选", Cat: "advanced"}})
}

func (q *MatrixRadioQuestion) Validate(value interface{}, sch schema.Question) error {
	if value == nil {
		if sch.Required {
			return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请完成所有行"}
		}
		return nil
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "矩阵单选答案必须为对象"}
	}
	props := sch.PropsMap()
	rows, _ := props["rows"].([]interface{})
	if sch.Required && len(m) < len(rows) {
		return &question.ValidationError{QuestionID: sch.ID, Field: sch.ID, Message: "请完成所有行"}
	}
	return nil
}

func (q *MatrixRadioQuestion) FormatValue(value interface{}, sch schema.Question) string {
	if value == nil {
		return ""
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		if s, ok := value.(string); ok {
			return s
		}
		return ""
	}
	props := sch.PropsMap()
	cols, _ := props["columns"].([]interface{})
	colLabels := map[string]string{}
	for _, c := range cols {
		if cm, ok := c.(map[string]interface{}); ok {
			v, _ := cm["value"].(string)
			l, _ := cm["label"].(string)
			colLabels[v] = l
		}
	}
	rows, _ := props["rows"].([]interface{})
	rowLabels := []string{}
	for _, r := range rows {
		rm, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		rk, _ := rm["value"].(string)
		rl, _ := rm["label"].(string)
		v, _ := m[rk].(string)
		if v == "" {
			rowLabels = append(rowLabels, rl+": -")
		} else {
			rowLabels = append(rowLabels, rl+": "+colLabels[v])
		}
	}
	out := ""
	for i, l := range rowLabels {
		if i > 0 {
			out += "; "
		}
		out += l
	}
	return out
}

func (q *MatrixRadioQuestion) DefaultProps() map[string]interface{} {
	return map[string]interface{}{
		"rows":    []map[string]interface{}{},
		"columns": []map[string]interface{}{},
	}
}
