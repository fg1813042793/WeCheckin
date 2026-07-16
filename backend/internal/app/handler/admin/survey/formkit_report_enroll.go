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

// ReportEnrollSchema GET /admin/survey/report/enroll?enrollId=xx
// @Tags PC端-表单工具
// @Summary 打卡报表
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/report/enroll [get]
func (h *AdminSurveyHandler) ReportEnrollSchema(_ context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	if enrollID == "" {
		response.Fail(c, "缺少 enrollId")
		return
	}
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	var joins []model.EnrollJoin
	database.DB.Where("`enroll_join_enroll_id` = ?", enrollID).
		Order("`enroll_join_add_time` DESC").Find(&joins)

	items := make([]report.AnswerItem, 0, len(joins))
	for _, j := range joins {
		items = append(items, report.AnswerItem{
			UserID:  j.UserID,
			AddTime: time.UnixMilli(j.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   j.Forms,
		})
	}
	table, _ := report.RenderAnswers(enroll.Forms, items)
	stats := report.FieldStats(enroll.Forms, items, "count")
	response.JSON(c, map[string]interface{}{
		"schema": enroll.Forms,
		"table":  table,
		"stats":  stats,
		"count":  len(joins),
		"title":  enroll.Title,
	})
}

// ExportEnrollSchemaCSV GET /admin/survey/export/enroll?enrollId=xx
// @Tags PC端-表单工具
// @Summary 导出打卡CSV
// @Param enrollId query string true "打卡项目ID"
// @Success 200 {file} string
// @Router /admin/survey/export/enroll [get]
func (h *AdminSurveyHandler) ExportEnrollSchemaCSV(_ context.Context, c *app.RequestContext) {
	enrollID := c.Query("enrollId")
	if enrollID == "" {
		response.Fail(c, "缺少 enrollId")
		return
	}
	var enroll model.Enroll
	if err := database.DB.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		response.Fail(c, "项目不存在")
		return
	}
	var joins []model.EnrollJoin
	database.DB.Where("`enroll_join_enroll_id` = ?", enrollID).
		Order("`enroll_join_add_time` DESC").Find(&joins)
	items := make([]report.AnswerItem, 0, len(joins))
	for _, j := range joins {
		items = append(items, report.AnswerItem{
			UserID:  j.UserID,
			AddTime: time.UnixMilli(j.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   j.Forms,
		})
	}
	table, _ := report.RenderAnswers(enroll.Forms, items)
	csvBytes := report.ToCSV(table)
	filename := fmt.Sprintf("enroll_%s_%d.csv", report.SanitizeFilename(enroll.Title), time.Now().Unix())
	writeCSV(c, filename, csvBytes)
}
