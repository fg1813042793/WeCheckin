package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/app/service/formkitadmin"
	"wecheckin/backend/pkg/response"
)

// ReportSurveySchema GET /admin/survey/report/survey?surveyId=xx
// @Tags PC端-表单工具
// @Summary 问卷报表（schema-aware）
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ReportSurveySchema(ctx context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	if surveyID == 0 {
		response.Fail(c, "缺少 surveyId")
		return
	}
	data, err := formkitadmin.SurveyReportContext(ctx, uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	response.JSON(c, data)
}

// ExportSurveySchemaCSV GET /admin/survey/export/survey?surveyId=xx
// @Tags PC端-表单工具
// @Summary 导出问卷CSV
// @Param surveyId query int true "问卷ID"
// @Success 200 {file} string
func (h *AdminSurveyHandler) ExportSurveySchemaCSV(ctx context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	if surveyID == 0 {
		response.Fail(c, "缺少 surveyId")
		return
	}
	csvBytes, filename, err := formkitadmin.SurveyReportCSVContext(ctx, uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	writeCSV(c, filename, csvBytes)
}
