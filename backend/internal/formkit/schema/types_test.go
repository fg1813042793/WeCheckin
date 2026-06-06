package schema

import (
	"strings"
	"testing"
)

func TestParse_Empty(t *testing.T) {
	s, err := Parse("")
	if err != nil {
		t.Fatalf("empty should not error: %v", err)
	}
	if s.Version != CurrentVersion {
		t.Fatalf("expected version %s, got %s", CurrentVersion, s.Version)
	}
	if len(s.Questions) != 0 {
		t.Fatalf("expected 0 questions, got %d", len(s.Questions))
	}
}

func TestParse_Whitespace(t *testing.T) {
	s, err := Parse("   \n\t  ")
	if err != nil {
		t.Fatalf("whitespace should not error: %v", err)
	}
	if s.Version != CurrentVersion {
		t.Fatalf("expected version, got %s", s.Version)
	}
}

func TestParse_Legal(t *testing.T) {
	raw := `{"version":"2.0","questions":[{"id":"q1","type":"input","title":"姓名"}]}`
	s, err := Parse(raw)
	if err != nil {
		t.Fatalf("legal schema should not error: %v", err)
	}
	if len(s.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(s.Questions))
	}
	if s.Questions[0].ID != "q1" {
		t.Fatalf("expected q1, got %s", s.Questions[0].ID)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid schema json") {
		t.Fatalf("expected json error, got: %v", err)
	}
}

func TestParse_MissingVersion(t *testing.T) {
	raw := `{"questions":[{"id":"q1","type":"input","title":"x"}]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "missing version") {
		t.Fatalf("expected missing version error, got: %v", err)
	}
}

func TestParse_UnsupportedVersion(t *testing.T) {
	raw := `{"version":"3.0","questions":[]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported schema version") {
		t.Fatalf("expected version error, got: %v", err)
	}
}

func TestParse_DuplicateID(t *testing.T) {
	raw := `{"version":"2.0","questions":[{"id":"q1","type":"input","title":"a"},{"id":"q1","type":"input","title":"b"}]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for duplicate id")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got: %v", err)
	}
}

func TestParse_MissingID(t *testing.T) {
	raw := `{"version":"2.0","questions":[{"type":"input","title":"x"}]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestParse_MissingType(t *testing.T) {
	raw := `{"version":"2.0","questions":[{"id":"q1","title":"x"}]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestParse_MissingTitle(t *testing.T) {
	raw := `{"version":"2.0","questions":[{"id":"q1","type":"input"}]}`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestMarshal(t *testing.T) {
	s := &FormSchema{Version: CurrentVersion, Questions: []Question{
		{ID: "q1", Type: "input", Title: "姓名"},
	}}
	out, err := Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(out, `"version":"2.0"`) {
		t.Fatalf("missing version: %s", out)
	}
	if !strings.Contains(out, `"q1"`) {
		t.Fatalf("missing q1: %s", out)
	}
}

func TestMarshal_AutoVersion(t *testing.T) {
	s := &FormSchema{Questions: []Question{{ID: "q1", Type: "input", Title: "x"}}}
	out, err := Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(out, `"version":"2.0"`) {
		t.Fatalf("auto version failed: %s", out)
	}
}

func TestMarshal_Nil(t *testing.T) {
	_, err := Marshal(nil)
	if err == nil {
		t.Fatal("expected error for nil schema")
	}
}

func TestEmpty(t *testing.T) {
	s := Empty()
	if s.Version != CurrentVersion {
		t.Fatalf("expected version, got %s", s.Version)
	}
	if s.Questions == nil {
		t.Fatal("expected empty slice, got nil")
	}
}

func TestMustParse(t *testing.T) {
	s := MustParse(`{"version":"2.0","questions":[{"id":"q1","type":"input","title":"x"}]}`)
	if s == nil || len(s.Questions) != 1 {
		t.Fatalf("unexpected: %+v", s)
	}
}

func TestMustParse_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	MustParse("invalid")
}
