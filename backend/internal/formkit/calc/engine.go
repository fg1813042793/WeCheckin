// Package calc 实现 formkit 表达式引擎。
//
// 不依赖第三方库，自实现一个支持算术/比较/逻辑/三元/函数的表达式求值器。
// 语法（按优先级高→低）：
//   1) 字面量：数字 / 字符串（"..."） / 布尔 (true/false) / null
//   2) 标识符 (env 变量查找) / 函数调用
//   3) 一元: - !
//   4) 乘除模: * / %
//   5) 加减: + -
//   6) 比较: < > <= >=
//   7) 相等: == !=
//   8) 逻辑与: &&
//   9) 逻辑或: ||
//  10) 三元: ? :
//
// 内置函数：contains, empty, len, if, sum, avg
//
// 引擎不引用任何外部包，所有求值在 sandbox 内进行。
package calc

import (
	"fmt"
	"strings"
)

// Engine 表达式求值器
type Engine struct{}

// New 返回新引擎
func New() *Engine { return &Engine{} }

// Eval 求值 expr，env 提供变量。
// env 中的每个 key 都可以在表达式里作为变量引用。
func (e *Engine) Eval(expr string, env map[string]interface{}) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	toks, err := tokenize(expr)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.peek().typ != tEOF {
		return nil, fmt.Errorf("trailing tokens at %d", p.peek().pos)
	}
	return evalNode(node, env)
}

// EvalBool 把 expr 当布尔表达式求值（>= 0.5 / 非空字符串 / 非零数字 → true）
func (e *Engine) EvalBool(expr string, env map[string]interface{}) (bool, error) {
	v, err := e.Eval(expr, env)
	if err != nil {
		return false, err
	}
	return toBool(v), nil
}

// toBool 转布尔：nil/0/false/"false"/"" → false；其他 → true
func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case int:
		return x != 0
	case int64:
		return x != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		return s != "" && s != "false" && s != "0"
	}
	return true
}
