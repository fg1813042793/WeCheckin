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

func TestExamSubmissionLimitsAcceptsNumericStrings(t *testing.T) {
	deviceLimit, ipLimit := examSubmissionLimits(`{"deviceLimit":"2","ipLimit":"3"}`)
	if deviceLimit != 2 || ipLimit != 3 {
		t.Fatalf("examSubmissionLimits numeric strings = (%d,%d), want (2,3)", deviceLimit, ipLimit)
	}
}

func TestExamSubmissionLimitsIgnoresInvalidValues(t *testing.T) {
	deviceLimit, ipLimit := examSubmissionLimits(`{"deviceLimit":-1,"ipLimit":"bad"}`)
	if deviceLimit != 0 || ipLimit != 0 {
		t.Fatalf("examSubmissionLimits invalid values = (%d,%d), want (0,0)", deviceLimit, ipLimit)
	}
}

func TestSafeExamQuestionsHidesAnswerAndAnalysisByDefault(t *testing.T) {
	got := safeExamQuestions([]model.ExamQuestion{{
		ID:       7,
		Type:     "radio",
		Title:    "题目",
		Options:  `[{"label":"A","value":"A"}]`,
		Answer:   "A",
		Analysis: "解析",
		Score:    5,
	}}, PaperQuestionOptions{})
	if len(got) != 1 {
		t.Fatalf("safeExamQuestions length = %d, want 1", len(got))
	}
	if got[0].Answer != nil {
		t.Fatalf("safeExamQuestions leaked answer: %#v", got[0])
	}
	if got[0].Analysis != nil {
		t.Fatalf("safeExamQuestions leaked analysis: %#v", got[0])
	}
	if got[0].Title != "题目" || got[0].Score != 5 {
		t.Fatalf("safeExamQuestions lost public fields: %#v", got[0])
	}
}

func TestSafeExamQuestionsCanIncludeAnswerAndAnalysis(t *testing.T) {
	got := safeExamQuestions([]model.ExamQuestion{{
		Answer:   "A",
		Analysis: "解析",
	}}, PaperQuestionOptions{IncludeAnswer: true, IncludeAnalysis: true})
	if got[0].Answer == nil || *got[0].Answer != "A" || got[0].Analysis == nil || *got[0].Analysis != "解析" {
		t.Fatalf("safeExamQuestions did not include requested fields: %#v", got[0])
	}
}
