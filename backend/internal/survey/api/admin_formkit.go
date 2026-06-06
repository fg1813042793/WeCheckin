package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/database"
	calcPkg "wecheckin-backend/backend/internal/formkit/calc"
	questionPkg "wecheckin-backend/backend/internal/formkit/question"
	"wecheckin-backend/backend/internal/formkit/report"
	schemaPkg "wecheckin-backend/backend/internal/formkit/schema"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// ==================== 题型元信息 / Schema 解析 / 表达式试算 ====================

// TypeMeta 题型元信息
type TypeMeta struct {
	Type         string                 `json:"type"`
	DisplayName  string                 `json:"displayName"`
	Category     string                 `json:"category"`
	DefaultProps map[string]interface{} `json:"defaultProps"`
}

// ListTypes GET /admin/survey/types
func (h *AdminSurveyHandler) ListTypes(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	all := questionPkg.All()
	out := make([]TypeMeta, 0, len(all))
	for _, q := range all {
		out = append(out, TypeMeta{
			Type:         q.Type(),
			DisplayName:  q.DisplayName(),
			Category:     q.Category(),
			DefaultProps: q.DefaultProps(),
		})
	}
	response.JSON(c, out)
}

// ParseSchema POST /admin/survey/schema/parse
func (h *AdminSurveyHandler) ParseSchema(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema string `json:"schema"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	if req.Schema == "" {
		response.Fail(c, "schema 不能为空")
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	response.JSON(c, s)
}

// EvalExpr POST /admin/survey/eval
func (h *AdminSurveyHandler) EvalExpr(_ context.Context, c *app.RequestContext) {
	var req struct {
		Expr   string                 `json:"expr"`
		Env    map[string]interface{} `json:"env"`
		AsBool bool                   `json:"asBool"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	e := calcPkg.New()
	if req.AsBool {
		v, err := e.EvalBool(req.Expr, req.Env)
		if err != nil {
			response.Fail(c, "eval 失败: "+err.Error())
			return
		}
		response.JSON(c, map[string]interface{}{"value": v})
		return
	}
	v, err := e.Eval(req.Expr, req.Env)
	if err != nil {
		response.Fail(c, "eval 失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]interface{}{"value": v})
}

// ValidateAnswers POST /admin/survey/validate
func (h *AdminSurveyHandler) ValidateAnswers(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema  string                 `json:"schema"`
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	type fieldErr struct {
		QuestionID string `json:"questionId"`
		Message    string `json:"message"`
	}
	var errs []fieldErr
	for _, q := range s.Questions {
		v, ok := req.Answers[q.ID]
		if !ok && q.Required {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: "此项为必填"})
			continue
		}
		inst := questionPkg.Get(q.Type)
		if inst == nil {
			continue
		}
		if err := inst.Validate(v, q); err != nil {
			errs = append(errs, fieldErr{QuestionID: q.ID, Message: err.Error()})
		}
	}
	response.JSON(c, map[string]interface{}{"ok": len(errs) == 0, "errors": errs})
}

// ApplyForm POST /admin/survey/apply
func (h *AdminSurveyHandler) ApplyForm(_ context.Context, c *app.RequestContext) {
	var req struct {
		Schema  string                 `json:"schema"`
		Answers map[string]interface{} `json:"answers"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "请求参数错误: "+err.Error())
		return
	}
	s, err := schemaPkg.Parse(req.Schema)
	if err != nil {
		response.Fail(c, "schema 解析失败: "+err.Error())
		return
	}
	eng := calcPkg.New()
	newAns, _ := eng.ApplyCalcValues(s, req.Answers)
	states, _ := eng.ApplyLogic(s, newAns)
	response.JSON(c, map[string]interface{}{"answers": newAns, "states": states})
}

// ==================== Schema-aware 报表 (P6) ====================

// ReportEnrollSchema GET /admin/survey/report/enroll?enrollId=xx
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
	stats := report.FieldStats(enroll.Forms, items)
	response.JSON(c, map[string]interface{}{
		"schema": enroll.Forms,
		"table":  table,
		"stats":  stats,
		"count":  len(joins),
		"title":  enroll.Title,
	})
}

// ExportEnrollSchemaCSV GET /admin/survey/export/enroll?enrollId=xx
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

// ReportEventSchema GET /admin/survey/report/event?eventId=xx
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
	stats := report.FieldStats(event.Forms, items)
	response.JSON(c, map[string]interface{}{
		"schema": event.Forms,
		"table":  table,
		"stats":  stats,
		"count":  len(parts),
		"title":  event.Title,
	})
}

// ExportEventSchemaCSV GET /admin/survey/export/event?eventId=xx
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

// ReportSurveySchema GET /admin/survey/report/survey?surveyId=xx
// SurveyKing 风格的 schema-aware 报表（针对 survey 业务答卷）
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
	table, _ := report.RenderAnswers(sv.Schema, items)
	stats := report.FieldStats(sv.Schema, items)
	response.JSON(c, map[string]interface{}{
		"schema": sv.Schema,
		"table":  table,
		"stats":  stats,
		"count":  len(respList),
		"title":  sv.Title,
	})
}

// ExportSurveySchemaCSV GET /admin/survey/export/survey?surveyId=xx
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

// writeCSV 写 CSV 文件（带 UTF-8 BOM 供 Excel 识别）
func writeCSV(c *app.RequestContext, filename string, data []byte) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)+3))
	_, _ = c.Response.BodyWriter().Write([]byte{0xEF, 0xBB, 0xBF})
	_, _ = c.Response.BodyWriter().Write(data)
}
