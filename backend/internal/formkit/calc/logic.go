package calc

import (
	"fmt"
	"strings"

	"wecheckin-backend/backend/internal/formkit/schema"
)

// Engine 扩展：Logic + CalcValue 专用 API

// EvalLogic 评估 LogicCondition 在 answers 上的真假。
// 内部把 answers 拍平到 env，把 condition 编译为简单表达式。
func (e *Engine) EvalLogic(cond schema.LogicCondition, answers map[string]interface{}) (bool, error) {
	if cond.QuestionID == "" {
		return false, fmt.Errorf("logic condition: missing questionId")
	}
	env := map[string]interface{}{}
	for k, v := range answers {
		env[k] = v
	}
	switch cond.Operator {
	case "empty":
		return isEmpty(answers[cond.QuestionID]), nil
	case "notEmpty":
		return !isEmpty(answers[cond.QuestionID]), nil
	case "==", "!=", ">", "<", ">=", "<=":
		expr := cond.QuestionID + " " + cond.Operator + " " + formatLiteral(cond.Value)
		v, err := e.Eval(expr, env)
		if err != nil {
			return false, err
		}
		return toBool(v), nil
	case "contains":
		haystack := answers[cond.QuestionID]
		return contains(haystack, cond.Value), nil
	case "notContains":
		return !contains(answers[cond.QuestionID], cond.Value), nil
	}
	return false, fmt.Errorf("unsupported operator %q", cond.Operator)
}

// formatLiteral 把 interface{} 字面量为 expr 字面量字符串
func formatLiteral(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return `"` + strings.ReplaceAll(x, `"`, `\"`) + `"`
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return toStr(v)
}

// ApplyLogic 评估 schema 中所有 logic rules，合并每个 question 的最终状态。
// 返回 qid -> 状态集合（用逗号分隔，例如 "hide,require"）。
// 状态集：show / hide / require / optional
//
// 同一 question 的多条 rule 都满足时，状态按 union 合并（hide > show, require > optional）。
func (e *Engine) ApplyLogic(s *schema.FormSchema, answers map[string]interface{}) (map[string]string, error) {
	out := map[string]string{}
	if s == nil {
		return out, nil
	}
	for i := range s.Questions {
		q := s.Questions[i]
		for _, rule := range q.Logic {
			match, err := e.EvalLogic(rule.When, answers)
			if err != nil {
				// 单条 rule 评估失败不中断整个流程
				continue
			}
			if !match {
				continue
			}
			t := rule.Target
			if t == "" {
				t = q.ID
			}
			out[t] = mergeState(out[t], rule.Action)
		}
	}
	return out, nil
}

// ApplyCalcValues 在 answers 上跑所有 CalcValue.Expr，把结果写回 answers。
// 如果 CalcValue.Target 显式指定，则写回 answers[Target]；否则写回 answers[q.ID]。
// 注意：本函数返回新 map，不修改入参。
func (e *Engine) ApplyCalcValues(s *schema.FormSchema, answers map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(answers))
	for k, v := range answers {
		out[k] = v
	}
	if s == nil {
		return out, nil
	}
	for i := range s.Questions {
		q := s.Questions[i]
		if q.CalcValue == nil || q.CalcValue.Expr == "" {
			continue
		}
		v, err := e.Eval(q.CalcValue.Expr, out)
		if err != nil {
			// 单条 calc 评估失败不中断
			continue
		}
		target := q.CalcValue.Target
		if target == "" {
			target = q.ID
		}
		out[target] = v
	}
	return out, nil
}

// mergeState 合并两条状态：hide 覆盖 show；require 覆盖 optional。
// 输入空字符串视为无状态。
func mergeState(existing, newState string) string {
	if existing == "" {
		return newState
	}
	set := map[string]bool{}
	for _, s := range strings.Split(existing, ",") {
		set[strings.TrimSpace(s)] = true
	}
	set[newState] = true
	// 冲突消解
	if set["hide"] {
		set["show"] = false
	}
	if set["require"] {
		set["optional"] = false
	}
	// 序列化（按固定顺序）
	parts := []string{}
	for _, s := range []string{"hide", "show", "require", "optional"} {
		if set[s] {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ",")
}
