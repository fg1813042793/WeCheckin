package event

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin/backend/internal/service/client/event"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-赛事活动管理
// @Summary 获取活动评分列表(管理端)
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetEventScores(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := eventservice.GetEventScoresForAdminContext(ctx, eventID, 1, 100, admin.ID)
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
func (h *AdminEventHandler) EditEventScore(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	score := c.PostForm("score")
	if id == "" {
		eventID := c.PostForm("eventId")
		participantID := c.PostForm("participantId")
		if eventID == "" || participantID == "" {
			response.Fail(c, "参数错误")
			return
		}
		if err := eventservice.SaveEventScoreForAdminContext(ctx, eventID, participantID, score, "admin", admin.ID); err != nil {
			response.Fail(c, "保存失败")
			return
		}
	} else {
		if err := eventservice.AdminEditEventScoreForAdminContext(ctx, id, score, admin.ID); err != nil {
			response.Fail(c, "编辑失败")
			return
		}
	}
	response.JSON(c, nil)
}
