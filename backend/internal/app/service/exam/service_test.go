package exam

import (
	"testing"

	"wecheckin-backend/backend/internal/model"
)

func TestNormalizeExamForCreateAppliesDefaults(t *testing.T) {
	exam := normalizeExamForCreate(model.Exam{Title: "测试"})
	if exam.Mode != "exam" {
		t.Fatalf("Mode = %q, want exam", exam.Mode)
	}
	if exam.Schema == "" {
		t.Fatalf("Schema should get default empty schema")
	}
	if exam.Settings != "{}" {
		t.Fatalf("Settings = %q, want {}", exam.Settings)
	}
	if exam.AddTime == 0 {
		t.Fatalf("AddTime should be assigned")
	}
}

func TestDecodeRecordDetailPayloadToleratesEmptyJSON(t *testing.T) {
	result := decodeRecordDetailPayload(model.ExamRecord{}, model.Exam{})
	if result.Answers != nil {
		t.Fatalf("empty answers should stay nil, got %#v", result.Answers)
	}
	if result.Scoring != nil {
		t.Fatalf("empty scoring should stay nil, got %#v", result.Scoring)
	}
	if result.Schema != nil {
		t.Fatalf("empty schema should stay nil, got %#v", result.Schema)
	}
}
