package calc

import "fmt"

// ast 节点类型
type node interface{}

// numberNode 字面量数字
type numberNode struct{ v float64 }

// stringNode 字面量字符串
type stringNode struct{ v string }

// boolNode 字面量布尔
type boolNode struct{ v bool }

// nullNode null
type nullNode struct{}

// identNode 变量引用
type identNode struct{ name string }

// binaryNode 二元运算
type binaryNode struct {
	op       string
	left, right node
}

// unaryNode 一元运算
type unaryNode struct {
	op   string
	expr node
}

// callNode 函数调用
type callNode struct {
	name string
	args []node
}

// ternaryNode 三元 cond ? a : b
type ternaryNode struct {
	cond, a, b node
}

// parser 递归下降解析器
type parser struct {
	toks []token
	pos  int
}

func (p *parser) peek() token { return p.toks[p.pos] }

func (p *parser) consume() token {
	t := p.toks[p.pos]
	p.pos++
	return t
}

func (p *parser) matchOp(op string) bool {
	if p.peek().typ == tOp && p.peek().val == op {
		p.pos++
		return true
	}
	return false
}

func (p *parser) expectOp(op string) error {
	if p.peek().typ == tOp && p.peek().val == op {
		p.pos++
		return nil
	}
	return fmt.Errorf("expected %q at %d, got %q", op, p.peek().pos, p.peek().val)
}

// parseExpr = parseTernary
func (p *parser) parseExpr() (node, error) { return p.parseTernary() }

func (p *parser) parseTernary() (node, error) {
	cond, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.matchOp("?") {
		a, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expectOp(":"); err != nil {
			return nil, err
		}
		b, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ternaryNode{cond: cond, a: a, b: b}, nil
	}
	return cond, nil
}

func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.matchOp("||") {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{"||", left, right}
	}
	return left, nil
}

func (p *parser) parseAnd() (node, error) {
	left, err := p.parseEq()
	if err != nil {
		return nil, err
	}
	for p.matchOp("&&") {
		right, err := p.parseEq()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{"&&", left, right}
	}
	return left, nil
}

func (p *parser) parseEq() (node, error) {
	left, err := p.parseCmp()
	if err != nil {
		return nil, err
	}
	for {
		op := ""
		if p.matchOp("==") {
			op = "=="
		} else if p.matchOp("!=") {
			op = "!="
		} else {
			break
		}
		right, err := p.parseCmp()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op, left, right}
	}
	return left, nil
}

func (p *parser) parseCmp() (node, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	for {
		op := ""
		switch p.peek().val {
		case "<", ">", "<=", ">=":
			op = p.peek().val
			p.pos++
		default:
			return left, nil
		}
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op, left, right}
	}
}

func (p *parser) parseAdd() (node, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for {
		op := ""
		switch p.peek().val {
		case "+", "-":
			op = p.peek().val
			p.pos++
		default:
			return left, nil
		}
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op, left, right}
	}
}

func (p *parser) parseMul() (node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		op := ""
		switch p.peek().val {
		case "*", "/", "%":
			op = p.peek().val
			p.pos++
		default:
			return left, nil
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op, left, right}
	}
}

func (p *parser) parseUnary() (node, error) {
	if p.matchOp("-") || p.matchOp("!") {
		op := "-"
		if p.toks[p.pos-1].val == "!" {
			op = "!"
		}
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNode{op, expr}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (node, error) {
	t := p.peek()
	switch t.typ {
	case tNumber:
		p.pos++
		v, err := parseNumber(t.val)
		if err != nil {
			return nil, fmt.Errorf("bad number %q at %d", t.val, t.pos)
		}
		return &numberNode{v}, nil
	case tString:
		p.pos++
		return &stringNode{t.val}, nil
	case tBool:
		p.pos++
		return &boolNode{t.val == "true"}, nil
	case tNull:
		p.pos++
		return &nullNode{}, nil
	case tIdent:
		p.pos++
		// 函数调用？
		if p.peek().typ == tLParen {
			p.pos++
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			return &callNode{t.val, args}, nil
		}
		return &identNode{t.val}, nil
	case tLParen:
		p.pos++
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tRParen {
			return nil, fmt.Errorf("expected ')' at %d", p.peek().pos)
		}
		p.pos++
		return expr, nil
	}
	return nil, fmt.Errorf("unexpected token %q at %d", t.val, t.pos)
}

func (p *parser) parseArgs() ([]node, error) {
	if p.peek().typ == tRParen {
		p.pos++
		return nil, nil
	}
	var args []node
	for {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.peek().typ == tComma {
			p.pos++
			continue
		}
		break
	}
	if p.peek().typ != tRParen {
		return nil, fmt.Errorf("expected ')' or ',' at %d", p.peek().pos)
	}
	p.pos++
	return args, nil
}
