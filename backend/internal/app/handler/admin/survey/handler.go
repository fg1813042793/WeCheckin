package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin-backend/backend/internal/app/service/formkitadmin"
	surveyservice "wecheckin-backend/backend/internal/app/service/survey"
	"wecheckin-backend/backend/internal/model"
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
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	status, _ := strconv.Atoi(c.Query("status"))
	list, total, err := h.survey.ListForAdminContext(ctx, keyword, category, status, page, pageSize, admin.ID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	var ids []uint
	for _, sv := range list {
		ids = append(ids, sv.ID)
	}
	countMap, err := h.survey.ResponseCountsContext(ctx, ids)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	var out []surveyListItem
	for _, sv := range list {
		out = append(out, newSurveyListItem(sv, countMap[sv.ID]))
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
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.Query("id"))
	detail, err := h.survey.DetailForAdminContext(ctx, uint(id), admin.ID)
	if err != nil {
		response.Fail(c, "问卷不存在")
		return
	}
	response.JSON(c, surveyDetailResponse{Survey: detail.Survey, ResponseCount: detail.ResponseCount, Schema: detail.Schema})
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
			deptID, err := formkitadmin.FirstAdminDeptIDContext(ctx, a.ID)
			if err != nil {
				response.Fail(c, "获取部门失败: "+err.Error())
				return
			}
			sv.DeptID = deptID
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
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	var sv model.Survey
	if err := c.BindAndValidate(&sv); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	old, _ := h.survey.GetForAdminContext(ctx, sv.ID, admin.ID)
	if err := h.survey.UpdateForAdminContext(ctx, &sv, admin.ID); err != nil {
		logger.Logger.Printf("[AdminSurveyEdit] 更新失败 id=%d err=%s", sv.ID, err.Error())
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	oldMaxResponse := 0
	if old != nil {
		oldMaxResponse = old.MaxResponse
	}
	logger.Logger.Printf("[AdminSurveyEdit] 更新成功 id=%d title=%s oldMaxResp=%d newMaxResp=%d", sv.ID, sv.Title, oldMaxResponse, sv.MaxResponse)
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
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.survey.DeleteForAdminContext(ctx, uint(id), admin.ID); err != nil {
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
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(c.PostForm("id"))
	status, _ := strconv.Atoi(c.PostForm("status"))
	if err := h.survey.SetStatusForAdminContext(ctx, uint(id), status, admin.ID); err != nil {
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
	var createBy uint
	var deptID uint
	var adminID uint
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			createBy = a.ID
			adminID = a.ID
			var err error
			deptID, err = formkitadmin.FirstAdminDeptIDContext(ctx, a.ID)
			if err != nil {
				response.Fail(c, "获取部门失败: "+err.Error())
				return
			}
		}
	}
	newSv, err := h.survey.CopyForAdminContext(ctx, uint(id), createBy, deptID, adminID)
	if err != nil {
		response.Fail(c, "复制失败: "+err.Error())
		return
	}
	response.JSON(c, newSv)
}
