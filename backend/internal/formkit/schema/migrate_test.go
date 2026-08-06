package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIsOldFormat(t *testing.T) {
	cases := []struct {
		input  string
		expect bool
	}{
		{"", false},
		{"   ", false},
		{`[{"label":"x"}]`, true},
		{`[ ]`, true},
		{`{"version":"2.0"}`, false},
		{`{ }`, false},
		{`not json`, false},
	}
	for _, c := range cases {
		if got := IsOldFormat(c.input); got != c.expect {
			t.Errorf("IsOldFormat(%q) = %v, want %v", c.input, got, c.expect)
		}
	}
}

func TestMigrateFromOld_Empty(t *testing.T) {
	out, err := MigrateFromOld("")
	if err != nil {
		t.Fatalf("empty migrate: %v", err)
	}
	if !strings.Contains(out, `"version":"2.0"`) {
		t.Fatalf("missing version: %s", out)
	}
}

func TestMigrateFromOld_BasicTypes(t *testing.T) {
	old := `[
		{"label":"姓名","type":"input","required":true,"placeholder":"请输入姓名"},
		{"label":"简介","type":"textarea","required":false},
		{"label":"年龄","type":"number","required":true},
		{"label":"性别","type":"select","required":true,"options":["男","女","其他"]},
		{"label":"来源","type":"picker","required":false}
	]`
	out, err := MigrateFromOld(old)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	s, err := Parse(out)
	if err != nil {
		t.Fatalf("parse migrated: %v", err)
	}
	if len(s.Questions) != 5 {
		t.Fatalf("expected 5 questions, got %d", len(s.Questions))
	}
	expectedIDs := []string{"q1", "q2", "q3", "q4", "q5"}
	for i, q := range s.Questions {
		if q.ID != expectedIDs[i] {
			t.Errorf("q[%d].ID = %s, want %s", i, q.ID, expectedIDs[i])
		}
	}
	if s.Questions[0].Type != "input" || !s.Questions[0].Required {
		t.Errorf("q1 wrong: %+v", s.Questions[0])
	}
	if s.Questions[0].Placeholder != "请输入姓名" {
		t.Errorf("q1 placeholder: %s", s.Questions[0].Placeholder)
	}
	if s.Questions[3].Type != "select" {
		t.Errorf("q4 type: %s", s.Questions[3].Type)
	}
	if s.Questions[4].Type != "select" {
		t.Errorf("q5 picker should map to select, got: %s", s.Questions[4].Type)
	}
	var props map[string]interface{}
	if err := json.Unmarshal(s.Questions[3].Props, &props); err != nil {
		t.Fatalf("q4 props unmarshal: %v", err)
	}
	opts, ok := props["options"].([]interface{})
	if !ok || len(opts) != 3 {
		t.Fatalf("q4 options wrong: %+v", props)
	}
	first := opts[0].(map[string]interface{})
	if first["label"] != "男" || first["value"] != "男" {
		t.Errorf("q4 option[0] wrong: %+v", first)
	}
}

func TestMigrateFromOld_UnknownTypeFallback(t *testing.T) {
	old := `[{"label":"x","type":"weird-type"}]`
	out, err := MigrateFromOld(old)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	s, _ := Parse(out)
	if s.Questions[0].Type != "input" {
		t.Fatalf("unknown type should fallback to input, got: %s", s.Questions[0].Type)
	}
}

