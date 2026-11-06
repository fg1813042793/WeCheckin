package exam

import (
	"encoding/json"
	"testing"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

func TestExplicitNoneSchemaSkipsStaleScoringAndPreservesOrdinals(t *testing.T) {
	schema := `{"questions":[{"id":"free","type":"textarea","required":true,"examScoreMode":"none","examScore":50,"examCorrectAnswer":"旧参考答案"},{"id":"section","type":"divider"},{"id":"blank","type":"multiInput","examScoreMode":"none","props":{"fields":[{"examScore":7,"examCorrectAnswer":"残留","examCorrect":true}]}},{"id":"graded","type":"input","examScoreMode":"single","examScore":5,"examCorrectAnswer":"A"}]}`
	plan := BuildSchemaGradingPlan(schema)
	if len(plan.Questions) != 1 || plan.IDMap["graded"] != 4 {
		t.Fatalf("plan=%+v", plan)
	}
	if _, ok := plan.IDMap["free"]; ok {
		t.Fatal("unscored textarea mapped for grading")
	}
	if _, ok := plan.IDMap["blank"]; ok {
		t.Fatal("stale field scores re-enabled grading")
	}
	result := examPkg.Grade(plan.Questions, map[string]interface{}{"4": "A"})
	if result.TotalScore != 5 || result.FullScore != 5 || result.ManualCount != 0 || result.CorrectCnt != 1 {
		t.Fatalf("result=%+v", result)
	}
	snapshot, err := decodeRecordReviewSnapshot(BuildSchemaReviewSnapshot(plan, 3))
	if err != nil || len(snapshot.Questions) != 3 || snapshot.Questions[2].QuestionID != 4 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	raw, _ := json.Marshal(result)
	record := model.ExamRecord{Status: 2, Score: 5, TotalScore: 5, Answers: `{"free":"反馈文字","blank":["补充信息"],"graded":"A"}`, Result: string(raw)}
	detail, err := buildRecordGradingDetail(record, model.Exam{}, snapshot, "")
	if err != nil || detail.CanGrade || len(detail.Questions) != 3 || !detail.Questions[0].Unscored || !detail.Questions[1].Unscored || detail.Questions[2].Unscored {
		t.Fatalf("grading detail=%+v err=%v", detail, err)
	}
}

func TestAllExplicitNoneCompletesWithoutManualWork(t *testing.T) {
	plan := BuildSchemaGradingPlan(`{"questions":[{"id":"free","type":"textarea","required":true,"examScoreMode":"none","examScore":99}]}`)
	result := examPkg.Grade(plan.Questions, nil)
	record := model.ExamRecord{}
	ApplySchemaSubmissionOutcome(&record, result, 0)
	if record.Status != 2 || record.TotalScore != 0 || record.Score != 0 || record.Pass != 1 || result.ManualCount != 0 || len(result.Results) != 0 {
		t.Fatalf("outcome=%+v result=%+v", record, result)
	}
}

func TestManualGradingMixedNoneOnlyRequiresScoredAnswers(t *testing.T) {
	plan := BuildSchemaGradingPlan(`{"questions":[{"id":"feedback","type":"textarea","examScoreMode":"none","examScore":50},{"id":"essay","type":"textarea","examScoreMode":"manual","examScore":10}]}`)
	result := examPkg.Grade(plan.Questions, map[string]interface{}{"2": "评分内容"})
	raw, _ := json.Marshal(result)
	record := model.ExamRecord{ID: 1, ExamID: 1, Status: 1, TotalScore: 10, Answers: `{"feedback":"附加说明","essay":"评分内容"}`, Result: string(raw), ReviewSnapshot: BuildSchemaReviewSnapshot(plan, 6)}
	snapshot, err := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := buildRecordGradingDetail(record, model.Exam{}, snapshot, "")
	if err != nil || !detail.CanGrade || len(detail.Questions) != 2 || !detail.Questions[0].Unscored || detail.Questions[1].QuestionID != 2 {
		t.Fatalf("pending=%+v %v", detail, err)
	}
	score := 8
	updated, err := applyManualGrading(record, detail, GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 2, Score: &score}}}, 7, 1234)
	if err != nil || updated.Status != 2 || updated.Score != 8 || updated.TotalScore != 10 || updated.Answers != record.Answers {
		t.Fatalf("manual result=%+v %v", updated, err)
	}
}

func TestPaperExplicitNoneOverridesStoredScore(t *testing.T) {
	q := model.ExamQuestion{ID: 10, Type: "textarea", Score: 20, Schema: `{"examScoreMode":"none"}`}
	if ShouldGradePaperQuestionDefinition(q) {
		t.Fatal("explicit none paper question still graded")
	}
	q.Schema = `{}`
	if !ShouldGradePaperQuestionDefinition(q) {
		t.Fatal("legacy paper grading changed")
	}
}

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

func TestBuildSchemaGradingPlanSupportsExplicitManualModeForAnyQuestionType(t *testing.T) {
	schema := `{"questions":[{"id":"q1","type":"input","title":"开放题","examScoreMode":"manual","examCorrectAnswer":"参考要点","examScore":7}]}`
	plan := BuildSchemaGradingPlan(schema)
	if len(plan.Questions) != 1 || plan.Questions[0].ScoreMode != "manual" || !plan.Questions[0].NeedManual {
		t.Fatalf("manual grading plan = %#v", plan)
	}
	result := examPkg.Grade(plan.Questions, map[string]interface{}{"1": "考生自由作答"})
	if result.ManualCount != 1 || result.TotalScore != 0 || result.FullScore != 7 {
		t.Fatalf("manual grading result = %#v", result)
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
