// Package exam 实现 formkit 题库 + 考试 + 自动阅卷。
//
// 自动阅卷：
//   - 客观题（input/select/radio/number/phone/email/idCard/date/time/...）：直接对比正确答案
//   - checkbox：集合相等（无序）
//   - 主观题（textarea）：标记为"待人工判分"，不计入自动得分
//   - 数值题：在 tolerance 范围内算正确
package exam

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"wecheckin-backend/backend/internal/formkit/question"
)

// Question 题目（来自题库）
type Question struct {
	ID         uint        `json:"id"`
	Type       string      `json:"type"`
	Title      string      `json:"title"`
	Options    interface{} `json:"options"`    // 旧/新格式均可
	Answer     string      `json:"answer"`     // JSON 字符串
	Score      int         `json:"score"`
	NeedManual bool        `json:"needManual"` // 是否需要人工判分
}

// Result 单题判分结果
type Result struct {
	QuestionID uint   `json:"questionId"`
	Correct    bool   `json:"correct"`
	GotScore   int    `json:"gotScore"`
	FullScore  int    `json:"fullScore"`
	NeedManual bool   `json:"needManual"`
	Reason     string `json:"reason,omitempty"`
}

// PaperResult 整卷判分结果
type PaperResult struct {
	Results     []Result `json:"results"`
	TotalScore  int      `json:"totalScore"`  // 自动得分
	FullScore   int      `json:"fullScore"`   // 满分
	ManualCount int      `json:"manualCount"` // 待人工判分题数
	CorrectCnt  int      `json:"correctCnt"`  // 答对题数
}

// Grade 阅卷主入口
func Grade(questions []Question, answers map[string]interface{}) PaperResult {
	res := PaperResult{Results: make([]Result, 0, len(questions))}
	for _, q := range questions {
		ansVal, _ := answers[fmt.Sprintf("%d", q.ID)]
		r := gradeOne(q, ansVal)
		res.Results = append(res.Results, r)
		res.FullScore += q.Score
		if r.NeedManual {
			res.ManualCount++
		} else if r.Correct {
			res.TotalScore += r.GotScore
			res.CorrectCnt++
		}
	}
	return res
}

// gradeOne 单题判分
func gradeOne(q Question, ansVal interface{}) Result {
	r := Result{QuestionID: q.ID, FullScore: q.Score}

	// 主观题直接判为需人工
	if q.Type == "textarea" || q.NeedManual {
		r.NeedManual = true
		r.Correct = false
		r.Reason = "主观题，需人工判分"
		return r
	}

	correctAns, err := parseAnswer(q.Answer)
	if err != nil {
		r.NeedManual = true
		r.Reason = "答案解析失败: " + err.Error()
		return r
	}

	// checkbox 答案（数组）
	if q.Type == "checkbox" {
		var got []interface{}
		if arr, isArr := ansVal.([]interface{}); isArr {
			got = arr
		} else if s, isStr := ansVal.(string); isStr && s != "" {
			_ = json.Unmarshal([]byte(s), &got)
		}
		correctArr, isArr := correctAns.([]interface{})
		if !isArr {
			r.NeedManual = true
			r.Reason = "checkbox 答案应为数组"
			return r
		}
		if setEqual(got, correctArr) {
			r.Correct = true
			r.GotScore = q.Score
		} else {
			r.Correct = false
			r.Reason = fmt.Sprintf("答案不匹配，应为 %v", correctArr)
		}
		return r
	}

	// 数值题（容差 0.01）
	if q.Type == "number" {
		gotF, ok1 := toFloat(ansVal)
		expF, ok2 := toFloat(correctAns)
		if !ok1 || !ok2 {
			r.Correct = false
			r.Reason = "数值解析失败"
			return r
		}
		if abs(gotF-expF) < 0.01 {
			r.Correct = true
			r.GotScore = q.Score
		} else {
			r.Correct = false
			r.Reason = fmt.Sprintf("应为 %v", expF)
		}
		return r
	}

	// 其他（input/select/radio/...）：字符串相等（不区分大小写 + 修剪空白）
	got := strings.TrimSpace(fmt.Sprintf("%v", ansVal))
	exp := strings.TrimSpace(fmt.Sprintf("%v", correctAns))
	if strings.EqualFold(got, exp) {
		r.Correct = true
		r.GotScore = q.Score
	} else if got == "" {
		r.Correct = false
		r.Reason = "未作答"
	} else {
		r.Correct = false
		r.Reason = "答案不匹配"
	}
	return r
}

// parseAnswer 解析答案 JSON。无法解析时返回 error（调用方应标记 NeedManual）
func parseAnswer(ans string) (interface{}, error) {
	ans = strings.TrimSpace(ans)
	if ans == "" {
		return "", nil
	}
	if strings.HasPrefix(ans, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(ans), &arr); err != nil {
			return nil, err
		}
		return arr, nil
	}
	if strings.HasPrefix(ans, "{") {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(ans), &m); err != nil {
			return nil, err
		}
		return m, nil
	}
	// 尝试数字
	if f, err := strconv.ParseFloat(ans, 64); err == nil {
		return f, nil
	}
	// 尝试 JSON 字符串
	var s string
	if err := json.Unmarshal([]byte(ans), &s); err == nil {
		return s, nil
	}
	return nil, fmt.Errorf("answer is not valid JSON: %s", ans)
}

func setEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	ma := map[string]int{}
	for _, v := range a {
		ma[fmt.Sprintf("%v", v)]++
	}
	for _, v := range b {
		ma[fmt.Sprintf("%v", v)]--
	}
	for _, c := range ma {
		if c != 0 {
			return false
		}
	}
	return true
}

func toFloat(v interface{}) (float64, bool) {
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

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// QWithType 返回是否需要人工判分（用 formkit question 注册表判断）
func QWithType(typ string) bool {
	if typ == "textarea" {
		return true
	}
	// 复杂矩阵/文件上传/签名/位置等也归类为人工
	switch typ {
	case "matrixRadio", "matrixCheckbox", "matrixFill", "file", "signature", "location", "autopop":
		return true
	}
	_ = question.Get(typ) // 题型不存在时返回 nil，调用方自行决定
	return false
}