func TestMigrateFromOld_InvalidJSON(t *testing.T) {
	_, err := MigrateFromOld("not json")
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestLoadSchema_Empty(t *testing.T) {
	s, err := LoadSchema("")
	if err != nil {
		t.Fatalf("load empty: %v", err)
	}
	if s.Version != CurrentVersion {
		t.Fatalf("version: %s", s.Version)
	}
}

func TestLoadSchema_OldFormat(t *testing.T) {
	old := `[{"label":"x","type":"input"}]`
	s, err := LoadSchema(old)
	if err != nil {
		t.Fatalf("load old: %v", err)
	}
	if len(s.Questions) != 1 || s.Questions[0].ID != "q1" {
		t.Fatalf("migrate failed: %+v", s.Questions)
	}
}

func TestLoadSchema_NewFormat(t *testing.T) {
	new := `{"version":"2.0","questions":[{"id":"q1","type":"input","title":"x"}]}`
	s, err := LoadSchema(new)
	if err != nil {
		t.Fatalf("load new: %v", err)
	}
	if s.Questions[0].ID != "q1" {
		t.Fatalf("id: %s", s.Questions[0].ID)
	}
}

func TestLoadSchemaString(t *testing.T) {
	out, err := LoadSchemaString(`[{"label":"x","type":"input"}]`)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !strings.Contains(out, `"version":"2.0"`) {
		t.Fatalf("not migrated: %s", out)
	}
}

func TestMigrateAnswer_Empty(t *testing.T) {
	out, err := MigrateAnswer("", nil)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if out != "{}" {
		t.Fatalf("expected {}, got %s", out)
	}
}

func TestMigrateAnswer_AlreadyObject(t *testing.T) {
	in := `{"q1":"张三","q2":"13800138000"}`
	out, err := MigrateAnswer(in, nil)
	if err != nil {
		t.Fatalf("object: %v", err)
	}
	if out != in {
		t.Fatalf("object passthrough: %s", out)
	}
}

func TestMigrateAnswer_ArrayWithSchema(t *testing.T) {
	schema := &FormSchema{
		Version: CurrentVersion,
		Questions: []Question{
			{ID: "q1", Type: "input", Title: "姓名"},
			{ID: "q2", Type: "input", Title: "电话"},
			{ID: "q3", Type: "input", Title: "备注"},
		},
	}
	out, err := MigrateAnswer(`["张三","13800138000","测试"]`, schema)
	if err != nil {
		t.Fatalf("array: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(out), &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if obj["q1"] != "张三" || obj["q2"] != "13800138000" || obj["q3"] != "测试" {
		t.Fatalf("values wrong: %+v", obj)
	}
}

func TestMigrateAnswer_ArrayExceedsSchema(t *testing.T) {
	schema := &FormSchema{
		Version: CurrentVersion,
		Questions: []Question{
			{ID: "q1", Type: "input", Title: "a"},
		},
	}
	out, err := MigrateAnswer(`["v1","v2","v3"]`, schema)
	if err != nil {
		t.Fatalf("array: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(out), &obj)
	if obj["q1"] != "v1" {
		t.Fatalf("q1 wrong: %v", obj["q1"])
	}
	if obj["q2"] != "v2" || obj["q3"] != "v3" {
		t.Fatalf("extras wrong: %+v", obj)
	}
}

func TestMigrateAnswer_ArrayNoSchema(t *testing.T) {
	out, err := MigrateAnswer(`["a","b","c"]`, nil)
	if err != nil {
		t.Fatalf("no schema: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(out), &obj)
	if obj["q1"] != "a" || obj["q2"] != "b" || obj["q3"] != "c" {
		t.Fatalf("default ids wrong: %+v", obj)
	}
}

func TestMigrateAnswer_InvalidJSON(t *testing.T) {
	out, err := MigrateAnswer("not json", nil)
	if err != nil {
		t.Fatalf("invalid: %v", err)
	}
	if out != "{}" {
		t.Fatalf("expected {} fallback, got %s", out)
	}
}

func TestLoadAnswer(t *testing.T) {
	out, err := LoadAnswer(`["x","y"]`, nil)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !strings.Contains(out, "q1") {
		t.Fatalf("no q1: %s", out)
	}
}

func TestAnswerToLegacyArray_FromObject(t *testing.T) {
	schema := &FormSchema{
		Version: CurrentVersion,
		Questions: []Question{
			{ID: "q1", Type: "input", Title: "a"},
			{ID: "q2", Type: "input", Title: "b"},
		},
	}
	arr, err := AnswerToLegacyArray(`{"q1":"x","q2":"y"}`, schema)
	if err != nil {
		t.Fatalf("to legacy: %v", err)
	}
	if len(arr) != 2 || arr[0] != "x" || arr[1] != "y" {
		t.Fatalf("wrong: %+v", arr)
	}
}

func TestAnswerToLegacyArray_MissingKey(t *testing.T) {
	schema := &FormSchema{
		Version: CurrentVersion,
		Questions: []Question{
			{ID: "q1", Type: "input", Title: "a"},
			{ID: "q2", Type: "input", Title: "b"},
		},
	}
	arr, err := AnswerToLegacyArray(`{"q1":"x"}`, schema)
	if err != nil {
		t.Fatalf("missing: %v", err)
	}
	if len(arr) != 2 || arr[0] != "x" {
		t.Fatalf("wrong: %+v", arr)
	}
}

func TestAnswerToLegacyArray_FromArray(t *testing.T) {
	arr, err := AnswerToLegacyArray(`["a","b"]`, nil)
	if err != nil {
		t.Fatalf("from arr: %v", err)
	}
	if len(arr) != 2 || arr[0] != "a" {
		t.Fatalf("wrong: %+v", arr)
	}
}

func TestAnswerToLegacyArray_Empty(t *testing.T) {
	arr, err := AnswerToLegacyArray("", nil)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if len(arr) != 0 {
		t.Fatalf("expected 0, got %d", len(arr))
	}
}
