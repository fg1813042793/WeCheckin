package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// ChannelList GET /admin/survey/channel_list?surveyId=
// @Tags PC端-问卷管理
// @Summary 渠道列表
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ChannelList(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	list, err := h.survey.ChannelListForAdminContext(ctx, uint(surveyID), admin.ID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.JSON(c, surveyChannelListResponse{List: list})
}

// ChannelInsert POST /admin/survey/channel_insert
// @Tags PC端-问卷管理
// @Summary 创建渠道
// @Param channel body model.SurveyChannel true "渠道数据"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ChannelInsert(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	var ch model.SurveyChannel
	if err := c.BindAndValidate(&ch); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if err := h.survey.ChannelCreateForAdminContext(ctx, &ch, admin.ID); err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.JSON(c, ch)
}

// ChannelDel POST /admin/survey/channel_del
// @Tags PC端-问卷管理
// @Summary 删除渠道
// @Param id formData int true "渠道ID"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ChannelDel(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.survey.ChannelDeleteForAdminContext(ctx, uint(id), admin.ID); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
