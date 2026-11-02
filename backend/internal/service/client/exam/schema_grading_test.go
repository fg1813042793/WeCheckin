package exam

import (
	"testing"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

func TestBuildSchemaGradingPlanExcludesPersonalAndUngradedQuestions(t *testing.T) {
	schema := `{"questions":[
      {"id":"name1","type":"name","title":"姓名"},
      {"id":"divider1","type":"divider","title":"基础题"},
      {"id":"q1","type":"input","title":"Agent全称","examScoreMode":"single","examCorrectAnswer":"Agent","examScore":5},
      {"id":"q2","type":"textarea","title":"简述","examScoreMode":"single","examScore":10}
    ]}`
	plan := BuildSchemaGradingPlan(schema)
	if len(plan.Questions) != 2 {
		t.Fatalf("graded questions = %#v", plan.Questions)
	}
	if _, ok := plan.IDMap["name1"]; ok {
		t.Fatal("personal information question must not be graded")
	}
	if plan.IDMap["q1"] != 3 || plan.IDMap["q2"] != 4 {
		t.Fatalf("original question ordinals must be preserved: %#v", plan.IDMap)
	}
	if plan.Questions[0].Answer != "Agent" || plan.Questions[0].Score != 5 {
		t.Fatalf("scalar grading config = %#v", plan.Questions[0])
	}
	if !plan.Questions[1].NeedManual {
		t.Fatal("scored textarea must require manual grading")
	}
}

func TestBuildSchemaGradingPlanCarriesMultiInputFieldScores(t *testing.T) {
	schema := `{"questions":[{"id":"q1","type":"multiInput","title":"填空","examScoreMode":"partialCorrect","props":{"fields":[{"label":"一","examCorrectAnswer":"Agent","examScore":3},{"label":"二","examCorrectAnswer":"LLM","examScore":2}]}}]}`
	plan := BuildSchemaGradingPlan(schema)
	if len(plan.Questions) != 1 || len(plan.Questions[0].Fields) != 2 {
		t.Fatalf("multi input plan = %#v", plan)
	}
	if plan.Questions[0].Fields[0].Answer != "Agent" || plan.Questions[0].Fields[1].Score != 2 {
		t.Fatalf("multi input fields = %#v", plan.Questions[0].Fields)
	}
}

func TestShouldGradePaperQuestionExcludesPersonalAndZeroScore(t *testing.T) {
	if ShouldGradePaperQuestion("name", 10) || ShouldGradePaperQuestion("input", 0) {
		t.Fatal("personal information and zero-score paper questions must not be graded")
	}
	if !ShouldGradePaperQuestion("input", 5) || !ShouldGradePaperQuestion("textarea", 10) {
		t.Fatal("scored objective and subjective paper questions must be graded")
	}
}

func TestApplySchemaSubmissionOutcomeUsesPassScoreAndManualStatus(t *testing.T) {
	record := model.ExamRecord{}
	ApplySchemaSubmissionOutcome(&record, examPkg.PaperResult{TotalScore: 75, FullScore: 100}, 60)
	if record.Status != 2 || record.Pass != 1 || record.Score != 75 || record.TotalScore != 100 {
		t.Fatalf("passed outcome = %#v", record)
	}

	record = model.ExamRecord{}
	ApplySchemaSubmissionOutcome(&record, examPkg.PaperResult{TotalScore: 50, FullScore: 100}, 60)
	if record.Status != 2 || record.Pass != 0 {
		t.Fatalf("failed outcome = %#v", record)
	}

	record = model.ExamRecord{}
	ApplySchemaSubmissionOutcome(&record, examPkg.PaperResult{TotalScore: 50, FullScore: 100, ManualCount: 1}, 60)
	if record.Status != 1 || record.Pass != 0 {
		t.Fatalf("manual outcome = %#v", record)
	}
}

func TestDecodeRecordDetailPayloadReturnsStructuredResults(t *testing.T) {
	detail := decodeRecordDetailPayload(model.ExamRecord{Result: `{"results":[{"questionId":2,"correct":true,"gotScore":5,"fullScore":5}]}`}, model.Exam{})
	if len(detail.Results) != 1 || detail.Results[0].QuestionID != 2 || !detail.Results[0].Correct {
		t.Fatalf("record detail results = %#v", detail.Results)
	}
}
