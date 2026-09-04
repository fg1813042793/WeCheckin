package exam

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	formkitadminservice "wecheckin/backend/internal/service/admin/formkitadmin"
	"wecheckin/backend/pkg/response"
)

// QuestionBankList GET /admin/exam/question_bank_list
// @Tags PC端-考试管理
// @Summary 获取考试题库列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) QuestionBankList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	query := formkitadminservice.NormalizeQuestionBankQuery(formkitadminservice.QuestionBankQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		Category: c.Query("category"),
		Type:     c.Query("type"),
	})
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	list, total, err := formkitadminservice.ListExamQuestionsForAdminContext(ctx, query, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, response.PageData{List: list, Total: total, Size: query.PageSize, Page: query.Page})
}

// QuestionBankInsert POST /admin/exam/question_bank_insert
// @Tags PC端-考试管理
// @Summary 添加题目到考试题库
// @Param title body string true "题干"
// @Param type body string true "题型"
// @Param schema body string true "完整 formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) QuestionBankInsert(ctx context.Context, c *app.RequestContext) {
	type req struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		Schema   string `json:"schema"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	var r req
	if err := c.BindAndValidate(&r); err != nil {
		response.FailInternal(ctx, c, "admin.exam.question_bank", "参数错误，请稍后重试", err)
		return
	}
	if r.Title == "" {
		response.Fail(c, "标题不能为空")
		return
	}
	createBy := uint(0)
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			createBy = a.ID
		}
	}
	q, err := formkitadminservice.CreateExamQuestionContext(ctx, formkitadminservice.QuestionBankInput{
		Title:    r.Title,
		Type:     r.Type,
		Schema:   r.Schema,
		Category: r.Category,
		Tags:     r.Tags,
		AdminID:  createBy,
	})
	if err != nil {
		response.Fail(c, "创建失败")
		return
	}
	response.JSON(c, q)
}

// QuestionBankEdit POST /admin/exam/question_bank_edit
// @Tags PC端-考试管理
// @Summary 编辑题库题目
// @Param id body uint true "题目ID"
// @Param title body string false "题干"
// @Param type body string false "题型"
// @Param schema body string false "formkit JSON"
// @Param category body string false "分类"
// @Param tags body string false "标签"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) QuestionBankEdit(ctx context.Context, c *app.RequestContext) {
	type req struct {
		ID       uint   `json:"id"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Schema   string `json:"schema"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	var r req
	if err := c.BindAndValidate(&r); err != nil {
		response.FailInternal(ctx, c, "admin.exam.question_bank", "参数错误，请稍后重试", err)
		return
	}
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	if err := formkitadminservice.UpdateExamQuestionForAdminContext(ctx, formkitadminservice.QuestionBankInput{
		ID:       r.ID,
		Title:    r.Title,
		Type:     r.Type,
		Schema:   r.Schema,
		Category: r.Category,
		Tags:     r.Tags,
	}, admin.ID); err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankDel POST /admin/exam/question_bank_del
// @Tags PC端-考试管理
// @Summary 删除题库题目
// @Param id formData int true "题目ID"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) QuestionBankDel(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	if err := formkitadminservice.DeleteExamQuestionForAdminContext(ctx, uint(id), admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankCategories GET /admin/exam/question_bank_categories
// @Tags PC端-考试管理
// @Summary 获取考试题库所有分类
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) QuestionBankCategories(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	categories, err := formkitadminservice.ExamQuestionCategoriesForAdminContext(ctx, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, categories)
}
