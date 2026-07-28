package survey

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/app/service/formkitadmin"
	"wecheckin-backend/backend/pkg/response"
)

// ReportEventSchema GET /admin/survey/report/event?eventId=xx
// @Tags PC端-表单工具
// @Summary 活动报表
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/report/event [get]
func (h *AdminSurveyHandler) ReportEventSchema(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "缺少 eventId")
		return
	}
	data, err := formkitadmin.EventReportContext(ctx, eventID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	response.JSON(c, data)
}

// ExportEventSchemaCSV GET /admin/survey/export/event?eventId=xx
// @Tags PC端-表单工具
// @Summary 导出活动CSV
// @Param eventId query string true "活动ID"
// @Success 200 {file} string
// @Router /admin/survey/export/event [get]
func (h *AdminSurveyHandler) ExportEventSchemaCSV(ctx context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "缺少 eventId")
		return
	}
	csvBytes, filename, err := formkitadmin.EventReportCSVContext(ctx, eventID)
	if err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	writeCSV(c, filename, csvBytes)
}
