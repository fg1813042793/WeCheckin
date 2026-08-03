package exam

import (
	"context"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	examservice "wecheckin/backend/internal/app/service/exam"
	"wecheckin/backend/pkg/response"
)

// Start GET /exam/start?examId=
// @Tags 客户端-考试
// @Summary 开始考试
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Resp
// @Router /exam/start [get]
func (h *ClientExamHandler) Start(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	examID, _ := strconv.Atoi(c.Query("examId"))
	if examID == 0 {
		response.Fail(c, "examId 必填")
		return
	}
	deviceId := c.Query("deviceId")
	result, err := h.service().StartContext(ctx, examID, uid, deviceId)
	if err != nil {
		switch {
		case errors.Is(err, examservice.ErrExamNotPublished):
			response.Fail(c, "考试未发布")
		case errors.Is(err, examservice.ErrExamMaxAttempts):
			response.Fail(c, "已达最大尝试次数")
		case errors.Is(err, examservice.ErrExamNotStarted):
			response.Fail(c, "考试未开始")
		case errors.Is(err, examservice.ErrExamEnded):
			response.Fail(c, "考试已结束")
		case errors.Is(err, examservice.ErrExamPaperNotFound):
			response.Fail(c, "试卷不存在")
		default:
			response.Fail(c, "考试不存在")
		}
		return
	}
	response.JSON(c, examStartResponse{
		Record:    result.Record,
		Paper:     result.Paper,
		Exam:      result.Exam,
		Questions: result.Questions,
		Answers:   result.Answers,
	})
}
