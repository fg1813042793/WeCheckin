package report

import (
	"strings"
	"testing"
)

func TestRenderAnswers_OldFormat(t *testing.T) {
	schema := `[{"label":"姓名","type":"input"},{"label":"是否学生","type":"radio","options":[{"label":"是","value":"y"},{"label":"否","value":"n"}]}]`
	items := []AnswerItem{
		{UserID: "u1", AddTime: "2024-01-01 10:00", Forms: `["张三","y"]`},
		{UserID: "u2", AddTime: "2024-01-01 10:05", Forms: `["李四","n"]`},
	}
	tbl, err := RenderAnswers(schema, items)
	if err != nil {
		t.Fatal(err)
	}
	if len(tbl.Headers) != 4 { // 用户ID, 提交时间, 姓名, 是否学生
		t.Errorf("expected 4 headers, got %d: %v", len(tbl.Headers), tbl.Headers)
	}
	if len(tbl.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.Rows))
	}
	if tbl.Rows[0].Values[0] != "张三" {
		t.Errorf("row 0 name: got %q", tbl.Rows[0].Values[0])
	}
	if tbl.Rows[0].Values[1] != "是" {
		t.Errorf("row 0 radio: got %q (expected label 是)", tbl.Rows[0].Values[1])
	}
}

func TestRenderAnswers_NewFormat(t *testing.T) {
	schema := `{"version":"2.0","questions":[
		{"id":"q1","type":"input","title":"姓名"},
		{"id":"q2","type":"checkbox","title":"爱好","props":{"options":[{"label":"篮球","value":"b"},{"label":"足球","value":"f"}]}}
	]}`
	items := []AnswerItem{
		{UserID: "u1", AddTime: "2024-01-01", Forms: `{"q1":"张三","q2":["b","f"]}`},
	}
	tbl, err := RenderAnswers(schema, items)
	if err != nil {
		t.Fatal(err)
	}
	if len(tbl.Headers) != 4 {
		t.Errorf("expected 4 headers, got %d", len(tbl.Headers))
	}
	if tbl.Rows[0].Values[0] != "张三" {
		t.Errorf("name wrong: %q", tbl.Rows[0].Values[0])
	}
	if tbl.Rows[0].Values[1] != "篮球、足球" {
		t.Errorf("checkbox format wrong: %q", tbl.Rows[0].Values[1])
	}
}

func TestToCSV(t *testing.T) {
	tbl := Table{
		Headers: []string{"A", "B"},
		Rows: []Row{
			{UserID: "u1", AddTime: "t1", Values: []string{"x", "y"}},
		},
	}
	csv := string(ToCSV(tbl))
	if !strings.Contains(csv, "A,B") {
		t.Error("header missing")
	}
	if !strings.Contains(csv, "u1,t1,x,y") {
		t.Errorf("row missing: %q", csv)
	}
}

func TestFieldStats_Number(t *testing.T) {
	schema := `[{"label":"年龄","type":"number"}]`
	items := []AnswerItem{
		{UserID: "u1", Forms: `[20]`},
		{UserID: "u2", Forms: `[30]`},
		{UserID: "u3", Forms: `[40]`},
		{UserID: "u4", Forms: `[null]`},
	}
	stats := FieldStats(schema, items)
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	fs := stats[0]
	if fs.NonEmpty != 3 {
		t.Errorf("nonEmpty: got %d", fs.NonEmpty)
	}
	if fs.Empty != 1 {
		t.Errorf("empty: got %d", fs.Empty)
	}
	if fs.NumericStat == nil {
		t.Fatal("NumericStat nil")
	}
	if fs.NumericStat.Sum != 90 {
		t.Errorf("sum: got %v", fs.NumericStat.Sum)
	}
	if fs.NumericStat.Avg != 30 {
		t.Errorf("avg: got %v", fs.NumericStat.Avg)
	}
	if fs.NumericStat.Min != 20 {
		t.Errorf("min: got %v", fs.NumericStat.Min)
	}
	if fs.NumericStat.Max != 40 {
		t.Errorf("max: got %v", fs.NumericStat.Max)
	}
}

func TestFieldStats_Distribution(t *testing.T) {
	schema := `{"version":"2.0","questions":[{"id":"q1","type":"select","title":"类型","props":{"options":[{"label":"A","value":"a"},{"label":"B","value":"b"}]}}]}`
	items := []AnswerItem{
		{UserID: "u1", Forms: `{"q1":"a"}`},
		{UserID: "u2", Forms: `{"q1":"a"}`},
		{UserID: "u3", Forms: `{"q1":"b"}`},
		{UserID: "u4", Forms: `{}`},
	}
	stats := FieldStats(schema, items)
	if len(stats) != 1 {
		t.Fatalf("expected 1, got %d", len(stats))
	}
	fs := stats[0]
	if fs.NonEmpty != 3 {
		t.Errorf("nonEmpty: got %d", fs.NonEmpty)
	}
	if fs.Dist["A"] != 2 {
		t.Errorf("A dist: got %d", fs.Dist["A"])
	}
	if fs.Dist["B"] != 1 {
		t.Errorf("B dist: got %d", fs.Dist["B"])
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hello world", "hello_world"},
		{"a/b\\c", "a_b_c"},
		{"", "export"},
	}
	for _, tt := range tests {
		got := SanitizeFilename(tt.in)
		if got != tt.want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
