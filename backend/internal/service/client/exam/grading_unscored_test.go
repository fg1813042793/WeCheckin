package exam

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

const unscoredReviewSchema = `{"questions":[{"id":"name","type":"name","title":"姓名"},{"id":"section","type":"divider","title":"题目"},{"id":"essay","type":"textarea","title":"介绍你的项目","examScoreMode":"none","examScore":99,"examCorrectAnswer":"旧参考答案"},{"id":"empty","type":"input","title":"补充说明","examScoreMode":"none"}]}`

func TestGradingShowsAllUnscoredAnswersAndFreezesNewSnapshots(t *testing.T) {
	plan := BuildSchemaGradingPlan(unscoredReviewSchema)
	if len(plan.Questions) != 0 {
		t.Fatal("unscored questions entered grading engine")
	}
	record := model.ExamRecord{ID: 22, ExamID: 5, Status: 2, Answers: `{"name":"测试考生","essay":"项目经历"}`, Result: `{"results":[],"totalScore":0,"fullScore":0}`}
	record.ReviewSnapshot = BuildSchemaReviewSnapshot(plan, 0)
	snapshot, err := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Questions) != 3 {
		t.Fatalf("snapshot must retain all answerable questions, got %d", len(snapshot.Questions))
	}
	detail, err := buildRecordGradingDetail(record, model.Exam{Schema: `{"questions":[]}`}, snapshot, "")
	if err != nil || len(detail.Questions) != 3 || detail.CanGrade {
		t.Fatalf("unscored review = %+v, err=%v", detail, err)
	}
	if detail.Questions[1].Title != "介绍你的项目" || detail.Questions[1].Answer != "项目经历" || detail.Questions[2].Answer != "未作答" {
		t.Fatalf("original questions/answers lost: %+v", detail.Questions)
	}
	for _, question := range detail.Questions {
		assertReviewUnscored(t, question)
	}
	if detail.Record.Result != record.Result || detail.Record.Score != 0 || detail.Record.TotalScore != 0 {
		t.Fatal("review must not rewrite grading result")
	}
}

func TestGradingSupplementsLegacyUnscoredAnswersWithoutInventingNewQuestions(t *testing.T) {
	record := model.ExamRecord{ID: 22, ExamID: 5, Status: 2, Answers: `{"name":"测试考生","essay":"旧答案"}`, Result: `[]`, ReviewSnapshot: `{"version":1,"passScore":0,"questions":[]}`}
	snapshot, _ := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	detail, err := buildRecordGradingDetail(record, model.Exam{Schema: unscoredReviewSchema}, snapshot, "")
	if err != nil || len(detail.Questions) != 2 {
		t.Fatalf("old unscored answer review = %+v, err=%v", detail, err)
	}
	if detail.Questions[1].Answer != "旧答案" || !strings.Contains(detail.Notice, "当前配置") {
		t.Fatalf("legacy fallback must explain its source: %+v", detail)
	}
	if detail.Record.ReviewSnapshot != record.ReviewSnapshot {
		t.Fatal("legacy snapshot must not be rewritten")
	}
	for _, question := range detail.Questions {
		assertReviewUnscored(t, question)
	}
}

