package exam

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
)

func TestExportExplicitNoneKeepsAnswersAndDoesNotBorrowFollowingGrade(t *testing.T) {
	// Numeric schema IDs are answer keys, not grading ordinals. The divider and
	// excluded question must not shift the following scored result onto this row.
	e := model.Exam{Schema: `{"questions":[{"id":"section","type":"divider","title":"题组"},{"id":"3","type":"textarea","title":"补充说明","examScoreMode":"none","examScore":99,"examCorrectAnswer":"旧答案"},{"id":"graded","type":"input","title":"计分题","examScoreMode":"single","examScore":5}]}`}
	r := model.ExamRecord{Answers: `{"3":"保留的非计分回答","graded":"A"}`, Result: `{"results":[{"questionId":3,"correct":true,"gotScore":5,"fullScore":5}]}`}
	d := buildRecordExportDocument(e, r, "")
	rows := d.Sections[0].Rows
	if len(rows) != 2 || rows[0][1] != "保留的非计分回答" || rows[0][2] != "不适用（不计分）" || rows[0][3] != "不计分" || rows[0][4] != "不计分" {
		t.Fatalf("unscored row=%#v", rows)
	}
	if rows[1][3] != "5/5" || rows[1][4] != "正确" {
		t.Fatalf("scored result shifted: %#v", rows)
	}
}

func TestExportKeepsStoredScoreWhenCurrentDefinitionBecomesUnscored(t *testing.T) {
	e := model.Exam{Schema: `{"questions":[{"id":"graded","type":"input","title":"旧计分题","examScoreMode":"none","examScore":0}]}`}
	r := model.ExamRecord{Answers: `{"graded":"A"}`, Result: `{"results":[{"questionId":1,"correct":true,"gotScore":5,"fullScore":5}]}`}
	row := buildRecordExportDocument(e, r, "").Sections[0].Rows[0]
	if row[3] != "5/5" || row[4] != "正确" {
		t.Fatalf("historical grading was reinterpreted: %#v", row)
	}
}

func TestExportUsesSnapshotGradeIDsAfterSchemaChanges(t *testing.T) {
	e := model.Exam{Schema: `{"questions":[{"id":"added","type":"input","title":"新收集项","examScoreMode":"none","examScore":99},{"id":"graded","type":"input","title":"原计分题","examScoreMode":"none","examScore":0}]}`}
	snapshot, _ := json.Marshal(recordReviewSnapshot{Version: 1, Questions: []reviewQuestionSnapshot{{QuestionID: 5, AnswerKey: "graded", Title: "原计分题", Type: "input"}}})
	r := model.ExamRecord{ReviewSnapshot: string(snapshot), Answers: `{"graded":"A"}`, Result: `{"results":[{"questionId":5,"correct":true,"gotScore":5,"fullScore":5}]}`}
	rows := buildRecordExportDocument(e, r, "").Sections[0].Rows
	if rows[0][3] != "不计分" || rows[1][3] != "5/5" || rows[1][4] != "正确" {
		t.Fatalf("snapshot score mapping=%#v", rows)
	}
}

func TestPaperExportPreservesExplicitUnscoredMode(t *testing.T) {
	schema := examSchemaFromPaperQuestions([]model.ExamQuestion{{ID: 42, Type: "textarea", Title: "补充说明", Schema: `{"examScoreMode":"none"}`, Score: 20}})
	d := buildRecordExportDocument(model.Exam{PaperID: 1, Schema: schema}, model.ExamRecord{PaperID: 1, Answers: `{"42":"我的反馈"}`, Result: `{"results":[]}`}, "")
	row := d.Sections[0].Rows[0]
	if row[1] != "我的反馈" || row[3] != "不计分" || row[4] != "不计分" {
		t.Fatalf("paper unscored row=%#v", row)
	}
}

func TestBuildRecordExportDocumentIncludesUserAnswersAndScoring(t *testing.T) {
	schema := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"id": "q1", "type": "radio", "title": "<p>中国首都是哪里？</p>",
				"examCorrectAnswer": "beijing", "examScore": float64(5),
				"props": map[string]interface{}{"options": []interface{}{
					map[string]interface{}{"label": "北京", "value": "beijing"},
					map[string]interface{}{"label": "上海", "value": "shanghai"},
				}},
			},
			map[string]interface{}{"id": "divider1", "type": "divider", "title": "第二部分"},
		},
	}
	schemaJSON, _ := json.Marshal(schema)
	resultJSON := `{"results":[{"questionId":1,"correct":true,"gotScore":5,"fullScore":5}],"totalScore":5,"fullScore":5}`
	document := buildRecordExportDocument(
		model.Exam{ID: 9, Title: "月度考试", Schema: string(schemaJSON)},
		model.ExamRecord{ID: 12, ExamID: 9, UserID: "7", Answers: `{"q1":"beijing"}`, Result: resultJSON, Score: 5, TotalScore: 5, Pass: 1, Status: 2},
		"张三",
	)

	flat := document.Title
	for _, row := range document.Metadata {
		flat += strings.Join(row, "")
	}
	for _, section := range document.Sections {
		flat += section.Title + strings.Join(section.Headers, "")
		for _, row := range section.Rows {
			flat += strings.Join(row, "")
		}
	}
	for _, want := range []string{"月度考试", "张三", "中国首都是哪里？", "北京", "正确", "5/5"} {
		if !strings.Contains(flat, want) {
			t.Fatalf("export document missing %q: %s", want, flat)
		}
	}
	if strings.Contains(flat, "divider1") {
		t.Fatalf("layout-only question should not be exported: %s", flat)
	}
}

