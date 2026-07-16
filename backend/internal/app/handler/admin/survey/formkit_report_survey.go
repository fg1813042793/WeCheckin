package survey

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// ReportSurveySchema GET /admin/survey/report/survey?surveyId=xx
// @Tags PC端-表单工具
// @Summary 问卷报表（schema-aware）
// @Param surveyId query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/report/survey [get]
func (h *AdminSurveyHandler) ReportSurveySchema(_ context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	if surveyID == 0 {
		response.Fail(c, "缺少 surveyId")
		return
	}
	h.lazyInit()
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	var respList []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ?", surveyID).
		Order("`survey_resp_id` DESC").Find(&respList)
	items := make([]report.AnswerItem, 0, len(respList))
	for _, r := range respList {
		items = append(items, report.AnswerItem{
			UserID:  r.UserID,
			AddTime: time.UnixMilli(r.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   r.Answers,
		})
	}
	statMode := "count"
	if sv.Settings != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(sv.Settings), &settings); err == nil {
			if rc, ok := settings["resultConfig"].(map[string]interface{}); ok {
				if st, ok := rc["statType"].(string); ok && st != "" {
					statMode = st
				}
			}
		}
	}
	table, _ := report.RenderAnswers(sv.Schema, items)
	stats := report.FieldStats(sv.Schema, items, statMode)
	response.JSON(c, map[string]interface{}{
		"schema": sv.Schema,
		"table":  table,
		"stats":  stats,
		"count":  len(respList),
		"title":  sv.Title,
	})
}

// ExportSurveySchemaCSV GET /admin/survey/export/survey?surveyId=xx
// @Tags PC端-表单工具
// @Summary 导出问卷CSV
// @Param surveyId query int true "问卷ID"
// @Success 200 {file} string
// @Router /admin/survey/export/survey [get]
func (h *AdminSurveyHandler) ExportSurveySchemaCSV(_ context.Context, c *app.RequestContext) {
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	if surveyID == 0 {
		response.Fail(c, "缺少 surveyId")
		return
	}
	h.lazyInit()
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	var respList []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ?", surveyID).
		Order("`survey_resp_id` DESC").Find(&respList)
	items := make([]report.AnswerItem, 0, len(respList))
	for _, r := range respList {
		items = append(items, report.AnswerItem{
			UserID:  r.UserID,
			AddTime: time.UnixMilli(r.AddTime).Format("2006-01-02 15:04:05"),
			Forms:   r.Answers,
		})
	}
	table, _ := report.RenderAnswers(sv.Schema, items)
	csvBytes := report.ToCSV(table)
	filename := fmt.Sprintf("survey_%s_%d.csv", report.SanitizeFilename(sv.Title), time.Now().Unix())
	writeCSV(c, filename, csvBytes)
}