func TestMixedReviewKeepsUnscoredOutOfManualGrading(t *testing.T) {
	schema := `{"questions":[{"id":"info","type":"textarea","title":"补充信息","examScoreMode":"none"},{"id":"graded","type":"textarea","title":"计分题","examScoreMode":"manual","examScore":5}]}`
	plan := BuildSchemaGradingPlan(schema)
	result := examPkg.Grade(plan.Questions, map[string]interface{}{"2": "需要评分"})
	encoded, _ := json.Marshal(result)
	record := model.ExamRecord{ID: 22, ExamID: 5, Status: 1, TotalScore: 5, Answers: `{"info":"只展示","graded":"需要评分"}`, Result: string(encoded), ReviewSnapshot: BuildSchemaReviewSnapshot(plan, 3)}
	snapshot, _ := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	detail, err := buildRecordGradingDetail(record, model.Exam{}, snapshot, "")
	if err != nil || len(detail.Questions) != 2 || !detail.CanGrade {
		t.Fatalf("mixed review = %+v, err=%v", detail, err)
	}
	assertReviewUnscored(t, detail.Questions[0])
	score := 4
	request := GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 2, Score: &score}}}
	updated, err := applyManualGrading(record, detail, request, 8, 1000)
	if err != nil || updated.Score != 4 || updated.TotalScore != 5 || len(DecodeRecordResults(updated.Result)) != 1 {
		t.Fatalf("unscored question affected grade: %+v, err=%v", updated, err)
	}
	request.Scores = append(request.Scores, ManualQuestionScore{QuestionID: 1, Score: &score})
	if _, err := applyManualGrading(record, detail, request, 8, 1000); !errors.Is(err, ErrRecordGradingInvalid) {
		t.Fatalf("unscored question must reject scores: %v", err)
	}
}

func TestPaperReviewShowsUnscoredSnapshotQuestions(t *testing.T) {
	questions := []model.ExamQuestion{{ID: 9, Type: "textarea", Title: "纸卷信息", Score: 99, Schema: `{"examScoreMode":"none"}`}}
	record := model.ExamRecord{ID: 22, ExamID: 5, PaperID: 3, Status: 2, Answers: `{"9":"纸卷答案"}`, Result: `[]`, ReviewSnapshot: BuildPaperReviewSnapshot(questions, 0)}
	snapshot, _ := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	detail, err := buildRecordGradingDetail(record, model.Exam{}, snapshot, "")
	if err != nil || len(detail.Questions) != 1 || detail.Questions[0].Answer != "纸卷答案" {
		t.Fatalf("paper unscored review = %+v, err=%v", detail, err)
	}
	assertReviewUnscored(t, detail.Questions[0])
}

func TestLegacySupplementPreservesScoredQuestionsAndAvoidsCollisions(t *testing.T) {
	record, _ := gradingFixture(t, true)
	record.Answers = `{"a":"A","essay":"我的答案","new":"新增内容","info":"附加信息"}`
	snapshot, _ := decodeRecordReviewSnapshot(record.ReviewSnapshot)
	exam := model.Exam{Schema: `{"questions":[{"id":"new","type":"input","title":"冲突序号","examScoreMode":"none"},{"id":"a","type":"radio","title":"后来改名","examScoreMode":"none"},{"id":"essay","type":"textarea","title":"后来改为不计分","examScoreMode":"none"},{"id":"info","type":"textarea","title":"附加信息","examScoreMode":"none"}]}`}
	detail, err := buildRecordGradingDetail(record, exam, snapshot, "")
	if err != nil || !detail.CanGrade || len(detail.Questions) != 3 {
		t.Fatalf("legacy mixed review = %+v, err=%v", detail, err)
	}
	if detail.Questions[0].Title != "选择题" || detail.Questions[0].GotScore != 4 || detail.Questions[0].Unscored || detail.Questions[1].Title != "简答题" || !detail.Questions[1].NeedManual {
		t.Fatalf("frozen scored questions were changed: %+v", detail.Questions)
	}
	assertReviewUnscored(t, detail.Questions[2])
	if len(snapshot.Questions) != 2 || snapshot.Questions[1].Title != "简答题" {
		t.Fatal("read-only supplement mutated caller snapshot")
	}
}

func assertReviewUnscored(t *testing.T, question GradingQuestion) {
	t.Helper()
	raw, _ := json.Marshal(question)
	var payload map[string]interface{}
	_ = json.Unmarshal(raw, &payload)
	if payload["unscored"] != true || question.NeedManual || question.GotScore != 0 || question.FullScore != 0 || question.ReferenceAnswer != "" {
		t.Fatalf("unscored contract = %s", raw)
	}
}