func TestExamExportUsesReadableQuestionBlocksAndKeepsSpreadsheetRows(t *testing.T) {
	document := buildRecordExportDocument(model.Exam{ID: 5, Title: "工程师考试", Schema: `{"questions":[{"id":"part","type":"divider","title":"Agent 相关题目"},{"id":"q1","type":"textarea","title":"什么是 Agent？","examScore":10},{"id":"q2","type":"input","title":"姓名","examScore":0}]}`}, model.ExamRecord{ID: 19, Answers: `{"q1":"第一段答案。\n第二段答案。","q2":"演示考生"}`, Result: `[{"questionId":2,"needManual":true,"fullScore":10,"gotScore":0,"reason":"主观题，需人工判分"},{"questionId":3,"correct":true,"fullScore":0,"gotScore":0}]`}, "演示考生")
	if document.Layout != workflowsummary.DocumentLayoutAnswerSheet {
		t.Fatal("exam exports must opt into answer-sheet reading layout")
	}
	if len(document.Sections) != 1 || len(document.Sections[0].AnswerBlocks) != 2 {
		t.Fatalf("sections=%#v", document.Sections)
	}
	section := document.Sections[0]
	answer := section.AnswerBlocks[0]
	if answer.Title != "1. 什么是 Agent？" || answer.Answer != "第一段答案。\n第二段答案。" || answer.ReferenceAnswer != "未设置参考答案" || answer.Score != "待评分 / 10" || answer.Result != "待人工评分" {
		t.Fatalf("reading block=%#v", answer)
	}
	if len(section.Rows) != 2 || section.Rows[0][1] != answer.Answer || section.Rows[0][2] != answer.ReferenceAnswer || section.Headers[2] != "参考答案" {
		t.Fatalf("spreadsheet content drifted: %#v", section.Rows)
	}
	if section.AnswerBlocks[1].Score != "不计分" {
		t.Fatalf("zero-score field=%#v", section.AnswerBlocks[1])
	}
}

func TestExamExportReferencePlaceholderDoesNotConfuseUnansweredOrZero(t *testing.T) {
	question := map[string]interface{}{"type": "input"}
	for _, value := range []interface{}{nil, "", " ", "null", `""`, []interface{}{}, map[string]interface{}{}} {
		if got := formatExamReferenceValue(question, value); got != "未设置参考答案" {
			t.Fatalf("empty reference %v=%q", value, got)
		}
	}
	for _, value := range []interface{}{0, float64(0), "0"} {
		if got := formatExamReferenceValue(question, value); got != "0" {
			t.Fatalf("zero reference=%q", got)
		}
	}
	if got := formatExamReferenceValue(question, "未作答"); got != "未作答" {
		t.Fatalf("literal configured answer lost: %q", got)
	}
	if got := formatExamQuestionValue(question, nil); got != "未作答" {
		t.Fatalf("unanswered placeholder=%q", got)
	}
}

func TestExamExportKeepsLongTextPunctuationAndParagraphs(t *testing.T) {
	text := "Agent, LLM 和 Tools 相互配合。\n第二段保留原文, 不转换逗号。"
	if got := formatExamQuestionValue(map[string]interface{}{"type": "textarea"}, text); got != text {
		t.Fatalf("answer text changed: %q", got)
	}
}

func TestExamSchemaFromPaperQuestionsPreservesQuestionIDs(t *testing.T) {
	schemaJSON := examSchemaFromPaperQuestions([]model.ExamQuestion{{
		ID: 42, Type: "radio", Title: "历史试卷题目", Options: `[{"label":"选项A","value":"A"}]`, Answer: `"A"`, Score: 10,
	}})
	document := buildRecordExportDocument(
		model.Exam{ID: 10, PaperID: 1, Title: "历史试卷", Schema: schemaJSON},
		model.ExamRecord{ID: 13, ExamID: 10, PaperID: 1, UserID: "8", Answers: `{"42":"A"}`, Result: `{"results":[{"questionId":42,"correct":true,"gotScore":10,"fullScore":10}]}`, Score: 10, TotalScore: 10, Pass: 1, Status: 2},
		"李四",
	)
	flat := ""
	for _, section := range document.Sections {
		for _, row := range section.Rows {
			flat += strings.Join(row, "")
		}
	}
	for _, want := range []string{"历史试卷题目", "选项A", "10/10", "正确"} {
		if !strings.Contains(flat, want) {
			t.Fatalf("paper-mode export missing %q: %s", want, flat)
		}
	}
}
