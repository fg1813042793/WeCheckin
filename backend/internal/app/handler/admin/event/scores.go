package event

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin-backend/backend/internal/app/service/event"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-赛事活动管理
// @Summary 获取活动评分列表(管理端)
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/event_scores [get]
func (h *AdminEventHandler) GetEventScores(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := eventservice.GetEventScores(eventID, 1, 100)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动评分
// @Param id formData string false "评分记录ID(为空时新增)"
// @Param score formData string true "评分"
// @Param eventId formData string false "活动ID(新增时必填)"
// @Param participantId formData string false "参赛者ID(新增时必填)"
// @Success 200 {object} response.Resp
// @Router /admin/event_score_edit [post]
func (h *AdminEventHandler) EditEventScore(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	score := c.PostForm("score")
	if id == "" {
		eventID := c.PostForm("eventId")
		participantID := c.PostForm("participantId")
		if eventID == "" || participantID == "" {
			response.Fail(c, "参数错误")
			return
		}
		if err := eventservice.SaveEventScore(eventID, participantID, score, "admin"); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	} else {
		if err := eventservice.AdminEditEventScore(id, score); err != nil {
			response.Fail(c, "编辑失败")
			return
		}
	}
	response.JSON(c, nil)
}
