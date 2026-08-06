package schema

import (
	"encoding/json"
	"strconv"
	"strings"
)

// FieldValue 标签-值对，用于兼容老代码读 Forms 数据的场景
type FieldValue struct {
	Label string      `json:"label"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ExtractFieldValues 把任意形式的 answer（数组/对象）+ schema（老/新）转为统一的 label-value 列表
// 兼容策略：
//   - 新格式 answer 对象 + 新格式 schema → 按 schema 顺序输出，value 从对象取
//   - 老格式 answer 数组 + 老格式 schema → 直接 zip
//   - 老格式 answer 数组 + 新格式 schema → 按 schema 顺序 zip
//   - 新格式 answer 对象 + 老格式 schema → 按老 schema 顺序用 q1/q2 索引
func ExtractFieldValues(answerJSON, schemaJSON string) []FieldValue {
	answerJSON = strings.TrimSpace(answerJSON)
	schemaJSON = strings.TrimSpace(schemaJSON)

	if answerJSON == "" {
		return []FieldValue{}
	}

	// 解析 answer 为通用 map（兼容老/新）
	var answerMap map[string]interface{}
	if strings.HasPrefix(answerJSON, "[") {
		// 老格式数组 → 转 map（q1/q2/...）
		var arr []interface{}
		if err := json.Unmarshal([]byte(answerJSON), &arr); err != nil {
			return []FieldValue{}
		}
		answerMap = make(map[string]interface{}, len(arr))
		for i, v := range arr {
			answerMap["q"+strconv.Itoa(i+1)] = v
		}
	} else if strings.HasPrefix(answerJSON, "{") {
		json.Unmarshal([]byte(answerJSON), &answerMap)
	} else {
		return []FieldValue{}
	}

	// 解析 schema
	if schemaJSON == "" {
		// 无 schema 时只输出 answerMap 的所有值
		out := make([]FieldValue, 0, len(answerMap))
		for k, v := range answerMap {
			out = append(out, FieldValue{Label: k, Type: "", Value: v})
		}
		return out
	}

	if IsOldFormat(schemaJSON) {
		// 老格式 schema: 数组，每个含 label/type
		var olds []OldField
		if err := json.Unmarshal([]byte(schemaJSON), &olds); err != nil {
			return []FieldValue{}
		}
		out := make([]FieldValue, 0, len(olds))
		for i, old := range olds {
			id := "q" + strconv.Itoa(i+1)
			out = append(out, FieldValue{
				Label: old.Label,
				Type:  old.Type,
				Value: answerMap[id],
			})
		}
		return out
	}

	// 新格式 schema
	s, err := Parse(schemaJSON)
	if err != nil {
		return []FieldValue{}
	}
	out := make([]FieldValue, 0, len(s.Questions))
	for _, q := range s.Questions {
		out = append(out, FieldValue{
			Label: q.Title,
			Type:  q.Type,
			Value: answerMap[q.ID],
		})
	}
	return out
}

// ExtractImagesLocation 从 answer+schema 中按"包含图/照片/img/pic/image"的 label 提取图片 URL，
// 按"包含位置但不包含纬度/经度"的 label 提取位置文本。返回的 images 至少为 []string{}。
// 兼容老/新格式数据。
func ExtractImagesLocation(answerJSON, schemaJSON string) (images []string, location string) {
	images = []string{}
	fvs := ExtractFieldValues(answerJSON, schemaJSON)
	for _, fv := range fvs {
		label := fv.Label
		val, _ := fv.Value.(string)
		if val == "" {
			continue
		}
		lower := strings.ToLower(label)
		if strings.Contains(lower, "图") || strings.Contains(lower, "照片") ||
			strings.Contains(lower, "img") || strings.Contains(lower, "pic") ||
			strings.Contains(lower, "image") {
			images = append(images, val)
		}
		if strings.Contains(lower, "位置") && !strings.Contains(lower, "纬度") && !strings.Contains(lower, "经度") {
			location = val
		}
	}
	return
}
