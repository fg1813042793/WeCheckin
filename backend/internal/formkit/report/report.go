// Package report 实现 formkit 的答题数据报表/导出功能。
//
// 主要能力：
//   - 把每条 answer（map 或 array）按 schema 转为 "label → displayValue" 的有序行
//   - 把多条 answer 渲染为表格
//   - 导出 CSV
//   - 按字段做基础统计（计数/分布）
package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"wecheckin-backend/backend/internal/formkit/question"
	_ "wecheckin-backend/backend/internal/formkit/question/builtin" // 注册 24 个内置题型
	"wecheckin-backend/backend/internal/formkit/schema"
)

// Row 单条答题的渲染结果：一行字段值，按 schema 顺序
type Row struct {
	UserID  string
	AddTime string
	Values  []string // 按 schema.Questions 顺序的展示值
}

// Table 报表表格
type Table struct {
	Headers []string // 第一行：固定列 + schema 题目标题
	Rows    []Row
}

// RenderAnswers 把所有 answers 渲染为 Table。
// answers 是 "userID/forms/addTime" 列表，forms 是 JSON 字符串（老/新格式均可）。
// schemaJSON 是 schema 字符串（老/新均可）。
func RenderAnswers(schemaJSON string, items []AnswerItem) (Table, error) {
	questions := schema.NormalizeSchemaForReport(schemaJSON)
	tbl := Table{Headers: make([]string, 0, len(questions)+3)}
	tbl.Headers = append(tbl.Headers, "用户ID", "提交时间")
	for _, q := range questions {
		tbl.Headers = append(tbl.Headers, q.Title)
	}

	for _, it := range items {
		row := Row{UserID: it.UserID, AddTime: it.AddTime, Values: make([]string, 0, len(questions))}
		ans := parseAnswers(it.Forms)
		for _, q := range questions {
			v := lookupAnswer(ans, q.ID, q.OldIndex)
			row.Values = append(row.Values, formatValue(v, q))
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	return tbl, nil
}

// AnswerItem 报表输入：单条答题
type AnswerItem struct {
	UserID  string
	AddTime string
	Forms   string
}

// parseAnswers 解析 forms JSON 为通用 map（兼容老/新）
func parseAnswers(forms string) map[string]interface{} {
	out := map[string]interface{}{}
	if forms == "" {
		return out
	}
	trimmed := strings.TrimSpace(forms)
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			for i, v := range arr {
				out["q"+strconv.Itoa(i+1)] = v
			}
		}
	} else if strings.HasPrefix(trimmed, "{") {
		_ = json.Unmarshal([]byte(trimmed), &out)
	}
	return out
}

// lookupAnswer 取出某题的答案值
func lookupAnswer(ans map[string]interface{}, id string, oldIndex int) interface{} {
	if v, ok := ans[id]; ok {
		return v
	}
	if oldIndex >= 0 {
		if v, ok := ans["q"+strconv.Itoa(oldIndex+1)]; ok {
			return v
		}
	}
	return nil
}

// formatValue 用 question 包把答案格式化为展示串
func formatValue(v interface{}, q schema.ReportQuestion) string {
	if v == nil {
		return ""
	}
	inst := question.Get(q.Type)
	if inst == nil {
		if s, ok := v.(string); ok {
			return s
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
	// 构造一个最小 schema.Question（带 options）
	props := map[string]interface{}{}
	if len(q.Options) > 0 {
		opts := make([]interface{}, len(q.Options))
		for i, o := range q.Options {
			opts[i] = o
		}
		props["options"] = opts
	}
	sch := schema.Question{Type: q.Type, Title: q.Title}
	if len(props) > 0 {
		b, _ := json.Marshal(props)
		sch.Props = b
	}
	return inst.FormatValue(v, sch)
}

// ToCSV 把 Table 写成 CSV 字节数组
func ToCSV(t Table) []byte {
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	_ = w.Write(t.Headers)
	for _, r := range t.Rows {
		row := []string{r.UserID, r.AddTime}
		row = append(row, r.Values...)
		_ = w.Write(row)
	}
	w.Flush()
	return []byte(buf.String())
}

// FieldStat 单字段统计
type FieldStat struct {
	QuestionID  string                   `json:"questionId"`
	Title       string                   `json:"title"`
	Type        string                   `json:"type"`
	TotalCount  int                      `json:"totalCount"`
	NonEmpty    int                      `json:"nonEmpty"`
	Empty       int                      `json:"empty"`
	Dist        map[string]int           `json:"dist,omitempty"`        // 字符串/数值 答案 → 次数
	NumericStat *NumericStat             `json:"numericStat,omitempty"` // 仅数字
}

// NumericStat 数字统计
type NumericStat struct {
	Sum    float64 `json:"sum"`
	Avg    float64 `json:"avg"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Median float64 `json:"median,omitempty"`
}

// FieldStats 计算 schema 各字段的统计
func FieldStats(schemaJSON string, items []AnswerItem) []FieldStat {
	questions := schema.NormalizeSchemaForReport(schemaJSON)
	out := make([]FieldStat, 0, len(questions))
	// 收集每字段的答案
	buckets := make([][]interface{}, len(questions))
	for _, it := range items {
		ans := parseAnswers(it.Forms)
		for i, q := range questions {
			buckets[i] = append(buckets[i], lookupAnswer(ans, q.ID, q.OldIndex))
		}
	}
	for i, q := range questions {
		fs := FieldStat{
			QuestionID: q.ID,
			Title:      q.Title,
			Type:       q.Type,
			TotalCount: len(buckets[i]),
		}
		// 跳过布局/无值题型
		if q.Type == "divider" || q.Type == "description" {
			continue
		}
		dist := map[string]int{}
		var nums []float64
		for _, v := range buckets[i] {
			if v == nil || v == "" {
				fs.Empty++
				continue
			}
			fs.NonEmpty++
			if q.Type == "number" {
				if f, ok := toFloat64(v); ok {
					nums = append(nums, f)
				}
			} else {
				key := formatValue(v, q)
				dist[key]++
			}
		}
		if len(dist) > 0 {
			fs.Dist = dist
		}
		if q.Type == "number" && len(nums) > 0 {
			stat := &NumericStat{Min: nums[0], Max: nums[0]}
			sum := 0.0
			for _, n := range nums {
				sum += n
				if n < stat.Min {
					stat.Min = n
				}
				if n > stat.Max {
					stat.Max = n
				}
			}
			stat.Sum = sum
			stat.Avg = sum / float64(len(nums))
			fs.NumericStat = stat
		}
		out = append(out, fs)
	}
	return out
}

func toFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

// SanitizeFilename 去掉 CSV 文件名非法字符
func SanitizeFilename(s string) string {
	bad := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "\t", "\n"}
	out := s
	for _, b := range bad {
		out = strings.ReplaceAll(out, b, "_")
	}
	if out == "" {
		return "export"
	}
	return out
}

// StringPtr 辅助：转 *string
func StringPtr(s string) *string { return &s }

// fmtInt 辅助：整数转字符串
func fmtInt(n int) string { return fmt.Sprintf("%d", n) }
var _ = fmtInt
