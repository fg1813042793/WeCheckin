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

	"wecheckin-backend/backend/internal/app/formkit/question"
	_ "wecheckin-backend/backend/internal/app/formkit/question/builtin" // 注册 24 个内置题型
	"wecheckin-backend/backend/internal/app/formkit/schema"
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
	TableData   [][]string               `json:"tableData,omitempty"`   // 矩阵填空/表格自增的明细
	TableCols   []string                 `json:"tableCols,omitempty"`   // 列标题
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
	// 解析原始 schema 用于获取矩阵题 props
	rawQuestions := parseRawQuestions(schemaJSON)
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
			} else if q.Type == "checkbox" {
				// 多选：每项单独计数，不做组合统计
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							key := checkboxOptionLabel(s, q.Options)
							dist[key]++
						}
					}
				}
			} else if q.Type == "user" || q.Type == "dept" {
				// 成员/部门：用 label 展示
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							key := checkboxOptionLabel(s, q.Options)
							dist[key]++
						}
					}
				} else if s, ok := v.(string); ok {
					key := checkboxOptionLabel(s, q.Options)
					dist[key]++
				}
			} else if q.Type == "matrixRadio" || q.Type == "matrixCheckbox" {
				m, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				colValToLabel := buildMatrixColMap(rawQuestions[q.ID])
				for _, rowVal := range m {
					if q.Type == "matrixRadio" {
						if s, ok := rowVal.(string); ok {
							label := colValToLabel[s]
							if label == "" {
								label = s
							}
							dist[label]++
						}
					} else if q.Type == "matrixCheckbox" {
						if arr, ok := rowVal.([]interface{}); ok {
							for _, item := range arr {
								s, _ := item.(string)
								label := colValToLabel[s]
								if label == "" {
									label = s
								}
								dist[label]++
							}
						}
					}
				}
			} else {
				key := formatValue(v, q)
				dist[key]++
			}
		}
		if len(dist) > 0 {
			fs.Dist = dist
		}
		// 文本类：收集所有回答文本
		if isTextType(q.Type) && fs.NonEmpty > 0 {
			fs.Dist = nil // 文本类不作分布统计
			if q.Type == "location" {
				// 位置：收集坐标列表用于地图展示
				var coords []string
				for _, v := range buckets[i] {
					if v == nil {
						continue
					}
					if s, ok := v.(string); ok && s != "" {
						coords = append(coords, s)
					}
				}
				if len(coords) > 0 {
					fs.TableCols = []string{"坐标"}
					rows := make([][]string, len(coords))
					for i, c := range coords {
						rows[i] = []string{c}
					}
					fs.TableData = rows
				}
			} else if q.Type == "multiInput" || q.Type == "hInput" {
				// 多项/横向填空：每列为字段名，每行为一个应答者
				colLabels := extractTextColLabels(rawQuestions[q.ID])
				fs.TableCols = colLabels
				var rows [][]string
				for _, v := range buckets[i] {
					if v == nil {
						continue
					}
					if arr, ok := v.([]interface{}); ok {
						row := make([]string, len(colLabels))
						for ci := 0; ci < len(colLabels) && ci < len(arr); ci++ {
							if s, ok := arr[ci].(string); ok {
								row[ci] = s
							}
						}
						rows = append(rows, row)
					} else if s, ok := v.(string); ok && s != "" {
						row := make([]string, len(colLabels))
						row[0] = s
						rows = append(rows, row)
					}
				}
				if len(rows) > 0 {
					fs.TableData = rows
				}
			} else {
				var texts []string
				for _, v := range buckets[i] {
					if v == nil {
						continue
					}
					if s, ok := v.(string); ok && s != "" {
						texts = append(texts, s)
					} else if arr, ok := v.([]interface{}); ok && q.Type == "dateRange" {
						// dateRange 存为 ["start","end"]
						if len(arr) == 2 {
							a, _ := arr[0].(string)
							b, _ := arr[1].(string)
							texts = append(texts, a+" ~ "+b)
						}
					}
				}
				if len(texts) > 0 {
					fs.TableCols = []string{"回答"}
					rows := make([][]string, len(texts))
					for i, t := range texts {
						rows[i] = []string{t}
					}
					fs.TableData = rows
				}
			}
		}
		// 文件上传：收集所有文件名
		if q.Type == "file" && fs.NonEmpty > 0 {
			var fileList []string
			for _, v := range buckets[i] {
				if v == nil {
					continue
				}
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok && s != "" {
							fileList = append(fileList, s)
						}
					}
				} else if s, ok := v.(string); ok && s != "" {
					fileList = append(fileList, s)
				}
			}
			if len(fileList) > 0 {
				fs.TableCols = []string{"文件名"}
				rows := make([][]string, len(fileList))
				for i, f := range fileList {
					rows[i] = []string{f}
				}
				fs.TableData = rows
			}
		}
		// 矩阵填空：收集所有应答者填充的单元格
		if (q.Type == "matrixFillBlank" || q.Type == "matrixAuto") && fs.NonEmpty > 0 {
			colLabels, colCount := extractMatrixColKeys(rawQuestions[q.ID], q.Type)
			fs.TableCols = colLabels
			var rows [][]string
			for _, v := range buckets[i] {
				if v == nil {
					continue
				}
				if q.Type == "matrixFillBlank" {
					row := make([]string, colCount)
					if m, ok := v.(map[string]interface{}); ok {
						for ci := 0; ci < colCount; ci++ {
							key := fmt.Sprintf("%d", ci)
							if cell, ok := m[key].(map[string]interface{}); ok {
								for _, fv := range cell {
									if s, ok := fv.(string); ok {
										row[ci] = s
										break
									}
								}
							}
						}
					}
					rows = append(rows, row)
				} else if q.Type == "matrixAuto" {
					if arr, ok := v.([]interface{}); ok {
						for _, rowItem := range arr {
							row := make([]string, colCount)
							if rowArr, rok := rowItem.([]interface{}); rok {
								for ci := 0; ci < colCount && ci < len(rowArr); ci++ {
									if s, ok := rowArr[ci].(string); ok {
										row[ci] = s
									}
								}
							}
							rows = append(rows, row)
						}
					}
				}
			}
			if len(rows) > 0 {
				fs.TableData = rows
			}
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

// parseRawQuestions 解析 schema JSON 为 questionId → 原始 question map
func parseRawQuestions(schemaJSON string) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	var raw interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &raw); err != nil {
		return out
	}
	if schema.IsOldFormat(schemaJSON) {
		if arr, ok := raw.([]interface{}); ok {
			for idx, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					id := fmt.Sprintf("q%d", idx+1)
					out[id] = m
				}
			}
		}
	} else {
		if root, ok := raw.(map[string]interface{}); ok {
			if qs, ok := root["questions"].([]interface{}); ok {
				for _, item := range qs {
					if m, ok := item.(map[string]interface{}); ok {
						id, _ := m["id"].(string)
						if id != "" {
							out[id] = m
						}
					}
				}
			}
		}
	}
	return out
}

