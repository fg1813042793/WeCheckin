package survey

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	_ "wecheckin-backend/backend/internal/app/formkit/question/builtin" // 注册 24 个内置题型
	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// ResponseList GET /admin/survey/response_list?surveyId=
// @Tags PC端-问卷管理
// @Summary 答卷列表
// @Param surveyId query int true "问卷ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_list [get]
func (h *AdminSurveyHandler) ResponseList(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	list, total, err := h.responses.List(uint(surveyID), page, pageSize, keyword)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	voList := make([]surveyResponseWithAnswers, len(list))
	for i, r := range list {
		voList[i] = surveyResponseWithAnswers{SurveyResponse: r, AnswersMap: h.responses.ParseAnswers(&r)}
	}
	response.JSON(c, surveyResponseListResponse{List: voList, Total: total, Page: page, Size: pageSize})
}

// ResponseDetail GET /admin/survey/response_detail?id=
// @Tags PC端-问卷管理
// @Summary 答卷详情
// @Param id query int true "答卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_detail [get]
func (h *AdminSurveyHandler) ResponseDetail(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	resp, err := h.responses.Get(uint(id))
	if err != nil {
		response.Fail(c, "答卷不存在")
		return
	}
	sv, _ := h.survey.Get(resp.SurveyID)
	answers := h.responses.ParseAnswers(resp)
	var schMap map[string]interface{}
	if sv != nil {
		_ = json.Unmarshal([]byte(sv.Schema), &schMap)
	}
	response.JSON(c, surveyResponseDetailResponse{Response: resp, Survey: sv, Answers: answers, Schema: schMap})
}

// ResponseDel POST /admin/survey/response_del
// @Tags PC端-问卷管理
// @Summary 删除答卷
// @Param id formData int true "答卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_del [post]
func (h *AdminSurveyHandler) ResponseDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`survey_resp_id` = ?", id).Delete(&model.SurveyResponse{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ResponseBatchDel POST /admin/survey/response_batch_del
// @Tags PC端-问卷管理
// @Summary 批量删除答卷
// @Param ids formData string true "逗号分隔的答卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/response_batch_del [post]
func (h *AdminSurveyHandler) ResponseBatchDel(_ context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	parts := strings.Split(idsStr, ",")
	var ids []int
	for _, p := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	if err := database.DB.Where("`survey_resp_id` IN ?", ids).Delete(&model.SurveyResponse{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// ResponseExport GET /admin/survey/response_export?surveyId=
// @Tags PC端-问卷管理
// @Summary 导出答卷CSV
// @Param surveyId query int true "问卷ID"
// @Success 200 {file} string
// @Router /admin/survey/response_export [get]
func (h *AdminSurveyHandler) ResponseExport(_ context.Context, c *app.RequestContext) {
	h.lazyInit()
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	sv, err := h.survey.Get(uint(surveyID))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	var list []model.SurveyResponse
	database.DB.Where("`survey_resp_survey_id` = ?", surveyID).Order("`survey_resp_id` ASC").Find(&list)
	items := make([]report.AnswerItem, len(list))
	for i, r := range list {
		items[i] = report.AnswerItem{UserID: r.UserID, AddTime: time.UnixMilli(r.AddTime).Format("2006-01-02 15:04:05"), Forms: r.Answers}
	}
	tbl, err := report.RenderAnswers(sv.Schema, items)
	if err != nil {
		response.Fail(c, "导出失败: "+err.Error())
		return
	}
	csvData := report.ToCSV(tbl)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=survey_"+strconv.Itoa(surveyID)+".csv")
	c.Write(csvData)
}
