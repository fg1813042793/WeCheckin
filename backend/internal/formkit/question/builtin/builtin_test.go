package builtin

import (
	"encoding/json"
	"testing"

	"wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/schema"
)

func TestRegisteredCount(t *testing.T) {
	types := question.Types()
	if len(types) < 8 {
		t.Fatalf("expected >=8 registered types, got %d (%v)", len(types), types)
	}
	expected := []string{"input", "text", "textarea", "number", "select", "radio", "checkbox", "date", "file"}
	for _, e := range expected {
		if !question.Has(e) {
			t.Errorf("missing type %q", e)
		}
	}
}

func TestInput_Validate(t *testing.T) {
	q := schema.Question{ID: "q1", Type: "input", Title: "姓名", Required: true}
	inst := question.Get("input")
	if inst == nil {
		t.Fatal("input not registered")
	}
	if err := inst.Validate(nil, q); err == nil {
		t.Error("expected required error")
	}
	if err := inst.Validate("张三", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate(123, q); err == nil {
		t.Error("expected type error for non-string")
	}
}

func TestInput_Pattern(t *testing.T) {
	q := schema.Question{
		ID: "q1", Type: "input", Title: "邮箱",
		Validate: []schema.ValidateRule{{Type: "pattern", Value: `^[\w.+-]+@[\w-]+\.[\w.]+$`, Message: "邮箱格式不对"}},
	}
	inst := question.Get("input")
	if err := inst.Validate("not-email", q); err == nil {
		t.Error("expected pattern fail")
	}
	if err := inst.Validate("a@b.com", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
}

func TestTextarea_MaxLength(t *testing.T) {
	q := schema.Question{
		ID: "q1", Type: "textarea", Title: "备注",
		Validate: []schema.ValidateRule{{Type: "maxLength", Value: float64(5), Message: "太长"}},
	}
	inst := question.Get("textarea")
	if err := inst.Validate("123456", q); err == nil {
		t.Error("expected maxLength fail")
	}
	if err := inst.Validate("hi", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
}

func TestNumber_Validate(t *testing.T) {
	props := json.RawMessage(`{"min":0,"max":150}`)
	q := schema.Question{ID: "q1", Type: "number", Title: "年龄", Required: true, Props: props}
	inst := question.Get("number")
	if err := inst.Validate(nil, q); err == nil {
		t.Error("expected required error")
	}
	if err := inst.Validate(20, q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate("abc", q); err == nil {
		t.Error("expected parse fail")
	}
	if err := inst.Validate(200, q); err == nil {
		t.Error("expected max fail")
	}
	if err := inst.Validate(-1, q); err == nil {
		t.Error("expected min fail")
	}
}

func TestNumber_Format(t *testing.T) {
	inst := question.Get("number").(*NumberQuestion)
	if got := inst.FormatValue(3.14, schema.Question{}); got != "3.14" {
		t.Errorf("format float wrong: %q", got)
	}
	if got := inst.FormatValue(nil, schema.Question{}); got != "" {
		t.Errorf("nil should be empty, got %q", got)
	}
}

func TestSelect_Validate(t *testing.T) {
	props := json.RawMessage(`{"options":[{"label":"A","value":"a"},{"label":"B","value":"b"}]}`)
	q := schema.Question{ID: "q1", Type: "select", Title: "类型", Required: true, Props: props}
	inst := question.Get("select")
	if err := inst.Validate("a", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate("c", q); err == nil {
		t.Error("expected invalid value")
	}
}

func TestSelect_Format(t *testing.T) {
	props := json.RawMessage(`{"options":[{"label":"苹果","value":"apple"}]}`)
	q := schema.Question{ID: "q1", Type: "select", Props: props}
	inst := question.Get("select")
	if got := inst.FormatValue("apple", q); got != "苹果" {
		t.Errorf("expected label '苹果', got %q", got)
	}
	if got := inst.FormatValue("unknown", q); got != "unknown" {
		t.Errorf("fallback to value, got %q", got)
	}
}

func TestCheckbox_Validate(t *testing.T) {
	props := json.RawMessage(`{"options":[{"label":"篮球","value":"basketball"},{"label":"足球","value":"soccer"}]}`)
	q := schema.Question{ID: "q1", Type: "checkbox", Title: "爱好", Required: true, Props: props}
	inst := question.Get("checkbox")
	if err := inst.Validate([]interface{}{}, q); err == nil {
		t.Error("required empty should fail")
	}
	if err := inst.Validate([]interface{}{"basketball", "soccer"}, q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate([]interface{}{"tennis"}, q); err == nil {
		t.Error("expected invalid value")
	}
}

func TestCheckbox_Format(t *testing.T) {
	props := json.RawMessage(`{"options":[{"label":"篮球","value":"b"},{"label":"足球","value":"s"}]}`)
	q := schema.Question{ID: "q1", Type: "checkbox", Props: props}
	inst := question.Get("checkbox")
	got := inst.FormatValue([]interface{}{"b", "s"}, q)
	if got != "篮球、足球" {
		t.Errorf("format wrong: %q", got)
	}
}

func TestDate_Validate(t *testing.T) {
	inst := question.Get("date")
	q := schema.Question{ID: "q1", Type: "date", Required: true}
	if err := inst.Validate("2024-12-25", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate("2024/12/25", q); err == nil {
		t.Error("expected format fail")
	}
	if err := inst.Validate("abc", q); err == nil {
		t.Error("expected fail")
	}
}

func TestFile_Validate(t *testing.T) {
	inst := question.Get("file")
	q := schema.Question{ID: "q1", Type: "file", Required: true}
	if err := inst.Validate(nil, q); err == nil {
		t.Error("required nil should fail")
	}
	if err := inst.Validate("http://a.png", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate([]interface{}{"http://a.png", "http://b.png"}, q); err != nil {
		t.Errorf("multi should pass, got %v", err)
	}
}

func TestExtractValue(t *testing.T) {
	inst := question.Get("input")
	data := map[string]interface{}{"q1": "hello", "q2": "world"}
	q := schema.Question{ID: "q1"}
	got := inst.ExtractValue(data, q)
	if got != "hello" {
		t.Errorf("expected hello, got %v", got)
	}
	q2 := schema.Question{ID: "q3"}
	if got := inst.ExtractValue(data, q2); got != nil {
		t.Errorf("missing key should be nil, got %v", got)
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	if q := question.Get("nonexistent"); q != nil {
		t.Error("expected nil for unknown type")
	}
}

func TestSwitch_Validate(t *testing.T) {
	inst := question.Get("switch")
	q := schema.Question{ID: "q1", Type: "switch", Required: true}
	if err := inst.Validate("1", q); err != nil {
		t.Errorf("1 should pass, got %v", err)
	}
	if err := inst.Validate("0", q); err != nil {
		t.Errorf("0 should pass, got %v", err)
	}
	if err := inst.Validate("2", q); err == nil {
		t.Error("2 should fail")
	}
	if err := inst.Validate(true, q); err != nil {
		t.Errorf("bool should pass, got %v", err)
	}
}

func TestSwitch_Format(t *testing.T) {
	inst := question.Get("switch")
	if got := inst.FormatValue("1", schema.Question{}); got != "是" {
		t.Errorf("expected 是, got %q", got)
	}
	if got := inst.FormatValue("0", schema.Question{}); got != "否" {
		t.Errorf("expected 否, got %q", got)
	}
}

func TestRating_Validate(t *testing.T) {
	inst := question.Get("rating")
	props := json.RawMessage(`{"max":5}`)
	q := schema.Question{ID: "q1", Type: "rating", Required: true, Props: props}
	if err := inst.Validate(3, q); err != nil {
		t.Errorf("3 should pass, got %v", err)
	}
	if err := inst.Validate(0, q); err == nil {
		t.Error("0 should fail (must be >=1)")
	}
	if err := inst.Validate(6, q); err == nil {
		t.Error("6 should fail (max=5)")
	}
	if err := inst.Validate("4", q); err != nil {
		t.Errorf("string number should pass, got %v", err)
	}
}

func TestPicker_Validate(t *testing.T) {
	props := json.RawMessage(`{"options":[{"label":"北京","value":"bj"},{"label":"上海","value":"sh"}]}`)
	q := schema.Question{ID: "q1", Type: "picker", Props: props}
	inst := question.Get("picker")
	if err := inst.Validate("bj", q); err != nil {
		t.Errorf("expected pass, got %v", err)
	}
	if err := inst.Validate("gz", q); err == nil {
		t.Error("gz should fail")
	}
}

func TestSignature_Validate(t *testing.T) {
	inst := question.Get("signature")
	q := schema.Question{ID: "q1", Type: "signature", Required: true}
	if err := inst.Validate("data:image/png;base64,xxx", q); err != nil {
		t.Errorf("base64 should pass, got %v", err)
	}
	if err := inst.Validate("", q); err == nil {
		t.Error("empty should fail when required")
	}
}

func TestDisplay_Divider(t *testing.T) {
	inst := question.Get("divider")
	if inst == nil {
		t.Fatal("divider not registered")
	}
	if err := inst.Validate("anything", schema.Question{}); err != nil {
		t.Errorf("divider should never fail, got %v", err)
	}
	if got := inst.FormatValue("anything", schema.Question{}); got != "" {
		t.Errorf("divider should have empty format, got %q", got)
	}
}

func TestDisplay_Description(t *testing.T) {
	inst := question.Get("description")
	if inst == nil {
		t.Fatal("description not registered")
	}
	q := schema.Question{ID: "q1", Type: "description"}
	if got := inst.ExtractValue(map[string]interface{}{"q1": "x"}, q); got != nil {
		t.Errorf("description should return nil, got %v", got)
	}
}

func TestRegisteredCount_Updated(t *testing.T) {
	types := question.Types()
	expected := []string{"input", "text", "textarea", "number", "select", "radio", "checkbox", "date", "file", "switch", "rating", "picker", "signature", "divider", "description", "phone", "email", "idCard", "password", "time", "dateRange", "location", "autopop", "matrixRadio"}
	for _, e := range expected {
		if !question.Has(e) {
			t.Errorf("missing type %q (have: %v)", e, types)
		}
	}
	if len(types) < len(expected) {
		t.Errorf("expected at least %d types, got %d (%v)", len(expected), len(types), types)
	}
}

func TestPhone_Validate(t *testing.T) {
	inst := question.Get("phone")
	q := schema.Question{ID: "q1", Type: "phone", Required: true}
	if err := inst.Validate("13800138000", q); err != nil {
		t.Errorf("valid phone should pass, got %v", err)
	}
	if err := inst.Validate("12345", q); err == nil {
		t.Error("invalid phone should fail")
	}
	if err := inst.Validate("23800138000", q); err == nil {
		t.Error("prefix 2 should fail")
	}
}

func TestEmail_Validate(t *testing.T) {
	inst := question.Get("email")
	q := schema.Question{ID: "q1", Type: "email", Required: true}
	if err := inst.Validate("a@b.com", q); err != nil {
		t.Errorf("valid email should pass, got %v", err)
	}
	if err := inst.Validate("not-email", q); err == nil {
		t.Error("invalid email should fail")
	}
}

func TestIDCard_Validate(t *testing.T) {
	inst := question.Get("idCard")
	q := schema.Question{ID: "q1", Type: "idCard", Required: true}
	if err := inst.Validate("11010519491231002X", q); err != nil {
		t.Errorf("valid id should pass, got %v", err)
	}
	if err := inst.Validate("110105194912310026", q); err != nil {
		t.Errorf("valid id with checksum digit should pass, got %v", err)
	}
	if err := inst.Validate("12345", q); err == nil {
		t.Error("short id should fail")
	}
}

func TestPassword_Validate(t *testing.T) {
	inst := question.Get("password")
	q := schema.Question{ID: "q1", Type: "password", Required: true}
	if err := inst.Validate("123456", q); err != nil {
		t.Errorf("6+ chars should pass, got %v", err)
	}
	if err := inst.Validate("12345", q); err == nil {
		t.Error("5 chars should fail")
	}
}

func TestPassword_Format(t *testing.T) {
	inst := question.Get("password")
	if got := inst.FormatValue("secret", schema.Question{}); got != "******" {
		t.Errorf("password format should mask, got %q", got)
	}
}

func TestTime_Validate(t *testing.T) {
	inst := question.Get("time")
	q := schema.Question{ID: "q1", Type: "time", Required: true}
	if err := inst.Validate("09:30", q); err != nil {
		t.Errorf("HH:MM should pass, got %v", err)
	}
	if err := inst.Validate("09:30:45", q); err != nil {
		t.Errorf("HH:MM:SS should pass, got %v", err)
	}
	if err := inst.Validate("9:30", q); err != nil {
		t.Errorf("single digit hour should also pass, got %v", err)
	}
	if err := inst.Validate("abc:30", q); err == nil {
		t.Error("non-numeric should fail")
	}
	if err := inst.Validate("09:30:45:00", q); err == nil {
		t.Error("too many parts should fail")
	}
}

func TestDateRange_Validate(t *testing.T) {
	inst := question.Get("dateRange")
	q := schema.Question{ID: "q1", Type: "dateRange", Required: true}
	if err := inst.Validate([]interface{}{"2024-01-01", "2024-01-31"}, q); err != nil {
		t.Errorf("valid range should pass, got %v", err)
	}
	if err := inst.Validate([]interface{}{"2024-01-01"}, q); err == nil {
		t.Error("single date should fail")
	}
	if err := inst.Validate("2024-01-01", q); err == nil {
		t.Error("string should fail (need array)")
	}
}

func TestDateRange_Format(t *testing.T) {
	inst := question.Get("dateRange")
	got := inst.FormatValue([]interface{}{"2024-01-01", "2024-01-31"}, schema.Question{})
	if got != "2024-01-01 ~ 2024-01-31" {
		t.Errorf("format wrong: %q", got)
	}
}

func TestLocation_Validate(t *testing.T) {
	inst := question.Get("location")
	q := schema.Question{ID: "q1", Type: "location", Required: true}
	if err := inst.Validate(map[string]interface{}{"address": "中关村", "lat": "39.9", "lng": "116.3"}, q); err != nil {
		t.Errorf("valid location should pass, got %v", err)
	}
	if err := inst.Validate(map[string]interface{}{"address": ""}, q); err == nil {
		t.Error("empty address should fail")
	}
	if err := inst.Validate("plain string", q); err == nil {
		t.Error("non-map should fail")
	}
}

func TestLocation_Format(t *testing.T) {
	inst := question.Get("location")
	v := map[string]interface{}{"address": "中关村", "lat": "39.9", "lng": "116.3"}
	got := inst.FormatValue(v, schema.Question{})
	if got != "中关村 (39.9,116.3)" {
		t.Errorf("format wrong: %q", got)
	}
}

func TestAutoPop_Validate(t *testing.T) {
	inst := question.Get("autopop")
	q := schema.Question{ID: "q1", Type: "autopop", Required: true}
	if err := inst.Validate("张三", q); err != nil {
		t.Errorf("string should pass, got %v", err)
	}
	if err := inst.Validate(123, q); err == nil {
		t.Error("non-string should fail")
	}
}

func TestMatrixRadio_Validate(t *testing.T) {
	props := json.RawMessage(`{"rows":[{"label":"行1","value":"r1"},{"label":"行2","value":"r2"}],"columns":[{"label":"列A","value":"a"},{"label":"列B","value":"b"}]}`)
	q := schema.Question{ID: "q1", Type: "matrixRadio", Required: true, Props: props}
	inst := question.Get("matrixRadio")
	if err := inst.Validate(map[string]interface{}{"r1": "a", "r2": "b"}, q); err != nil {
		t.Errorf("all rows should pass, got %v", err)
	}
	if err := inst.Validate(map[string]interface{}{"r1": "a"}, q); err == nil {
		t.Error("missing row should fail (required)")
	}
	// 非必填时部分填允许
	q.Required = false
	if err := inst.Validate(map[string]interface{}{"r1": "a"}, q); err != nil {
		t.Errorf("partial ok when not required, got %v", err)
	}
}

func TestMatrixRadio_Format(t *testing.T) {
	props := json.RawMessage(`{"rows":[{"label":"行1","value":"r1"}],"columns":[{"label":"列A","value":"a"}]}`)
	q := schema.Question{ID: "q1", Type: "matrixRadio", Props: props}
	inst := question.Get("matrixRadio")
	v := map[string]interface{}{"r1": "a"}
	got := inst.FormatValue(v, q)
	if got != "行1: 列A" {
		t.Errorf("format wrong: %q", got)
	}
}
