package exam

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	examPkg "wecheckin/backend/internal/app/formkit/exam"
	examservice "wecheckin/backend/internal/app/service/exam"
	"wecheckin/backend/pkg/response"
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
	result, err := h.service().RecordForUserContext(ctx, id, uid)
	if err != nil {
		response.Fail(c, "记录不存在")
		return
	}
	response.JSON(c, examRecordResponse{
		Record:    result.Record,
		Exam:      result.Exam,
		Paper:     result.Paper,
		Questions: result.Questions,
		Answers:   result.Answers,
		Results:   result.Results,
	})
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
	list, err := h.service().UserRecordsContext(ctx, uid, 50)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
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
	result, err := h.service().SessionResultContext(ctx, session)
	if err != nil {
		if errors.Is(err, examservice.ErrExamRecordNotFound) {
			response.Fail(c, "记录不存在")
			return
		}
		response.Fail(c, "记录不存在")
		return
	}
	var settingsMap map[string]interface{}
	_ = json.Unmarshal([]byte(result.Exam.Settings), &settingsMap)
	if settingsMap == nil {
		settingsMap = make(map[string]interface{})
	}
	if _, ok := settingsMap["answerVisible"]; !ok {
		settingsMap["answerVisible"] = true
	}
	var questions []map[string]interface{}
	var schMap map[string]interface{}
	_ = json.Unmarshal([]byte(result.Exam.Schema), &schMap)
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
	if result.Record.Answers != "" {
		_ = json.Unmarshal([]byte(result.Record.Answers), &answers)
	}
	var results []examPkg.Result
	if result.Record.Result != "" {
		_ = json.Unmarshal([]byte(result.Record.Result), &results)
	}
	response.JSON(c, examResultBySessionResponse{Exam: result.Exam, Record: result.Record, Questions: questions, Answers: answers, Results: results, Settings: settingsMap})
}
