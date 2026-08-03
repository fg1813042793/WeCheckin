package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/app/service/formkitadmin"
	"wecheckin/backend/pkg/response"
)

// ReportEnrollSchema GET /admin/survey/report/enroll?enrollId=xx
// @Tags PC端-表单工具
// @Summary 打卡报表
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {object} response.Resp
func (h *AdminSurveyHandler) ReportEnrollSchema(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	if enrollID == "" {
		response.Fail(c, "缺少 enrollId")
		return
	}
	data, err := formkitadmin.EnrollReportContext(ctx, enrollID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	response.JSON(c, data)
}

// ExportEnrollSchemaCSV GET /admin/survey/export/enroll?enrollId=xx
// @Tags PC端-表单工具
// @Summary 导出打卡CSV
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {file} string
func (h *AdminSurveyHandler) ExportEnrollSchemaCSV(ctx context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	if enrollID == "" {
		response.Fail(c, "缺少 enrollId")
		return
	}
	csvBytes, filename, err := formkitadmin.EnrollReportCSVContext(ctx, enrollID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	writeCSV(c, filename, csvBytes)
}
