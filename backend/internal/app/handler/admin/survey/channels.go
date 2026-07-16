package survey

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// ChannelList GET /admin/survey/channel_list?surveyId=
// @Tags PC端-问卷管理
// @Summary 渠道列表
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/channel_list [get]
func (h *AdminSurveyHandler) ChannelList(ctx context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	var list []model.SurveyChannel
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Where("`survey_ch_survey_id` = ?", surveyID).Order("`survey_ch_id` DESC").Find(&list)
	response.JSON(c, surveyChannelListResponse{List: list})
}

// ChannelInsert POST /admin/survey/channel_insert
// @Tags PC端-问卷管理
// @Summary 创建渠道
// @Param channel body model.SurveyChannel true "渠道数据"
// @Success 200 {object} response.Resp
// @Router /admin/survey/channel_insert [post]
func (h *AdminSurveyHandler) ChannelInsert(ctx context.Context, c *app.RequestContext) {
	var ch model.SurveyChannel
	if err := c.BindAndValidate(&ch); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	ch.AddTime = time.Now().UnixMilli()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Create(&ch).Error; err != nil {
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
// @Router /admin/survey/channel_del [post]
func (h *AdminSurveyHandler) ChannelDel(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Where("`survey_ch_id` = ?", id).Delete(&model.SurveyChannel{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
