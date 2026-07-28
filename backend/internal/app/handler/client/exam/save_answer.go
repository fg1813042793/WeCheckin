package exam

import (
	"context"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	examservice "wecheckin-backend/backend/internal/app/service/exam"
	"wecheckin-backend/backend/pkg/response"
)

// SaveAnswer POST /exam/save_answer
// @Tags 客户端-考试
// @Summary 保存答案
// @Param recordId formData int true "记录ID"
// @Param answers formData string true "答案JSON"
// @Success 200 {object} response.Resp
// @Router /exam/save_answer [post]
func (h *ClientExamHandler) SaveAnswer(ctx context.Context, c *app.RequestContext) {
	uid := getUID(c)
	if uid == 0 {
		response.Fail(c, "未登录")
		return
	}
	recordID, _ := strconv.Atoi(c.PostForm("recordId"))
	answersJSON := c.PostForm("answers")
	if recordID == 0 {
		response.Fail(c, "recordId 必填")
		return
	}
	if err := h.service().SaveAnswerContext(ctx, recordID, uid, answersJSON); err != nil {
		if errors.Is(err, examservice.ErrExamRecordSubmitted) {
			response.Fail(c, "已提交，不可修改")
			return
		}
		if errors.Is(err, examservice.ErrExamRecordNotFound) {
			response.Fail(c, "记录不存在")
			return
		}
		response.Fail(c, "保存失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
