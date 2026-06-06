package schema

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// OldField 老格式字段（来自 event_forms/enroll_forms）
type OldField struct {
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
}

// typeMapping 老类型 → 新类型
var typeMapping = map[string]string{
	"input":    "input",
	"text":     "input",
	"textarea": "textarea",
	"select":   "select",
	"picker":   "select",
	"number":   "number",
}

// IsOldFormat 检测输入是否为老格式（数组）vs 新格式（对象）
func IsOldFormat(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "[") {
		return true
	}
	if strings.HasPrefix(s, "{") {
		return false
	}
	return false
}

// MigrateFromOld 老 schema 数组 → 新 FormSchema JSON 字符串
func MigrateFromOld(oldJSON string) (string, error) {
	oldJSON = strings.TrimSpace(oldJSON)
	if oldJSON == "" {
		return Marshal(&FormSchema{Version: CurrentVersion, Questions: []Question{}})
	}
	var olds []OldField
	if err := json.Unmarshal([]byte(oldJSON), &olds); err != nil {
		return "", fmt.Errorf("invalid old schema: %w", err)
	}
	schema := &FormSchema{
		Version:   CurrentVersion,
		Questions: make([]Question, 0, len(olds)),
	}
	for i, old := range olds {
		id := "q" + strconv.Itoa(i+1)
		newType, ok := typeMapping[old.Type]
		if !ok {
			newType = "input"
		}
		q := Question{
			ID:          id,
			Type:        newType,
			Title:       old.Label,
			Required:    old.Required,
			Placeholder: old.Placeholder,
		}
		if len(old.Options) > 0 {
			opts := make([]map[string]string, 0, len(old.Options))
			for _, opt := range old.Options {
				opts = append(opts, map[string]string{"label": opt, "value": opt})
			}
			propsBytes, _ := json.Marshal(map[string]interface{}{"options": opts})
			q.Props = propsBytes
		}
		schema.Questions = append(schema.Questions, q)
	}
	return Marshal(schema)
}

// LoadSchema 智能加载：自动判断老/新格式并返回新格式
func LoadSchema(raw string) (*FormSchema, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Empty(), nil
	}
	if IsOldFormat(raw) {
		migrated, err := MigrateFromOld(raw)
		if err != nil {
			return nil, err
		}
		return Parse(migrated)
	}
	return Parse(raw)
}

// LoadSchemaString 同 LoadSchema，但返回 JSON 字符串（便于直接覆盖原字段）
func LoadSchemaString(raw string) (string, error) {
	s, err := LoadSchema(raw)
	if err != nil {
		return "", err
	}
	return Marshal(s)
}

// MigrateAnswer 老 answer 数组 → 新 answer 对象 JSON 字符串
// 老: ["张三","13800138000"]
// 新: {"q1":"张三","q2":"13800138000"}
func MigrateAnswer(oldAnswerJSON string, schema *FormSchema) (string, error) {
	oldAnswerJSON = strings.TrimSpace(oldAnswerJSON)
	if oldAnswerJSON == "" {
		return "{}", nil
	}
	if strings.HasPrefix(oldAnswerJSON, "{") {
		var tmp map[string]interface{}
		if err := json.Unmarshal([]byte(oldAnswerJSON), &tmp); err == nil {
			return oldAnswerJSON, nil
		}
	}
	var arr []interface{}
	if err := json.Unmarshal([]byte(oldAnswerJSON), &arr); err != nil {
		return "{}", nil
	}
	obj := make(map[string]interface{})
	for i, v := range arr {
		if schema != nil && i < len(schema.Questions) {
			obj[schema.Questions[i].ID] = v
		} else {
			obj["q"+strconv.Itoa(i+1)] = v
		}
	}
	b, _ := json.Marshal(obj)
	return string(b), nil
}

// LoadAnswer 智能加载 answer：老数组 → 新对象
func LoadAnswer(rawAnswerJSON string, schema *FormSchema) (string, error) {
	return MigrateAnswer(rawAnswerJSON, schema)
}

// AnswerToLegacyArray 新 answer 对象 → 老 answer 数组（仅当 schema 已知时可用）
// 主要用于兼容那些还在用数组的旧代码路径
func AnswerToLegacyArray(answerJSON string, schema *FormSchema) ([]interface{}, error) {
	answerJSON = strings.TrimSpace(answerJSON)
	if answerJSON == "" {
		return []interface{}{}, nil
	}
	if strings.HasPrefix(answerJSON, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(answerJSON), &arr); err != nil {
			return nil, err
		}
		return arr, nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(answerJSON), &obj); err != nil {
		return nil, err
	}
	if schema == nil {
		out := make([]interface{}, 0, len(obj))
		for _, v := range obj {
			out = append(out, v)
		}
		return out, nil
	}
	out := make([]interface{}, 0, len(schema.Questions))
	for _, q := range schema.Questions {
		if v, ok := obj[q.ID]; ok {
			out = append(out, v)
		} else {
			out = append(out, nil)
		}
	}
	return out, nil
}
