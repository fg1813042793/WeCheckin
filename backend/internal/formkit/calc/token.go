package calc

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// tokenType 词法单元类型
type tokenType int

const (
	tEOF tokenType = iota
	tNumber
	tString
	tIdent
	tBool
	tNull
	tOp     // 操作符
	tLParen // (
	tRParen // )
	tComma  // ,
	tDot    // .
)

type token struct {
	typ tokenType
	val string
	pos int
}

// tokenize 把 expr 切成 token 流
func tokenize(expr string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(expr) {
		c := expr[i]
		// 跳过空白
		if unicode.IsSpace(rune(c)) {
			i++
			continue
		}
		// 数字
		if c >= '0' && c <= '9' || c == '.' {
			start := i
			hasDot := false
			for i < len(expr) {
				ch := expr[i]
				if ch >= '0' && ch <= '9' {
					i++
				} else if ch == '.' && !hasDot {
					hasDot = true
					i++
				} else {
					break
				}
			}
			toks = append(toks, token{tNumber, expr[start:i], start})
			continue
		}
		// 字符串 "..."
		if c == '"' || c == '\'' {
			quote := c
			i++
			start := i
			var sb strings.Builder
			for i < len(expr) && expr[i] != quote {
				if expr[i] == '\\' && i+1 < len(expr) {
					// escape
					sb.WriteByte(expr[i+1])
					i += 2
					continue
				}
				sb.WriteByte(expr[i])
				i++
			}
			if i >= len(expr) {
				return nil, fmt.Errorf("unterminated string at %d", start)
			}
			i++ // skip closing quote
			toks = append(toks, token{tString, sb.String(), start})
			continue
		}
		// 标识符 / 关键字
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			start := i
			for i < len(expr) {
				ch := expr[i]
				if ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
					i++
				} else {
					break
				}
			}
			word := expr[start:i]
			switch word {
			case "true", "false":
				toks = append(toks, token{tBool, word, start})
			case "null", "nil":
				toks = append(toks, token{tNull, word, start})
			default:
				toks = append(toks, token{tIdent, word, start})
			}
			continue
		}
		// 操作符 (1-2 字符)
		if isOpStart(c) {
			start := i
			op := string(c)
			if i+1 < len(expr) && isOpCont(expr[i+1]) {
				op += string(expr[i+1])
				i += 2
			} else {
				i++
			}
			toks = append(toks, token{tOp, op, start})
			continue
		}
		// 标点
		switch c {
		case '(':
			toks = append(toks, token{tLParen, "(", i})
			i++
		case ')':
			toks = append(toks, token{tRParen, ")", i})
			i++
		case ',':
			toks = append(toks, token{tComma, ",", i})
			i++
		default:
			return nil, fmt.Errorf("unexpected char %q at %d", c, i)
		}
	}
	toks = append(toks, token{tEOF, "", len(expr)})
	return toks, nil
}

func isOpStart(c byte) bool {
	switch c {
	case '+', '-', '*', '/', '%', '=', '!', '<', '>', '&', '|', '?', ':':
		return true
	}
	return false
}

func isOpCont(c byte) bool {
	switch c {
	case '=', '&', '|':
		return true
	}
	return false
}

// parseNumber 辅助：把字符串转 float64
func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
