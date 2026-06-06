package calc

import (
	"testing"
)

func TestEval_Arithmetic(t *testing.T) {
	e := New()
	tests := []struct {
		expr string
		env  map[string]interface{}
		want interface{}
	}{
		{"1 + 2", nil, float64(3)},
		{"10 - 3", nil, float64(7)},
		{"6 * 7", nil, float64(42)},
		{"10 / 4", nil, float64(2.5)},
		{"10 % 3", nil, float64(1)},
		{"-5 + 3", nil, float64(-2)},
		{"2 + 3 * 4", nil, float64(14)},
		{"(2 + 3) * 4", nil, float64(20)},
	}
	for _, tt := range tests {
		got, err := e.Eval(tt.expr, tt.env)
		if err != nil {
			t.Errorf("Eval(%q) err: %v", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Eval(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEval_Variable(t *testing.T) {
	e := New()
	got, err := e.Eval("q1 + q2", map[string]interface{}{"q1": 10.0, "q2": 5.0})
	if err != nil {
		t.Fatal(err)
	}
	if got != float64(15) {
		t.Errorf("got %v, want 15", got)
	}
}

func TestEval_Comparison(t *testing.T) {
	e := New()
	tests := []struct {
		expr string
		want bool
	}{
		{"1 < 2", true},
		{"2 < 1", false},
		{"3 == 3", true},
		{"3 != 3", false},
		{"3 >= 3", true},
		{"3 <= 3", true},
		{`"abc" == "abc"`, true},
		{`"abc" != "def"`, true},
	}
	for _, tt := range tests {
		got, err := e.EvalBool(tt.expr, nil)
		if err != nil {
			t.Errorf("EvalBool(%q) err: %v", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("EvalBool(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEval_Logical(t *testing.T) {
	e := New()
	tests := []struct {
		expr string
		want bool
	}{
		{"true && false", false},
		{"true || false", true},
		{"!false", true},
		{"1 < 2 && 3 < 4", true},
		{"1 < 2 && 3 > 4", false},
		{"(1 < 2) || (3 > 4)", true},
	}
	for _, tt := range tests {
		got, err := e.EvalBool(tt.expr, nil)
		if err != nil {
			t.Errorf("EvalBool(%q) err: %v", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("EvalBool(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEval_Ternary(t *testing.T) {
	e := New()
	got, err := e.Eval(`1 < 2 ? "yes" : "no"`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "yes" {
		t.Errorf("got %v, want yes", got)
	}
	got, err = e.Eval(`1 > 2 ? "yes" : "no"`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "no" {
		t.Errorf("got %v, want no", got)
	}
}

func TestEval_StringConcat(t *testing.T) {
	e := New()
	got, err := e.Eval(`"hello" + " " + "world"`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello world" {
		t.Errorf("got %v", got)
	}
}

func TestEval_FuncContains(t *testing.T) {
	e := New()
	got, err := e.EvalBool(`contains("hello world", "world")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true")
	}
	got, err = e.EvalBool(`contains("hello", "xyz")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected false")
	}
}

func TestEval_FuncEmpty(t *testing.T) {
	e := New()
	tests := []struct {
		expr string
		want bool
	}{
		{`empty("")`, true},
		{`empty("x")`, false},
		{`empty(null)`, true},
	}
	for _, tt := range tests {
		got, err := e.EvalBool(tt.expr, nil)
		if err != nil {
			t.Errorf("EvalBool(%q) err: %v", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("EvalBool(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEval_FuncIf(t *testing.T) {
	e := New()
	got, err := e.Eval(`if(1 < 2, "yes", "no")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "yes" {
		t.Errorf("got %v", got)
	}
}

func TestEval_FuncSumAvg(t *testing.T) {
	e := New()
	got, _ := e.Eval(`sum(1, 2, 3, 4)`, nil)
	if got != float64(10) {
		t.Errorf("sum: got %v", got)
	}
	got, _ = e.Eval(`avg(2, 4, 6)`, nil)
	if got != float64(4) {
		t.Errorf("avg: got %v", got)
	}
}

func TestEval_VariableInExpr(t *testing.T) {
	e := New()
	got, err := e.Eval(`q1 * 0.1`, map[string]interface{}{"q1": 100.0})
	if err != nil {
		t.Fatal(err)
	}
	if got != float64(10) {
		t.Errorf("got %v, want 10", got)
	}
}

func TestEval_StringEqualityWithVar(t *testing.T) {
	e := New()
	env := map[string]interface{}{"q1": "是"}
	got, err := e.EvalBool(`q1 == "是"`, env)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true")
	}
}

func TestEval_UndefinedVar(t *testing.T) {
	e := New()
	got, err := e.Eval(`q1 + 0`, map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	// q1 未定义 → nil → 0
	if got != float64(0) {
		t.Errorf("undefined var should default to 0, got %v", got)
	}
}

func TestEval_EmptyExpr(t *testing.T) {
	e := New()
	got, err := e.Eval("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("empty should return nil, got %v", got)
	}
}

func TestEval_InvalidExpr(t *testing.T) {
	e := New()
	if _, err := e.Eval("1 +", nil); err == nil {
		t.Error("expected error for invalid expr")
	}
	if _, err := e.Eval("@@@", nil); err == nil {
		t.Error("expected error for invalid tokens")
	}
}

func TestEval_StringWithEscape(t *testing.T) {
	e := New()
	got, err := e.Eval(`"a\"b"`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != `a"b` {
		t.Errorf("got %q", got)
	}
}

func TestEval_FuncLen(t *testing.T) {
	e := New()
	got, _ := e.Eval(`len("hello")`, nil)
	if got != float64(5) {
		t.Errorf("got %v", got)
	}
}
