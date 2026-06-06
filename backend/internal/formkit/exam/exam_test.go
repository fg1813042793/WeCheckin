package exam

import "testing"

func TestGrade_Input(t *testing.T) {
	qs := []Question{{ID: 1, Type: "input", Title: "首都", Answer: `"北京"`, Score: 10}}
	res := Grade(qs, map[string]interface{}{"1": "北京"})
	if res.TotalScore != 10 {
		t.Errorf("got total %d, want 10", res.TotalScore)
	}
	if !res.Results[0].Correct {
		t.Errorf("expected correct")
	}
}

func TestGrade_Radio(t *testing.T) {
	qs := []Question{{ID: 1, Type: "radio", Title: "选择", Answer: `"A"`, Score: 5}}
	res := Grade(qs, map[string]interface{}{"1": "A"})
	if !res.Results[0].Correct {
		t.Errorf("A should be correct")
	}
	res = Grade(qs, map[string]interface{}{"1": "B"})
	if res.Results[0].Correct {
		t.Errorf("B should be wrong")
	}
}

func TestGrade_Checkbox(t *testing.T) {
	qs := []Question{{ID: 1, Type: "checkbox", Title: "多选", Answer: `["A","B"]`, Score: 5}}
	// 顺序无关
	res := Grade(qs, map[string]interface{}{"1": []interface{}{"B", "A"}})
	if !res.Results[0].Correct {
		t.Errorf("unordered should still be correct: %s", res.Results[0].Reason)
	}
	// 缺少一个
	res = Grade(qs, map[string]interface{}{"1": []interface{}{"A"}})
	if res.Results[0].Correct {
		t.Errorf("partial should be wrong")
	}
}

func TestGrade_Number(t *testing.T) {
	qs := []Question{{ID: 1, Type: "number", Title: "1+1", Answer: `2`, Score: 5}}
	res := Grade(qs, map[string]interface{}{"1": 2.0})
	if !res.Results[0].Correct {
		t.Errorf("expected correct")
	}
	res = Grade(qs, map[string]interface{}{"1": 2.005})
	if !res.Results[0].Correct {
		t.Errorf("within tolerance should be correct")
	}
	res = Grade(qs, map[string]interface{}{"1": 3.0})
	if res.Results[0].Correct {
		t.Errorf("3 should be wrong")
	}
}

func TestGrade_Textarea_NeedsManual(t *testing.T) {
	qs := []Question{{ID: 1, Type: "textarea", Title: "论述", Answer: `""`, Score: 10}}
	res := Grade(qs, map[string]interface{}{"1": "很好"})
	if !res.Results[0].NeedManual {
		t.Error("textarea should need manual")
	}
	if res.TotalScore != 0 {
		t.Errorf("manual should not count to total, got %d", res.TotalScore)
	}
}

func TestGrade_EmptyAnswer(t *testing.T) {
	qs := []Question{{ID: 1, Type: "input", Title: "x", Answer: `"yes"`, Score: 5}}
	res := Grade(qs, map[string]interface{}{"1": ""})
	if res.Results[0].Correct {
		t.Error("empty should be wrong")
	}
	if res.Results[0].Reason == "" {
		t.Error("reason should be set")
	}
}

func TestGrade_MultiQuestion(t *testing.T) {
	qs := []Question{
		{ID: 1, Type: "radio", Answer: `"A"`, Score: 10},
		{ID: 2, Type: "number", Answer: `42`, Score: 20},
		{ID: 3, Type: "textarea", Answer: `""`, Score: 30}, // 主观
	}
	ans := map[string]interface{}{"1": "A", "2": 42.0, "3": "长篇论述"}
	res := Grade(qs, ans)
	if res.FullScore != 60 {
		t.Errorf("full: got %d, want 60", res.FullScore)
	}
	if res.TotalScore != 30 {
		t.Errorf("total: got %d, want 30 (10+20)", res.TotalScore)
	}
	if res.CorrectCnt != 2 {
		t.Errorf("correct count: got %d, want 2", res.CorrectCnt)
	}
	if res.ManualCount != 1 {
		t.Errorf("manual count: got %d, want 1", res.ManualCount)
	}
}

func TestGrade_AnswerAsNumberString(t *testing.T) {
	qs := []Question{{ID: 1, Type: "input", Answer: `42`, Score: 5}}
	res := Grade(qs, map[string]interface{}{"1": "42"})
	if !res.Results[0].Correct {
		t.Errorf("string number should match: %s", res.Results[0].Reason)
	}
}

func TestGrade_InvalidJSON(t *testing.T) {
	qs := []Question{{ID: 1, Type: "input", Answer: `not json`, Score: 5}}
	res := Grade(qs, map[string]interface{}{"1": "x"})
	// not json → 原样字符串 "not json"
	if !res.Results[0].NeedManual {
		t.Error("invalid json should fall back to manual")
	}
}
