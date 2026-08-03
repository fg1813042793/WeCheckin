package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	formkitadminservice "wecheckin/backend/internal/app/service/formkitadmin"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// QuestionBankList GET /admin/survey/question_bank_list
func (h *AdminSurveyHandler) QuestionBankList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	list, total, err := formkitadminservice.ListSurveyQuestionsForAdminContext(ctx, formkitadminservice.QuestionBankQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		Category: c.Query("category"),
		Type:     c.Query("type"),
	}, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, surveyQuestionListResponse{List: list, Total: total})
}

// QuestionBankInsert POST /admin/survey/question_bank_insert
func (h *AdminSurveyHandler) QuestionBankInsert(ctx context.Context, c *app.RequestContext) {
	type req struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		Schema   string `json:"schema"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	var r req
	if err := c.BindAndValidate(&r); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
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
	q, err := formkitadminservice.CreateSurveyQuestionContext(ctx, formkitadminservice.QuestionBankInput{
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

// QuestionBankEdit POST /admin/survey/question_bank_edit
func (h *AdminSurveyHandler) QuestionBankEdit(ctx context.Context, c *app.RequestContext) {
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
		response.Fail(c, "参数错误: "+err.Error())
		return
	}
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	if err := formkitadminservice.UpdateSurveyQuestionForAdminContext(ctx, formkitadminservice.QuestionBankInput{
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

// QuestionBankDel POST /admin/survey/question_bank_del
func (h *AdminSurveyHandler) QuestionBankDel(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	if err := formkitadminservice.DeleteSurveyQuestionForAdminContext(ctx, uint(id), admin.ID); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankCategories GET /admin/survey/question_bank_categories
// 返回题库中所有已有的分类列表，用于前端下拉选择
func (h *AdminSurveyHandler) QuestionBankCategories(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	categories, err := formkitadminservice.SurveyQuestionCategoriesForAdminContext(ctx, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, categories)
}
