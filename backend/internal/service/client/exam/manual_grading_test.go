package exam

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

func gradingFixture(t *testing.T, object bool) (model.ExamRecord, *RecordGradingDetail) {
	t.Helper()
	result := examPkg.PaperResult{Results: []examPkg.Result{
		{QuestionID: 1, FullScore: 4, GotScore: 4, Correct: true},
		{QuestionID: 3, FullScore: 6, NeedManual: true, Reason: "主观题，需人工判分"},
	}, TotalScore: 4, FullScore: 10, ManualCount: 1, CorrectCnt: 1}
	var raw []byte
	if object {
		raw, _ = json.Marshal(result)
	} else {
		raw, _ = json.Marshal(result.Results)
	}
	record := model.ExamRecord{ID: 19, ExamID: 5, Status: 1, Score: 4, TotalScore: 10, Result: string(raw), Answers: `{"a":"A","essay":"我的答案"}`}
	snapshot := recordReviewSnapshot{Version: 1, PassScore: 7, Questions: []reviewQuestionSnapshot{
		{QuestionID: 1, AnswerKey: "a", Title: "<b>选择题</b>", Type: "radio", ReferenceAnswer: "A"},
		{QuestionID: 3, AnswerKey: "essay", Title: "简答题", Type: "textarea", ReferenceAnswer: "参考要点"},
	}}
	data, _ := json.Marshal(snapshot)
	record.ReviewSnapshot = string(data)
	detail, err := buildRecordGradingDetail(record, model.Exam{ID: 5, Title: "测试考试", PassScore: 99}, snapshot, "")
	if err != nil {
		t.Fatal(err)
	}
	return record, detail
}

func TestManualGradingPreservesAutomaticScoreAndCompletesRecord(t *testing.T) {
	for _, object := range []bool{false, true} {
		record, detail := gradingFixture(t, object)
		score := 3
		updated, err := applyManualGrading(record, detail, GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 3, Score: &score}}, Comment: " 有基本要点 "}, 8, 123456)
		if err != nil {
			t.Fatal(err)
		}
		if updated.Score != 7 || updated.TotalScore != 10 || updated.Pass != 1 || updated.Status != 2 || updated.GradedBy != 8 || updated.GradedAt != 123456 || updated.GradingComment != "有基本要点" {
			t.Fatalf("incorrect graded record: %+v", updated)
		}
		results := DecodeRecordResults(updated.Result)
		if results[0].GotScore != 4 || results[1].GotScore != 3 || results[1].NeedManual || results[1].Correct {
			t.Fatalf("incorrect result: %+v", results)
		}
		if updated.Answers != record.Answers || updated.ReviewSnapshot != record.ReviewSnapshot || record.Status != 1 {
			t.Fatal("grading must preserve original answers, snapshot and caller's source record")
		}
		if (updated.Result[0] == '{') != object {
			t.Fatal("grading must preserve stored array/object compatibility")
		}
	}
}

func TestManualGradingRejectsInvalidOrStaleSubmissions(t *testing.T) {
	score, low, high := 3, -1, 7
	for _, tt := range []struct {
		name    string
		scores  []ManualQuestionScore
		version string
		status  int
		want    error
	}{
		{"missing score", []ManualQuestionScore{{QuestionID: 3}}, "current", 1, ErrRecordGradingInvalid},
		{"missing question", nil, "current", 1, ErrRecordGradingInvalid},
		{"below zero", []ManualQuestionScore{{QuestionID: 3, Score: &low}}, "current", 1, ErrRecordGradingInvalid},
		{"above full score", []ManualQuestionScore{{QuestionID: 3, Score: &high}}, "current", 1, ErrRecordGradingInvalid},
		{"duplicate", []ManualQuestionScore{{QuestionID: 3, Score: &score}, {QuestionID: 3, Score: &score}}, "current", 1, ErrRecordGradingInvalid},
		{"automatic question", []ManualQuestionScore{{QuestionID: 1, Score: &score}, {QuestionID: 3, Score: &score}}, "current", 1, ErrRecordGradingInvalid},
		{"unknown question", []ManualQuestionScore{{QuestionID: 99, Score: &score}}, "current", 1, ErrRecordGradingInvalid},
		{"stale", []ManualQuestionScore{{QuestionID: 3, Score: &score}}, "old", 1, ErrRecordGradingChanged},
		{"ongoing", []ManualQuestionScore{{QuestionID: 3, Score: &score}}, "current", 0, ErrRecordNotGradable},
		{"duplicate submit", []ManualQuestionScore{{QuestionID: 3, Score: &score}}, "current", 2, ErrRecordNotGradable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			record, detail := gradingFixture(t, false)
			record.Status = tt.status
			version := tt.version
			if version == "current" {
				version = detail.Version
			}
			_, err := applyManualGrading(record, detail, GradeRecordRequest{ExpectedVersion: version, Scores: tt.scores}, 8, 123456)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestManualGradingAcceptsExplicitZeroAndRejectsLongComment(t *testing.T) {
	record, detail := gradingFixture(t, false)
	zero := 0
	req := GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 3, Score: &zero}}}
	updated, err := applyManualGrading(record, detail, req, 8, 123456)
	if err != nil || updated.Score != 4 || updated.Pass != 0 || updated.Status != 2 {
		t.Fatalf("zero score: %+v %v", updated, err)
	}
	req.Comment = string(make([]rune, 1001))
	if _, err := applyManualGrading(record, detail, req, 8, 123456); !errors.Is(err, ErrRecordGradingInvalid) {
		t.Fatalf("long comment error=%v", err)
	}
}

