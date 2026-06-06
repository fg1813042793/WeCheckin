package calc

import (
	"testing"

	"wecheckin-backend/backend/internal/formkit/schema"
)

func TestEvalLogic(t *testing.T) {
	e := New()
	tests := []struct {
		name string
		cond schema.LogicCondition
		ans  map[string]interface{}
		want bool
	}{
		{"empty_string", schema.LogicCondition{QuestionID: "q1", Operator: "empty"}, map[string]interface{}{"q1": ""}, true},
		{"notEmpty", schema.LogicCondition{QuestionID: "q1", Operator: "notEmpty"}, map[string]interface{}{"q1": "x"}, true},
		{"== string", schema.LogicCondition{QuestionID: "q1", Operator: "==", Value: "是"}, map[string]interface{}{"q1": "是"}, true},
		{"!= string", schema.LogicCondition{QuestionID: "q1", Operator: "!=", Value: "否"}, map[string]interface{}{"q1": "是"}, true},
		{"> number", schema.LogicCondition{QuestionID: "q1", Operator: ">", Value: 5}, map[string]interface{}{"q1": 10.0}, true},
		{"< number", schema.LogicCondition{QuestionID: "q1", Operator: "<", Value: 5}, map[string]interface{}{"q1": 10.0}, false},
		{"contains", schema.LogicCondition{QuestionID: "q1", Operator: "contains", Value: "world"}, map[string]interface{}{"q1": "hello world"}, true},
		{"notContains", schema.LogicCondition{QuestionID: "q1", Operator: "notContains", Value: "xyz"}, map[string]interface{}{"q1": "abc"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.EvalLogic(tt.cond, tt.ans)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyLogic_Hide(t *testing.T) {
	e := New()
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "q1", Type: "radio", Title: "是否学生"},
			{ID: "q2", Type: "input", Title: "学校", Logic: []schema.LogicRule{{
				When:   schema.LogicCondition{QuestionID: "q1", Operator: "==", Value: "是"},
				Action: "hide",
				Target: "q2",
			}}},
		},
	}
	ans := map[string]interface{}{"q1": "是"}
	got, err := e.ApplyLogic(s, ans)
	if err != nil {
		t.Fatal(err)
	}
	if got["q2"] != "hide" {
		t.Errorf("expected q2=hide, got %q", got["q2"])
	}

	ans = map[string]interface{}{"q1": "否"}
	got, _ = e.ApplyLogic(s, ans)
	if _, ok := got["q2"]; ok {
		t.Errorf("q2 should not have state, got %q", got["q2"])
	}
}

func TestApplyLogic_Require(t *testing.T) {
	e := New()
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "q1", Type: "select", Title: "类型"},
			{ID: "q2", Type: "input", Title: "说明", Logic: []schema.LogicRule{{
				When:   schema.LogicCondition{QuestionID: "q1", Operator: "==", Value: "其他"},
				Action: "require",
				Target: "q2",
			}}},
		},
	}
	ans := map[string]interface{}{"q1": "其他"}
	got, _ := e.ApplyLogic(s, ans)
	if got["q2"] != "require" {
		t.Errorf("expected q2=require, got %q", got["q2"])
	}
}

func TestApplyLogic_ConflictResolution(t *testing.T) {
	e := New()
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "q1", Type: "input"},
			{ID: "q2", Type: "input", Logic: []schema.LogicRule{
				{When: schema.LogicCondition{QuestionID: "q1", Operator: "==", Value: "1"}, Action: "hide", Target: "q2"},
				{When: schema.LogicCondition{QuestionID: "q1", Operator: "==", Value: "1"}, Action: "show", Target: "q2"},
			}},
		},
	}
	ans := map[string]interface{}{"q1": "1"}
	got, _ := e.ApplyLogic(s, ans)
	// hide 应覆盖 show
	if got["q2"] == "" || (got["q2"] != "hide" && !contains_(got["q2"], "hide")) {
		t.Errorf("expected hide wins, got %q", got["q2"])
	}
}

func contains_(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestApplyCalcValues(t *testing.T) {
	e := New()
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "price", Type: "number"},
			{ID: "qty", Type: "number"},
			{ID: "total", Type: "number", CalcValue: &schema.CalcValue{Expr: "price * qty"}},
			{ID: "discount", Type: "number", CalcValue: &schema.CalcValue{Expr: "total * 0.9", Target: "discount"}},
		},
	}
	ans := map[string]interface{}{"price": 100.0, "qty": 3.0}
	out, err := e.ApplyCalcValues(s, ans)
	if err != nil {
		t.Fatal(err)
	}
	if out["total"] != float64(300) {
		t.Errorf("total: got %v, want 300", out["total"])
	}
	if out["discount"] != float64(270) {
		t.Errorf("discount: got %v, want 270", out["discount"])
	}
	// 原 answers 不变
	if _, ok := ans["total"]; ok {
		t.Error("original answers should not be modified")
	}
}

func TestApplyCalcValues_InvalidExpr_Skip(t *testing.T) {
	e := New()
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "q1", Type: "input", CalcValue: &schema.CalcValue{Expr: "@@@"}}, // 编译失败
			{ID: "q2", Type: "input"},
		},
	}
	ans := map[string]interface{}{"q1": "x"}
	out, err := e.ApplyCalcValues(s, ans)
	if err != nil {
		t.Fatal(err)
	}
	if out["q1"] != "x" {
		t.Errorf("invalid calc should not overwrite, got %v", out["q1"])
	}
}

func TestApplyLogic_MultiTarget(t *testing.T) {
	e := New()
	// 一条 rule 影响多个 target
	s := &schema.FormSchema{
		Version: "2.0",
		Questions: []schema.Question{
			{ID: "q1", Type: "input"},
			{ID: "q2", Type: "input", Logic: []schema.LogicRule{
				{When: schema.LogicCondition{QuestionID: "q1", Operator: "notEmpty"}, Action: "require", Target: "q2"},
				{When: schema.LogicCondition{QuestionID: "q1", Operator: "notEmpty"}, Action: "require", Target: "q3"},
			}},
		},
	}
	ans := map[string]interface{}{"q1": "x"}
	got, _ := e.ApplyLogic(s, ans)
	if got["q2"] != "require" {
		t.Errorf("q2 should be require, got %q", got["q2"])
	}
	if got["q3"] != "require" {
		t.Errorf("q3 should be require, got %q", got["q3"])
	}
}
