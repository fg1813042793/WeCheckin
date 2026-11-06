package exam

import (
	"fmt"
	"strings"

	examPkg "wecheckin/backend/internal/formkit/exam"
	"wecheckin/backend/internal/model"
)

// Version 1 schema snapshots omitted unscored questions. Recover only saved
// answers whose current definitions remain unscored, never replace frozen data.
func supplementLegacyUnscoredQuestions(record model.ExamRecord, exam model.Exam, snapshot recordReviewSnapshot, answers map[string]interface{}, results []examPkg.Result, notice string) (recordReviewSnapshot, string) {
	if snapshot.Version != 1 || record.PaperID > 0 {
		return snapshot, notice
	}
	plan := BuildSchemaGradingPlan(exam.Schema)
	occupiedIDs := make(map[uint]bool)
	occupiedKeys := make(map[string]bool)
	for _, question := range snapshot.Questions {
		occupiedIDs[question.QuestionID], occupiedKeys[question.AnswerKey] = true, true
	}
	for _, result := range results {
		occupiedIDs[result.QuestionID] = true
	}
	for _, question := range plan.Questions {
		occupiedIDs[question.ID] = true
	}
	added := false
	questions := append([]reviewQuestionSnapshot(nil), snapshot.Questions...)
	for _, question := range plan.ReviewQuestions {
		if _, answered := answers[question.AnswerKey]; !answered || occupiedIDs[question.QuestionID] || occupiedKeys[question.AnswerKey] {
			continue
		}
		questions = append(questions, question)
		occupiedIDs[question.QuestionID], occupiedKeys[question.AnswerKey] = true, true
		added = true
	}
	if added {
		snapshot.Questions = questions
		notice = strings.TrimSpace(notice + " 历史答卷未保存部分不计分题的快照，缺失题目按当前配置补充展示，原始答案和成绩保持不变。")
	}
	return snapshot, notice
}

func unscoredGradingQuestion(question reviewQuestionSnapshot, answers map[string]interface{}) GradingQuestion {
	item := GradingQuestion{QuestionID: question.QuestionID, Type: question.Type, Title: stripExamExportHTML(question.Title), Answer: "未作答", Unscored: true}
	if item.Title == "" {
		item.Title = fmt.Sprintf("题目 #%d", question.QuestionID)
	}
	item.Materials, item.Images, _, _ = gradingMaterials(question)
	if value, exists := answers[question.AnswerKey]; exists {
		item.Answer = gradingQuestionAnswer(value, question)
	}
	return item
}