func TestGradingSnapshotFreezesQuestionAndPassScore(t *testing.T) {
	record, _ := gradingFixture(t, false)
	snapshot, err := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := buildRecordGradingDetail(record, model.Exam{ID: 5, Title: "修改后的考试", PassScore: 99, Schema: `{"questions":[]}`}, snapshot, "")
	if err != nil {
		t.Fatal(err)
	}
	if detail.PassScore != 7 || detail.Questions[1].Title != "简答题" || detail.Questions[1].Answer != "我的答案" || detail.Questions[1].ReferenceAnswer != "参考要点" {
		t.Fatalf("snapshot not used: %+v", detail)
	}
	raw, _ := json.Marshal(record)
	if !json.Valid(raw) {
		t.Fatal("record JSON should be valid")
	}
	if strings.Contains(string(raw), "参考要点") {
		t.Fatal("reference answer leaked through public record JSON")
	}
	var public map[string]interface{}
	_ = json.Unmarshal(raw, &public)
	if _, leaked := public["reviewSnapshot"]; leaked {
		t.Fatal("private reference answer snapshot leaked through record JSON")
	}
}

func TestGradingSnapshotPreservesMatrixMaterialsAndImages(t *testing.T) {
	schema := `{"questions":[{"id":"matrix","type":"matrixRadio","title":"判断说法<img src=\"https://example.invalid/diagram.png\" alt=\"分析图\">","description":"<p>根据图示判断</p><img src=\"javascript:alert(1)\">","examScore":6,"props":{"rows":[{"title":"第一条原始陈述"},{"title":"第二条原始陈述"}],"columns":[{"label":"正确","value":"yes"},{"label":"错误","value":"no"}]}}]}`
	record := model.ExamRecord{ID: 19, ExamID: 5, Status: 1, TotalScore: 6, Answers: `{"matrix":{"0":"yes","1":"no"}}`, Result: `[{"questionId":1,"fullScore":6,"gotScore":0,"needManual":true}]`}
	record.ReviewSnapshot = BuildSchemaReviewSnapshot(BuildSchemaGradingPlan(schema), 3)
	snapshot, err := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := buildRecordGradingDetail(record, model.Exam{ID: 5, Schema: `{"questions":[]}`}, snapshot, "")
	if err != nil {
		t.Fatal(err)
	}
	question := detail.Questions[0]
	raw, _ := json.Marshal(question.Materials)
	if !detail.CanGrade || !strings.Contains(string(raw), "第一条原始陈述") || !strings.Contains(string(raw), "根据图示判断") || !strings.Contains(question.Answer, "第一条原始陈述：正确") {
		t.Fatalf("grading material missing: %+v", question)
	}
	if len(question.Images) != 1 || question.Images[0].URL != "https://example.invalid/diagram.png" {
		t.Fatalf("unsafe/missing image: %+v", question.Images)
	}
	if got := gradingQuestionAnswer(`{"0":"yes","1":"no"}`, snapshot.Questions[0]); !strings.Contains(got, "第一条原始陈述：正确") {
		t.Fatalf("encoded matrix answers lost: %s", got)
	}
	if got := gradingQuestionAnswer("历史原始答案", snapshot.Questions[0]); got != "历史原始答案" {
		t.Fatalf("unknown matrix answer must remain visible: %s", got)
	}
}

func TestGradingDisablesMatrixWithoutRowDefinitions(t *testing.T) {
	record, _ := gradingFixture(t, false)
	snapshot, _ := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	snapshot.Questions[1].Type = "matrixRadio"
	detail, err := buildRecordGradingDetail(record, model.Exam{ID: 5}, snapshot, "")
	if err != nil {
		t.Fatal(err)
	}
	if detail.CanGrade {
		t.Fatal("matrix grading without row/column definitions must be disabled")
	}
}
