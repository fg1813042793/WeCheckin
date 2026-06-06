package schema

import (
	"testing"
)

func TestExtractFieldValues_Empty(t *testing.T) {
	out := ExtractFieldValues("", "")
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}
}

func TestExtractFieldValues_AnswerOnly_NoSchema(t *testing.T) {
	out := ExtractFieldValues(`{"q1":"a","q2":"b"}`, "")
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
}

func TestExtractFieldValues_OldArrayAnswer_OldArraySchema(t *testing.T) {
	answer := `["张三","13800138000"]`
	schema := `[{"label":"姓名","type":"input"},{"label":"电话","type":"input"}]`
	out := ExtractFieldValues(answer, schema)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[0].Label != "姓名" || out[0].Type != "input" || out[0].Value != "张三" {
		t.Errorf("[0] wrong: %+v", out[0])
	}
	if out[1].Label != "电话" || out[1].Value != "13800138000" {
		t.Errorf("[1] wrong: %+v", out[1])
	}
}

func TestExtractFieldValues_NewObjectAnswer_NewSchema(t *testing.T) {
	answer := `{"q1":"张三","q2":"13800138000","q3":"测试"}`
	schema := `{"version":"2.0","questions":[
		{"id":"q1","type":"input","title":"姓名"},
		{"id":"q2","type":"input","title":"电话"},
		{"id":"q3","type":"textarea","title":"备注"}
	]}`
	out := ExtractFieldValues(answer, schema)
	if len(out) != 3 {
		t.Fatalf("expected 3, got %d", len(out))
	}
	if out[0].Label != "姓名" || out[0].Value != "张三" {
		t.Errorf("[0] wrong: %+v", out[0])
	}
	if out[2].Label != "备注" || out[2].Type != "textarea" || out[2].Value != "测试" {
		t.Errorf("[2] wrong: %+v", out[2])
	}
}

func TestExtractFieldValues_OldArrayAnswer_NewSchema(t *testing.T) {
	// 客户端写入老格式数组，schema 已是新格式
	answer := `["张三","13800138000"]`
	schema := `{"version":"2.0","questions":[
		{"id":"q1","type":"input","title":"姓名"},
		{"id":"q2","type":"input","title":"电话"}
	]}`
	out := ExtractFieldValues(answer, schema)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[0].Value != "张三" || out[1].Value != "13800138000" {
		t.Errorf("values wrong: %+v", out)
	}
}

func TestExtractFieldValues_NewObjectAnswer_OldSchema(t *testing.T) {
	// 新格式 answer 对象 + 老格式 schema
	answer := `{"q1":"张三","q2":"13800138000"}`
	schema := `[{"label":"姓名","type":"input"},{"label":"电话","type":"input"}]`
	out := ExtractFieldValues(answer, schema)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[0].Label != "姓名" || out[0].Type != "input" || out[0].Value != "张三" {
		t.Errorf("[0] wrong: %+v", out[0])
	}
}

func TestExtractFieldValues_MissingValue(t *testing.T) {
	answer := `{"q1":"张三"}`
	schema := `{"version":"2.0","questions":[
		{"id":"q1","type":"input","title":"姓名"},
		{"id":"q2","type":"input","title":"电话"}
	]}`
	out := ExtractFieldValues(answer, schema)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[1].Value != nil {
		t.Errorf("missing should be nil: %+v", out[1])
	}
}

func TestExtractFieldValues_InvalidJSON(t *testing.T) {
	out := ExtractFieldValues("not json", `[{"label":"x","type":"input"}]`)
	if len(out) != 0 {
		t.Fatalf("expected empty, got %+v", out)
	}
}

func TestExtractImagesLocation_OldFormat(t *testing.T) {
	answer := `["http://a.png","中关村大街1号","40.0","116.0","备注"]`
	schema := `[{"label":"打卡照片","type":"input"},{"label":"打卡位置","type":"input"}]`
	imgs, loc := ExtractImagesLocation(answer, schema)
	if len(imgs) != 1 || imgs[0] != "http://a.png" {
		t.Errorf("images wrong: %+v", imgs)
	}
	if loc != "中关村大街1号" {
		t.Errorf("location wrong: %q", loc)
	}
}

func TestExtractImagesLocation_NewFormat(t *testing.T) {
	answer := `{"q1":"http://a.png","q2":"中关村大街1号","q3":"40.0","q4":"116.0"}`
	schema := `{"version":"2.0","questions":[
		{"id":"q1","type":"input","title":"打卡照片"},
		{"id":"q2","type":"input","title":"打卡位置"},
		{"id":"q3","type":"input","title":"打卡位置-纬度"},
		{"id":"q4","type":"input","title":"打卡位置-经度"}
	]}`
	imgs, loc := ExtractImagesLocation(answer, schema)
	if len(imgs) != 1 || imgs[0] != "http://a.png" {
		t.Errorf("images wrong: %+v", imgs)
	}
	if loc != "中关村大街1号" {
		t.Errorf("location wrong: %q", loc)
	}
}

func TestExtractImagesLocation_NoImageNoLocation(t *testing.T) {
	imgs, loc := ExtractImagesLocation(`{"q1":"hello"}`, `{"version":"2.0","questions":[{"id":"q1","type":"input","title":"姓名"}]}`)
	if imgs == nil || len(imgs) != 0 {
		t.Errorf("images should be empty slice, got %+v", imgs)
	}
	if loc != "" {
		t.Errorf("location should be empty, got %q", loc)
	}
}
