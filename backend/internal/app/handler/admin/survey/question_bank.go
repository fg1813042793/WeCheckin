package survey

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// QuestionBankList GET /admin/survey/question_bank_list
func (h *AdminSurveyHandler) QuestionBankList(ctx context.Context, c *app.RequestContext) {
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
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.SurveyQuestion{})
	if keyword != "" {
		q = q.Where("`survey_q_title` LIKE ? OR `survey_q_type` LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`survey_q_category` = ?", category)
	}
	if qType != "" {
		q = q.Where("`survey_q_type` = ?", qType)
	}
	var total int64
	q.Count(&total)
	var list []model.SurveyQuestion
	q.Order("`survey_q_add_time` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
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
	deptID := uint(0)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if admin, ok := c.Get("admin"); ok {
		if a, ok := admin.(*model.Admin); ok {
			createBy = a.ID
			var adminDept model.AdminDept
			if err := db.Where("admin_dept_admin_id = ?", a.ID).First(&adminDept).Error; err == nil {
				deptID = adminDept.DeptID
			}
		}
	}
	q := model.SurveyQuestion{
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
	if err := db.Create(&q).Error; err != nil {
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
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Model(&model.SurveyQuestion{}).Where("`survey_q_id` = ?", r.ID).Updates(map[string]interface{}{
		"survey_q_title":    r.Title,
		"survey_q_type":     r.Type,
		"survey_q_schema":   r.Schema,
		"survey_q_category": r.Category,
		"survey_q_tags":     r.Tags,
	}).Error; err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankDel POST /admin/survey/question_bank_del
func (h *AdminSurveyHandler) QuestionBankDel(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Where("`survey_q_id` = ?", id).Delete(&model.SurveyQuestion{}).Error; err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// QuestionBankCategories GET /admin/survey/question_bank_categories
// 返回题库中所有已有的分类列表，用于前端下拉选择
func (h *AdminSurveyHandler) QuestionBankCategories(ctx context.Context, c *app.RequestContext) {
	var categories []string
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Model(&model.SurveyQuestion{}).
		Where("`survey_q_category` != '' AND `survey_q_category` IS NOT NULL").
		Select("DISTINCT `survey_q_category`").
		Order("`survey_q_category` ASC").
		Pluck("`survey_q_category`", &categories)
	response.JSON(c, categories)
}
