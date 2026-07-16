package exam

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// QuestionBankList GET /admin/exam/question_bank_list
// @Tags PC端-考试管理
// @Summary 获取考试题库列表
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_list [get]
func (h *AdminExamHandler) QuestionBankList(_ context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	qType := c.Query("type")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	q := database.DB.Model(&model.ExamQuestion{})
	if keyword != "" {
		q = q.Where("`exam_q_title` LIKE ? OR `exam_q_type` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`exam_q_category` = ?", category)
	}
	if qType != "" {
		q = q.Where("`exam_q_type` = ?", qType)
	}
	var total int64
	q.Count(&total)
	var list []model.ExamQuestion
	q.Order("`exam_q_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.JSON(c, response.PageData{List: list, Total: total, Size: pageSize, Page: page})
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
// @Router /admin/exam/question_bank_insert [post]
func (h *AdminExamHandler) QuestionBankInsert(_ context.Context, c *app.RequestContext) {
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
	deptID := uint(0)
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			createBy = a.ID
			var adminDept model.AdminDept
			if err := database.DB.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
				deptID = adminDept.DeptID
			}
		}
	}
	q := model.ExamQuestion{
		Title:    r.Title,
		Type:     r.Type,
		Schema:   r.Schema,
		Category: r.Category,
		Tags:     r.Tags,
		Status:   1,
		DeptID:   deptID,
		CreateBy: createBy,
		AddTime:  database.Now(),
	}
	if err := database.DB.Create(&q).Error; err != nil {
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
// @Router /admin/exam/question_bank_edit [post]
func (h *AdminExamHandler) QuestionBankEdit(_ context.Context, c *app.RequestContext) {
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
	if err := database.DB.Model(&model.ExamQuestion{}).Where("`exam_q_id` = ?", r.ID).Updates(map[string]interface{}{
		"exam_q_title":    r.Title,
		"exam_q_type":     r.Type,
		"exam_q_schema":   r.Schema,
		"exam_q_category": r.Category,
		"exam_q_tags":     r.Tags,
	}).Error; err != nil {
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
// @Router /admin/exam/question_bank_del [post]
func (h *AdminExamHandler) QuestionBankDel(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := database.DB.Where("`exam_q_id` = ?", id).Delete(&model.ExamQuestion{}).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankCategories GET /admin/exam/question_bank_categories
// @Tags PC端-考试管理
// @Summary 获取考试题库所有分类
// @Success 200 {object} response.Resp
// @Router /admin/exam/question_bank_categories [get]
func (h *AdminExamHandler) QuestionBankCategories(_ context.Context, c *app.RequestContext) {
	var categories []string
	database.DB.Model(&model.ExamQuestion{}).
		Where("`exam_q_category` != '' AND `exam_q_category` IS NOT NULL").
		Select("DISTINCT `exam_q_category`").
		Order("`exam_q_category` ASC").
		Pluck("`exam_q_category`", &categories)
	response.JSON(c, categories)
}
