package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type EnrollHandler struct{}

func NewEnrollHandler() *EnrollHandler { return &EnrollHandler{} }

// @Tags 报名
// @Summary 获取报名列表
// @Success 200 {object} response.Resp
// @Router /enroll/list [get]
func (h *EnrollHandler) GetEnrollList(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetEnrollList()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 查看报名详情
// @Param id query string true "报名ID"
// @Success 200 {object} response.Resp
// @Router /enroll/view [get]
func (h *EnrollHandler) ViewEnroll(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.ViewEnroll(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 按日获取报名打卡记录
// @Param enroll_id query string true "报名ID"
// @Param day query string true "日期"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/join_day [get]
func (h *EnrollHandler) GetEnrollJoinByDay(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enroll_id")
	day := c.Query("day")
	userID := c.Query("user_id")
	data, err := service.GetEnrollJoinByDay(enrollID, userID, day)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 用户报名打卡
// @Param enroll_id formData string true "报名ID"
// @Param day formData string true "日期"
// @Param user_id formData string false "用户ID"
// @Param forms formData string false "表单数据"
// @Success 200 {object} response.Resp
// @Router /enroll/join [post]
func (h *EnrollHandler) EnrollJoin(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enroll_id")
	day := c.PostForm("day")
	userID := c.PostForm("user_id")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	err := service.EnrollJoin(enrollID, userID, day, forms, addIP, 1)
	if err != nil {
		response.Fail(c, "打卡失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 报名
// @Summary 获取我的打卡记录
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/my_join_list [get]
func (h *EnrollHandler) GetMyEnrollJoinList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := service.GetMyEnrollJoinList(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 报名
// @Summary 获取我的报名用户列表
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /enroll/my_user_list [get]
func (h *EnrollHandler) GetMyEnrollUserList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := service.GetMyEnrollUserList(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
