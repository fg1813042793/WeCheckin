package enroll

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	admincontentservice "wecheckin-backend/backend/internal/app/service/admincontent"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-打卡管理
// @Summary 获取参与用户列表(含报名表单数据)
// @Param enrollId query string true "项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_user_list [get]
func (h *AdminEnrollHandler) GetEnrollUserList(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	keyword := c.Query("keyword")
	list, err := admincontentservice.GetEnrollUserList(enrollID, keyword)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

func (h *AdminEnrollHandler) GetEnrollStats(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	startDay := c.Query("startTime")
	endDay := c.Query("endTime")
	if enrollID == "" {
		response.Fail(c, "参数错误")
		return
	}
	list, err := admincontentservice.GetEnrollStats(enrollID, startDay, endDay)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, list)
}

// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_list [get]
func (h *AdminEnrollHandler) GetEnrollJoinList(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("pageSize"))
	if size == 0 {
		size, _ = strconv.Atoi(c.Query("size"))
	}
	list, total, err := admincontentservice.GetEnrollJoinList(enrollID, keyword, page, size)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, pagedListResponse{List: list, Total: total})
}

// @Tags PC端-打卡管理
// @Summary 从打卡项目中移除用户（删除用户及所有打卡记录）
// @Param enrollId formData string true "项目ID"
// @Param userId formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_remove_user [post]
func (h *AdminEnrollHandler) RemoveEnrollUser(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	userID := c.PostForm("userId")
	if enrollID == "" || userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	err := admincontentservice.RemoveEnrollUser(enrollID, userID)
	if err != nil {
		response.Fail(c, "移除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) RemoveEnrollUsers(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	userIDsStr := c.PostForm("userIds")
	if enrollID == "" || userIDsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	userIDs := strings.Split(userIDsStr, ",")
	if err := admincontentservice.RemoveEnrollUsers(enrollID, userIDs); err != nil {
		response.Fail(c, "移除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) EditEnrollUserForms(ctx context.Context, c *app.RequestContext) {
	enrollID := c.PostForm("enrollId")
	userID := c.PostForm("userId")
	forms := c.PostForm("forms")
	if enrollID == "" || userID == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := admincontentservice.EditEnrollUserForms(enrollID, userID, forms); err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-打卡管理
// @Summary 删除打卡记录
// @Param id formData string true "记录ID"
// @Success 200 {object} response.Resp
// @Router /admin/enroll_join_del [post]
func (h *AdminEnrollHandler) DelEnrollJoin(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("enrollJoinId")
	if id == "" {
		id = c.PostForm("id")
	}
	err := admincontentservice.DelEnrollJoin(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminEnrollHandler) DelEnrollJoins(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := admincontentservice.DelEnrollJoins(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}
