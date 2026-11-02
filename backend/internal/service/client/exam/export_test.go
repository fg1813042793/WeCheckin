package exam

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

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

func TestExamSchemaFromPaperQuestionsPreservesQuestionIDs(t *testing.T) {
	schemaJSON := examSchemaFromPaperQuestions([]model.ExamQuestion{{
		ID: 42, Type: "radio", Title: "历史试卷题目", Options: `[{"label":"选项A","value":"A"}]`, Answer: `"A"`, Score: 10,
	}})
	document := buildRecordExportDocument(
		model.Exam{ID: 10, Title: "历史试卷", Schema: schemaJSON},
		model.ExamRecord{ID: 13, ExamID: 10, UserID: "8", Answers: `{"42":"A"}`, Result: `{"results":[{"questionId":42,"correct":true,"gotScore":10,"fullScore":10}]}`, Score: 10, TotalScore: 10, Pass: 1, Status: 2},
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
