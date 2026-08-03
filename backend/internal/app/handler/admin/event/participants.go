package event

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	eventservice "wecheckin/backend/internal/app/service/event"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-赛事活动管理
// @Summary 获取活动参与成员列表
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetEventParticipantList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := eventservice.GetEventParticipantListForAdminContext(ctx, eventID, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, listResponse{List: list})
}

// @Tags PC端-赛事活动管理
// @Summary 删除活动参与成员
// @Param id formData string true "参与记录ID"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEventParticipant(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := eventservice.DelEventParticipantForAdminContext(ctx, id, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 编辑活动参与成员信息
// @Param id formData string true "参与记录ID"
// @Param forms formData string false "表单数据(JSON)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) EditEventParticipant(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id := c.PostForm("id")
	forms := c.PostForm("forms")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := eventservice.EditEventParticipantForAdminContext(ctx, id, forms, admin.ID); err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 批量删除活动参与成员
// @Param ids formData string true "参与记录ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) DelEventParticipants(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := eventservice.DelEventParticipantsForAdminContext(ctx, ids, admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-赛事活动管理
// @Summary 获取部门用户列表
// @Param deptIds query string true "部门ID列表(逗号分隔)"
// @Success 200 {object} response.Resp
func (h *AdminEventHandler) GetDeptUsers(ctx context.Context, c *app.RequestContext) {
	deptIDsStr := c.Query("deptIds")
	if deptIDsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	var deptIDs []uint
	for _, s := range strings.Split(deptIDsStr, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(s))
		if err == nil && id > 0 {
			deptIDs = append(deptIDs, uint(id))
		}
	}
	users, err := eventservice.GetDeptUsersContext(ctx, deptIDs)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, listResponse{List: users})
}
