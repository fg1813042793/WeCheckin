package survey

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// ReportEventSchema GET /admin/survey/report/event?eventId=xx
// @Tags PC端-表单工具
// @Summary 活动报表
// @Param eventId query string true "活动ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/report/event [get]
func (h *AdminSurveyHandler) ReportEventSchema(_ context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "缺少 eventId")
		return
	}
	var event model.Event
	if err := database.DB.Where("`id` = ?", eventID).First(&event).Error; err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	var parts []model.EventParticipant
	database.DB.Where("`event_part_event_id` = ?", eventID).
		Order("`event_part_add_time` DESC").Find(&parts)
	items := make([]report.AnswerItem, 0, len(parts))
	for _, p := range parts {
		items = append(items, report.AnswerItem{
			UserID:  p.MiniOpenID,
			AddTime: time.UnixMilli(p.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   p.Forms,
		})
	}
	table, _ := report.RenderAnswers(event.Forms, items)
	stats := report.FieldStats(event.Forms, items, "count")
	response.JSON(c, map[string]interface{}{
		"schema": event.Forms,
		"table":  table,
		"stats":  stats,
		"count":  len(parts),
		"title":  event.Title,
	})
}

// ExportEventSchemaCSV GET /admin/survey/export/event?eventId=xx
// @Tags PC端-表单工具
// @Summary 导出活动CSV
// @Param eventId query string true "活动ID"
// @Success 200 {file} string
// @Router /admin/survey/export/event [get]
func (h *AdminSurveyHandler) ExportEventSchemaCSV(_ context.Context, c *app.RequestContext) {
	eventID := c.Query("eventId")
	if eventID == "" {
		response.Fail(c, "缺少 eventId")
		return
	}
	var event model.Event
	if err := database.DB.Where("`id` = ?", eventID).First(&event).Error; err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	var parts []model.EventParticipant
	database.DB.Where("`event_part_event_id` = ?", eventID).
		Order("`event_part_add_time` DESC").Find(&parts)
	items := make([]report.AnswerItem, 0, len(parts))
	for _, p := range parts {
		items = append(items, report.AnswerItem{
			UserID:  p.MiniOpenID,
			AddTime: time.UnixMilli(p.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   p.Forms,
		})
	}
	table, _ := report.RenderAnswers(event.Forms, items)
	csvBytes := report.ToCSV(table)
	filename := fmt.Sprintf("event_%s_%d.csv", report.SanitizeFilename(event.Title), time.Now().Unix())
	writeCSV(c, filename, csvBytes)
}
