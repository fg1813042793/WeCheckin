package exam

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	examPkg "wecheckin-backend/backend/internal/app/formkit/exam"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// Record GET /exam/record?id=
// @Tags 客户端-考试
// @Summary 查看考试记录
// @Param id query int true "记录ID"
// @Success 200 {object} response.Resp
// @Router /exam/record [get]
func (h *ClientExamHandler) Record(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	id, _ := strconv.Atoi(c.Query("id"))
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var rec model.ExamRecord
	if err := db.Where("`exam_r_id` = ? AND `exam_r_user_id` = ?", id, uid).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	var p model.ExamPaper
	db.Where("`exam_p_id` = ?", rec.PaperID).First(&p)
	var e model.Exam
	db.Where("`exam_id` = ?", rec.ExamID).First(&e)
	var qids []uint
	_ = json.Unmarshal([]byte(p.QuestionIDs), &qids)
	var qs []model.ExamQuestion
	if len(qids) > 0 {
		db.Where("`exam_q_id` IN ?", qids).Find(&qs)
	}
	safe := make([]map[string]interface{}, 0, len(qs))
	for _, q := range qs {
		item := map[string]interface{}{
			"id":      q.ID,
			"type":    q.Type,
			"title":   q.Title,
			"options": q.Options,
			"score":   q.Score,
		}
		if p.ShowAnswer == 1 || rec.Status == 2 {
			item["answer"] = q.Answer
			item["analysis"] = q.Analysis
		}
		safe = append(safe, item)
	}
	var prevAnswers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &prevAnswers)
	}
	var results []examPkg.Result
	if rec.Result != "" {
		_ = json.Unmarshal([]byte(rec.Result), &results)
	}
	response.JSON(c, examRecordResponse{Record: rec, Exam: e, Paper: p, Questions: safe, Answers: prevAnswers, Results: results})
}

// MyRecords GET /exam/my_records
// @Tags 客户端-考试
// @Summary 我的考试记录
// @Success 200 {object} response.Resp
// @Router /exam/my_records [get]
func (h *ClientExamHandler) MyRecords(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	var list []model.ExamRecord
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Where("`exam_r_user_id` = ?", uid).Order("`exam_r_id` DESC").Limit(50).Find(&list)
	response.JSON(c, examMyRecordsResponse{List: list})
}

// ResultBySession GET /exam/result?session=
// @Tags 客户端-考试
// @Summary 通过 session 查看考试结果
// @Router /exam/result [get]
func (h *ClientExamHandler) ResultBySession(ctx context.Context, c *app.RequestContext) {
	session := c.Query("session")
	if session == "" {
		response.Fail(c, "参数错误")
		return
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var rec model.ExamRecord
	if err := db.Where("`exam_r_session` = ?", session).First(&rec).Error; err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	var e model.Exam
	db.Where("`exam_id` = ?", rec.ExamID).First(&e)
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Settings), &settingsMap)
	if settingsMap == nil {
		settingsMap = make(map[string]interface{})
	}
	if _, ok := settingsMap["answerVisible"]; !ok {
		settingsMap["answerVisible"] = true
	}
	var questions []map[string]interface{}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(e.Schema), &schMap)
	rawQuestions, _ := schMap["questions"].([]interface{})
	for _, raw := range rawQuestions {
		q, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := q["id"].(string)
		if id == "" {
			continue
		}
		item := map[string]interface{}{
			"id":    id,
			"type":  q["type"],
			"title": q["title"],
		}
		if showAnalysis, _ := settingsMap["showAnalysis"].(bool); showAnalysis {
			item["examCorrectAnswer"] = q["examCorrectAnswer"]
			item["examAnalysis"] = q["examAnalysis"]
		}
		questions = append(questions, item)
	}
	var answers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &answers)
	}
	var results []examPkg.Result
	if rec.Result != "" {
		_ = json.Unmarshal([]byte(rec.Result), &results)
	}
	response.JSON(c, examResultBySessionResponse{Exam: e, Record: rec, Questions: questions, Answers: answers, Results: results, Settings: settingsMap})
}