// extractMatrixColKeys 提取矩阵题的列标题和列数
func extractMatrixColKeys(qm map[string]interface{}, qType string) (labels []string, count int) {
	if qm == nil {
		return nil, 0
	}
	props, _ := qm["props"].(map[string]interface{})
	if props == nil {
		return nil, 0
	}
	cols, _ := props["columns"].([]interface{})
	count = len(cols)
	labels = make([]string, count)
	for i, c := range cols {
		if cm, ok := c.(map[string]interface{}); ok {
			if l, ok := cm["label"].(string); ok && l != "" {
				labels[i] = l
			} else if v, ok := cm["value"].(string); ok {
				labels[i] = v
			}
		}
	}
	return
}

// buildMatrixColMap 从原始 question map 提取列 value → label 映射
func buildMatrixColMap(qm map[string]interface{}) map[string]string {
	colMap := map[string]string{}
	if qm == nil {
		return colMap
	}
	props, _ := qm["props"].(map[string]interface{})
	if props == nil {
		return colMap
	}
	cols, _ := props["columns"].([]interface{})
	for _, c := range cols {
		if cm, ok := c.(map[string]interface{}); ok {
			v, _ := cm["value"].(string)
			l, _ := cm["label"].(string)
			if v != "" {
				if l != "" {
					colMap[v] = l
				} else {
					colMap[v] = v
				}
			}
		}
	}
	return colMap
}

// checkboxOptionLabel 从 options 数组查找选项 value 对应的 label，找不到则返回 value 自身
// isTextType 判断是否为开放文本题型（不做分布统计）
func isTextType(t string) bool {
	switch t {
	case "input", "textarea", "text", "phone", "email", "idCard", "password", "multiInput", "hInput", "location", "date", "time", "dateRange", "richText", "autopop", "signature", "name", "studentId", "employeeId", "class":
		return true
	}
	return false
}

// extractTextColLabels 提取 multiInput/hInput 的字段名列表
func extractTextColLabels(qm map[string]interface{}) []string {
	if qm == nil {
		return []string{"字段1"}
	}
	props, _ := qm["props"].(map[string]interface{})
	if props == nil {
		return []string{"字段1"}
	}
	fields, _ := props["fields"].([]interface{})
	if len(fields) == 0 {
		return []string{"字段1"}
	}
	labels := make([]string, len(fields))
	for i, f := range fields {
		if fm, ok := f.(map[string]interface{}); ok {
			if l, ok := fm["label"].(string); ok && l != "" {
				labels[i] = l
			} else if v, ok := fm["value"].(string); ok {
				labels[i] = v
			} else {
				labels[i] = fmt.Sprintf("字段%d", i+1)
			}
		} else {
			labels[i] = fmt.Sprintf("字段%d", i+1)
		}
	}
	return labels
}

func checkboxOptionLabel(value string, opts []map[string]interface{}) string {
	for _, o := range opts {
		if v, ok := o["value"].(string); ok && v == value {
			if l, ok := o["label"].(string); ok && l != "" {
				return l
			}
			return value
		}
	}
	return value
}
