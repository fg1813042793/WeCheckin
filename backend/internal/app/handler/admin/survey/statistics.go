package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/pkg/response"
)

// Statistic GET /admin/survey/statistic?surveyId=
// @Tags PC端-问卷管理
// @Summary 问卷统计
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/statistic [get]
func (h *AdminSurveyHandler) Statistic(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	viewURL := "/admin/survey/response_list?surveyId=" + c.Query("surveyId")
	data, err := h.survey.StatisticContext(ctx, uint(surveyID), viewURL)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	response.JSON(c, data)
}
