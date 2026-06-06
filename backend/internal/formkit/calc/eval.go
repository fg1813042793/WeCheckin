package calc

import (
	"fmt"
	"strings"
)

// evalNode 求值 AST
func evalNode(n node, env map[string]interface{}) (interface{}, error) {
	switch x := n.(type) {
	case *numberNode:
		return x.v, nil
	case *stringNode:
		return x.v, nil
	case *boolNode:
		return x.v, nil
	case *nullNode:
		return nil, nil
	case *identNode:
		if env == nil {
			return nil, nil
		}
		return env[x.name], nil
	case *unaryNode:
		v, err := evalNode(x.expr, env)
		if err != nil {
			return nil, err
		}
		switch x.op {
		case "-":
			f, ok := toFloat(v)
			if !ok {
				return nil, fmt.Errorf("cannot negate non-number %v", v)
			}
			return -f, nil
		case "!":
			return !toBool(v), nil
		}
	case *binaryNode:
		l, err := evalNode(x.left, env)
		if err != nil {
			return nil, err
		}
		r, err := evalNode(x.right, env)
		if err != nil {
			return nil, err
		}
		return evalBinary(x.op, l, r)
	case *ternaryNode:
		c, err := evalNode(x.cond, env)
		if err != nil {
			return nil, err
		}
		if toBool(c) {
			return evalNode(x.a, env)
		}
		return evalNode(x.b, env)
	case *callNode:
		return evalCall(x.name, x.args, env)
	}
	return nil, fmt.Errorf("unsupported node %T", n)
}

func evalBinary(op string, l, r interface{}) (interface{}, error) {
	switch op {
	case "+":
		// 字符串拼接
		if ls, ok := l.(string); ok {
			return ls + toStr(r), nil
		}
		if rs, ok := r.(string); ok {
			return toStr(l) + rs, nil
		}
		lf, lok := toFloat(l)
		rf, rok := toFloat(r)
		if !lok || !rok {
			return nil, fmt.Errorf("cannot add %v and %v", l, r)
		}
		return lf + rf, nil
	case "-", "*", "/", "%":
		lf, lok := toFloat(l)
		rf, rok := toFloat(r)
		if !lok || !rok {
			return nil, fmt.Errorf("cannot %s %v and %v", op, l, r)
		}
		switch op {
		case "-":
			return lf - rf, nil
		case "*":
			return lf * rf, nil
		case "/":
			if rf == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return lf / rf, nil
		case "%":
			// 整数取模
			li, _ := toInt(lf)
			ri, _ := toInt(rf)
			if ri == 0 {
				return nil, fmt.Errorf("modulo by zero")
			}
			return float64(li % ri), nil
		}
	case "<", ">", "<=", ">=":
		cf, ok := compare(l, r)
		if !ok {
			return nil, fmt.Errorf("cannot compare %v and %v", l, r)
		}
		switch op {
		case "<":
			return cf < 0, nil
		case ">":
			return cf > 0, nil
		case "<=":
			return cf <= 0, nil
		case ">=":
			return cf >= 0, nil
		}
	case "==":
		return eq(l, r), nil
	case "!=":
		return !eq(l, r), nil
	case "&&":
		return toBool(l) && toBool(r), nil
	case "||":
		return toBool(l) || toBool(r), nil
	}
	return nil, fmt.Errorf("unsupported op %q", op)
}

func evalCall(name string, args []node, env map[string]interface{}) (interface{}, error) {
	evaled := make([]interface{}, 0, len(args))
	for _, a := range args {
		v, err := evalNode(a, env)
		if err != nil {
			return nil, err
		}
		evaled = append(evaled, v)
	}
	switch name {
	case "contains":
		if len(evaled) != 2 {
			return nil, fmt.Errorf("contains(s, sub) takes 2 args")
		}
		return contains(evaled[0], evaled[1]), nil
	case "empty":
		if len(evaled) != 1 {
			return nil, fmt.Errorf("empty(v) takes 1 arg")
		}
		return isEmpty(evaled[0]), nil
	case "len":
		if len(evaled) != 1 {
			return nil, fmt.Errorf("len(v) takes 1 arg")
		}
		return lengthOf(evaled[0]), nil
	case "if":
		if len(evaled) != 3 {
			return nil, fmt.Errorf("if(cond, a, b) takes 3 args")
		}
		if toBool(evaled[0]) {
			return evaled[1], nil
		}
		return evaled[2], nil
	case "sum":
		sum := 0.0
		for _, v := range evaled {
			f, ok := toFloat(v)
			if !ok {
				return nil, fmt.Errorf("sum: cannot convert %v to number", v)
			}
			sum += f
		}
		return sum, nil
	case "avg":
		if len(evaled) == 0 {
			return nil, fmt.Errorf("avg: no args")
		}
		sum := 0.0
		for _, v := range evaled {
			f, ok := toFloat(v)
			if !ok {
				return nil, fmt.Errorf("avg: cannot convert %v to number", v)
			}
			sum += f
		}
		return sum / float64(len(evaled)), nil
	}
	return nil, fmt.Errorf("unknown function %q", name)
}

func contains(haystack, needle interface{}) bool {
	hs := toStr(haystack)
	ns := toStr(needle)
	return strings.Contains(hs, ns)
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	if arr, ok := v.([]interface{}); ok {
		return len(arr) == 0
	}
	return false
}

func lengthOf(v interface{}) float64 {
	switch x := v.(type) {
	case string:
		return float64(len(x))
	case []interface{}:
		return float64(len(x))
	case map[string]interface{}:
		return float64(len(x))
	}
	return 0
}

func toFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0, true
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case string:
		// 尝试解析
		var f float64
		if _, err := fmt.Sscanf(x, "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toInt(f float64) (int, int) {
	return int(f), 0
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if f, ok := toFloat(v); ok {
		// 整数不打小数点
		if f == float64(int(f)) {
			return fmt.Sprintf("%d", int(f))
		}
		return fmt.Sprintf("%g", f)
	}
	if b, ok := v.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return fmt.Sprintf("%v", v)
}

// compare 返回 -1/0/1，不支持时返回 false
func compare(l, r interface{}) (int, bool) {
	if lf, ok := toFloat(l); ok {
		if rf, ok2 := toFloat(r); ok2 {
			if lf < rf {
				return -1, true
			}
			if lf > rf {
				return 1, true
			}
			return 0, true
		}
	}
	ls, rs := toStr(l), toStr(r)
	if ls < rs {
		return -1, true
	}
	if ls > rs {
		return 1, true
	}
	return 0, true
}

func eq(l, r interface{}) bool {
	// 数字比较
	if lf, ok := toFloat(l); ok {
		if rf, ok2 := toFloat(r); ok2 {
			return lf == rf
		}
	}
	// 字符串比较
	ls, rs := toStr(l), toStr(r)
	return ls == rs
}
