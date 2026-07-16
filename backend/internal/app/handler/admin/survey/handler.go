package survey

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	surveyservice "wecheckin-backend/backend/internal/app/service/survey"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/logger"
	"wecheckin-backend/backend/pkg/response"
)

type AdminSurveyHandler struct {
	survey    *surveyservice.SurveyService
	responses *surveyservice.ResponseService
}

func NewAdminSurveyHandler() *AdminSurveyHandler {
	return &AdminSurveyHandler{}
}

func (h *AdminSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = surveyservice.NewSurveyService()
	}
	if h.responses == nil {
		h.responses = surveyservice.NewResponseService()
	}
}

// List GET /admin/survey/survey_list
// @Tags PC端-问卷管理
// @Summary 问卷列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param category query string false "分类"
// @Param status query int false "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_list [get]
func (h *AdminSurveyHandler) List(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	status, _ := strconv.Atoi(c.Query("status"))
	list, total, err := h.survey.ListContext(ctx, keyword, category, status, page, pageSize)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	var ids []uint
	for _, sv := range list {
		ids = append(ids, sv.ID)
	}
	type RespCount struct {
		SurveyID uint `gorm:"column:survey_resp_survey_id"`
		Count    int  `gorm:"column:cnt"`
	}
	var counts []RespCount
	if len(ids) > 0 {
		db, cancel := database.WithContext(ctx)
		defer cancel()
		db.Model(&model.SurveyResponse{}).
			Select("`survey_resp_survey_id`, COUNT(*) AS cnt").
			Where("`survey_resp_survey_id` IN ?", ids).
			Group("`survey_resp_survey_id`").
			Scan(&counts)
	}
	countMap := make(map[uint]int)
	for _, c := range counts {
		countMap[c.SurveyID] = c.Count
	}
	var out []surveyWithCount
	for _, sv := range list {
		out = append(out, surveyWithCount{Survey: sv, ResponseCount: countMap[sv.ID]})
	}
	response.JSON(c, surveyListResponse{List: out, Total: total, Page: page, Size: pageSize})
}

// Detail GET /admin/survey/survey_detail?id=
// @Tags PC端-问卷管理
// @Summary 问卷详情
// @Param id query int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_detail [get]
func (h *AdminSurveyHandler) Detail(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.Query("id"))
	sv, err := h.survey.GetContext(ctx, uint(id))
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	var respCnt int64
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", sv.ID).Count(&respCnt)
	var rawSchema string
	db.Model(&model.Survey{}).Select("survey_schema").Where("`survey_id` = ?", id).Scan(&rawSchema)
	response.JSON(c, surveyDetailResponse{Survey: sv, ResponseCount: respCnt, Schema: rawSchema})
}

// Insert POST /admin/survey/survey_insert
// @Tags PC端-问卷管理
// @Summary 创建问卷
// @Param survey body model.Survey true "问卷数据"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_insert [post]
func (h *AdminSurveyHandler) Insert(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var sv model.Survey
	if err := c.BindAndValidate(&sv); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			sv.CreateBy = a.ID
			var adminDept model.AdminDept
			db, cancel := database.WithContext(ctx)
			defer cancel()
			if err := db.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
				sv.DeptID = adminDept.DeptID
			}
		}
	}
	if err := h.survey.CreateContext(ctx, &sv); err != nil {
		logger.Logger.Printf("[AdminSurveyInsert] 创建失败 title=%s err=%s", sv.Title, err.Error())
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	logger.Logger.Printf("[AdminSurveyInsert] 创建成功 id=%d title=%s", sv.ID, sv.Title)
	response.JSON(c, sv)
}

// Edit POST /admin/survey/survey_edit
// @Tags PC端-问卷管理
// @Summary 编辑问卷
// @Param survey body model.Survey true "问卷数据（需包含ID）"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_edit [post]
func (h *AdminSurveyHandler) Edit(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	var sv model.Survey
	if err := c.BindAndValidate(&sv); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	old := model.Survey{}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Where("`survey_id` = ?", sv.ID).First(&old)
	if err := h.survey.UpdateContext(ctx, &sv); err != nil {
		logger.Logger.Printf("[AdminSurveyEdit] 更新失败 id=%d err=%s", sv.ID, err.Error())
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	logger.Logger.Printf("[AdminSurveyEdit] 更新成功 id=%d title=%s oldMaxResp=%d newMaxResp=%d", sv.ID, sv.Title, old.MaxResponse, sv.MaxResponse)
	response.JSON(c, nil)
}

// Del POST /admin/survey/survey_del
// @Tags PC端-问卷管理
// @Summary 删除问卷
// @Param id formData int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_del [post]
func (h *AdminSurveyHandler) Del(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.survey.DeleteContext(ctx, uint(id)); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Status POST /admin/survey/survey_status
// @Tags PC端-问卷管理
// @Summary 更新问卷状态
// @Param id formData int true "问卷ID"
// @Param status formData int true "状态(0草稿 1发布)"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_status [post]
func (h *AdminSurveyHandler) Status(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Model(&model.Survey{}).Where("`survey_id` = ?", id).
		Update("survey_status", status).Error; err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}

// Copy POST /admin/survey/survey_copy
// @Tags PC端-问卷管理
// @Summary 复制问卷
// @Param id formData int true "问卷ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/survey_copy [post]
func (h *AdminSurveyHandler) Copy(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	id, _ := strconv.Atoi(c.PostForm("id"))
	sv, err := h.survey.GetContext(ctx, uint(id))
	if err != nil {
		response.Fail(c, "原问卷不存在")
		return
	}
	now := time.Now().UnixMilli()
	newSv := *sv
	newSv.ID = 0
	newSv.Title = sv.Title + " (副本)"
	newSv.Status = 0
	newSv.AddTime = now
	newSv.EditTime = now
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			newSv.CreateBy = a.ID
			var adminDept model.AdminDept
			db, cancel := database.WithContext(ctx)
			defer cancel()
			if err := db.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
				newSv.DeptID = adminDept.DeptID
			}
		}
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Create(&newSv).Error; err != nil {
		response.Fail(c, "复制失败: "+err.Error())
		return
	}
	response.JSON(c, newSv)
}
