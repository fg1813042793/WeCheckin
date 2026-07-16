package event

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin-backend/backend/internal/app/service/event"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags 客户端-赛事活动
// @Summary 保存活动评分
// @Param event_id formData string true "活动ID"
// @Param participant_id formData string true "参赛者ID"
// @Param score formData string true "评分"
// @Param judge_id formData string true "评委ID"
// @Success 200 {object} response.Resp
// @Router /event/score_save [post]
func (h *EventHandler) SaveEventScore(ctx context.Context, c *app.RequestContext) {
	eventID := c.PostForm("event_id")
	participantID := c.PostForm("participant_id")
	score := c.PostForm("score")
	judgeID := c.PostForm("judge_id")
	if eventID == "" || participantID == "" || score == "" || judgeID == "" {
		response.Fail(c, "参数错误")
		return
	}
	err := eventservice.SaveEventScore(eventID, participantID, score, judgeID)
	if err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 客户端-赛事活动
// @Summary 获取活动评分列表
// @Param event_id query string true "活动ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /event/scores [get]
func (h *EventHandler) GetEventScores(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("event_id")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := eventservice.GetEventScores(eventID, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, result)
}
