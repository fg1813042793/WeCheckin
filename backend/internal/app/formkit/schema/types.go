package schema

import "encoding/json"

const CurrentVersion = "2.0"

type FormSchema struct {
	Version   string     `json:"version"`
	Questions []Question `json:"questions"`
	Setting   Setting    `json:"setting,omitempty"`
}

type Setting struct {
	SubmitBtnText string `json:"submitBtnText,omitempty"`
	ShowProgress  bool   `json:"showProgress,omitempty"`
	AllowDraft    bool   `json:"allowDraft,omitempty"`
	Password      string `json:"password,omitempty"`
	Deadline      int64  `json:"deadline,omitempty"`
	PostSubmitMsg string `json:"postSubmitMsg,omitempty"`
}

type Question struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Title        string          `json:"title"`
	Description  string          `json:"description,omitempty"`
	Required     bool            `json:"required,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Placeholder  string          `json:"placeholder,omitempty"`
	Help         string          `json:"help,omitempty"`
	Props        json.RawMessage `json:"props,omitempty"`
	Validate     []ValidateRule  `json:"validate,omitempty"`
	Logic        []LogicRule     `json:"logic,omitempty"`
	CalcValue    *CalcValue      `json:"calcValue,omitempty"`
}

type ValidateRule struct {
	Type    string      `json:"type"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message,omitempty"`
}

type LogicRule struct {
	When   LogicCondition `json:"when"`
	Action string         `json:"action"`
	Target string         `json:"target"`
}

type LogicCondition struct {
	QuestionID string      `json:"questionId"`
	Operator   string      `json:"operator"`
	Value      interface{} `json:"value,omitempty"`
}

type CalcValue struct {
	Expr   string `json:"expr"`
	Target string `json:"target,omitempty"`
}

// PropsMap 把 Props (json.RawMessage) 解析为 map[string]interface{}，方便访问。
// Props 为空时返回空 map（不报错）。
func (q Question) PropsMap() map[string]interface{} {
	out := map[string]interface{}{}
	if len(q.Props) == 0 {
		return out
	}
	_ = json.Unmarshal(q.Props, &out)
	return out
}

// DefaultValueString 把 DefaultValue (json.RawMessage) 解析为字符串。
// 若不是字符串或为空，返回 ""。
func (q Question) DefaultValueString() string {
	if len(q.DefaultValue) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(q.DefaultValue, &s); err == nil {
		return s
	}
	return string(q.DefaultValue)
}
