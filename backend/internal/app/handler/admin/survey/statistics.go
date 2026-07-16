package survey

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
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
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var total int64
	db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID).Count(&total)
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	var todayCnt int64
	db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ?", surveyID, dayStart).Count(&todayCnt)
	daily := make([]surveyDailyCount, 7)
	for i := 0; i < 7; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		next := day.Add(24 * time.Hour)
		var c int64
		db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_add_time` >= ? AND `survey_resp_add_time` < ?", surveyID, day.UnixMilli(), next.UnixMilli()).Count(&c)
		daily[6-i] = surveyDailyCount{Date: day.Format("01-02"), Count: c}
	}
	var mobileCnt, pcCnt int64
	db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_device` LIKE ?", surveyID, "%Mobile%").Count(&mobileCnt)
	db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ? AND `survey_resp_device` != '' AND `survey_resp_device` NOT LIKE ?", surveyID, "%Mobile%").Count(&pcCnt)

	var allResp []model.SurveyResponse
	db.Where("`survey_resp_survey_id` = ?", surveyID).Find(&allResp)
	items := make([]report.AnswerItem, len(allResp))
	for i, r := range allResp {
		items[i] = report.AnswerItem{Forms: r.Answers}
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
	fieldStats := report.FieldStats(sv.Schema, items, statMode)

	response.JSON(c, surveyStatisticResponse{
		Survey:     sv,
		Total:      total,
		TodayCount: todayCnt,
		Daily:      daily,
		DeviceStat: surveyDeviceStat{Mobile: mobileCnt, PC: pcCnt},
		FieldStats: fieldStats,
		ViewURL:    "/admin/survey/response_list?surveyId=" + c.Query("surveyId"),
	})
}
