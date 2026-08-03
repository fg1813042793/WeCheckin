package formkitadmin

import (
	"context"

	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type QuestionBankQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Category string
	Type     string
}

type QuestionBankInput struct {
	ID       uint
	Title    string
	Type     string
	Schema   string
	Category string
	Tags     string
	AdminID  uint
}

func NormalizeQuestionBankQuery(q QuestionBankQuery) QuestionBankQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 50
	}
	return q
}

func ListSurveyQuestionsContext(ctx context.Context, input QuestionBankQuery) ([]model.SurveyQuestion, int64, error) {
	input = NormalizeQuestionBankQuery(input)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Model(&model.SurveyQuestion{})
	if input.Keyword != "" {
		query = query.Where("`survey_q_title` LIKE ? OR `survey_q_type` LIKE ?", "%"+input.Keyword+"%", "%"+input.Keyword+"%")
	}
	if input.Category != "" {
		query = query.Where("`survey_q_category` = ?", input.Category)
	}
	if input.Type != "" {
		query = query.Where("`survey_q_type` = ?", input.Type)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SurveyQuestion
	if err := query.Order("`add_time` DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func ListSurveyQuestionsForAdminContext(ctx context.Context, input QuestionBankQuery, adminID uint) ([]model.SurveyQuestion, int64, error) {
	input = NormalizeQuestionBankQuery(input)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.SurveyQuestion{}, access.SurveyQuestionAuditFields)
	if err != nil {
		return nil, 0, err
	}
	if input.Keyword != "" {
		query = query.Where("`survey_q_title` LIKE ? OR `survey_q_type` LIKE ?", "%"+input.Keyword+"%", "%"+input.Keyword+"%")
	}
	if input.Category != "" {
		query = query.Where("`survey_q_category` = ?", input.Category)
	}
	if input.Type != "" {
		query = query.Where("`survey_q_type` = ?", input.Type)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SurveyQuestion
	if err := query.Order("`add_time` DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func CreateSurveyQuestionContext(ctx context.Context, input QuestionBankInput) (model.SurveyQuestion, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	deptID, err := firstAdminDeptID(db, input.AdminID)
	if err != nil {
		return model.SurveyQuestion{}, err
	}
	now := database.Now()
	q := model.SurveyQuestion{
		Title:        input.Title,
		Type:         input.Type,
		Schema:       input.Schema,
		Category:     input.Category,
		Tags:         input.Tags,
		Status:       1,
		DeptID:       deptID,
		CreateBy:     input.AdminID,
		UpdateBy:     input.AdminID,
		UpdateDeptID: deptID,
		AddTime:      now,
		EditTime:     now,
	}
	if err := db.Create(&q).Error; err != nil {
		return model.SurveyQuestion{}, err
	}
	return q, nil
}

func UpdateSurveyQuestionContext(ctx context.Context, input QuestionBankInput) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.SurveyQuestion{}).Where("`survey_q_id` = ?", input.ID).Updates(map[string]interface{}{
		"survey_q_title":    input.Title,
		"survey_q_type":     input.Type,
		"survey_q_schema":   input.Schema,
		"survey_q_category": input.Category,
		"survey_q_tags":     input.Tags,
		"edit_time":         database.Now(),
	}).Error
}

func UpdateSurveyQuestionForAdminContext(ctx context.Context, input QuestionBankInput, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.SurveyQuestion{}, access.SurveyQuestionAuditFields)
	if err != nil {
		return err
	}
	deptID, err := firstAdminDeptID(db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(query.Where("`survey_q_id` = ?", input.ID).Updates(map[string]interface{}{
		"survey_q_title":    input.Title,
		"survey_q_type":     input.Type,
		"survey_q_schema":   input.Schema,
		"survey_q_category": input.Category,
		"survey_q_tags":     input.Tags,
		"update_by":         adminID,
		"update_dept_id":    deptID,
		"edit_time":         database.Now(),
	}))
}

func DeleteSurveyQuestionContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`survey_q_id` = ?", id).Delete(&model.SurveyQuestion{}).Error
}

func DeleteSurveyQuestionForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.SurveyQuestion{}, access.SurveyQuestionAuditFields)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(query.Where("`survey_q_id` = ?", id).Delete(&model.SurveyQuestion{}))
}

func SurveyQuestionCategoriesContext(ctx context.Context) ([]string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var categories []string
	err := db.Model(&model.SurveyQuestion{}).
		Where("`survey_q_category` != '' AND `survey_q_category` IS NOT NULL").
		Select("DISTINCT `survey_q_category`").
		Order("`survey_q_category` ASC").
		Pluck("`survey_q_category`", &categories).Error
	return categories, err
}

func SurveyQuestionCategoriesForAdminContext(ctx context.Context, adminID uint) ([]string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.SurveyQuestion{}, access.SurveyQuestionAuditFields)
	if err != nil {
		return nil, err
	}
	var categories []string
	err = query.Where("`survey_q_category` != '' AND `survey_q_category` IS NOT NULL").
		Select("DISTINCT `survey_q_category`").
		Order("`survey_q_category` ASC").
		Pluck("`survey_q_category`", &categories).Error
	return categories, err
}

func ListExamQuestionsContext(ctx context.Context, input QuestionBankQuery) ([]model.ExamQuestion, int64, error) {
	input = NormalizeQuestionBankQuery(input)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Model(&model.ExamQuestion{})
	if input.Keyword != "" {
		query = query.Where("`exam_q_title` LIKE ? OR `exam_q_type` LIKE ?", "%"+input.Keyword+"%", "%"+input.Keyword+"%")
	}
	if input.Category != "" {
		query = query.Where("`exam_q_category` = ?", input.Category)
	}
	if input.Type != "" {
		query = query.Where("`exam_q_type` = ?", input.Type)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ExamQuestion
	if err := query.Order("`add_time` DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func ListExamQuestionsForAdminContext(ctx context.Context, input QuestionBankQuery, adminID uint) ([]model.ExamQuestion, int64, error) {
	input = NormalizeQuestionBankQuery(input)
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.ExamQuestion{}, access.ExamQuestionAuditFields)
	if err != nil {
		return nil, 0, err
	}
	if input.Keyword != "" {
		query = query.Where("`exam_q_title` LIKE ? OR `exam_q_type` LIKE ?", "%"+input.Keyword+"%", "%"+input.Keyword+"%")
	}
	if input.Category != "" {
		query = query.Where("`exam_q_category` = ?", input.Category)
	}
	if input.Type != "" {
		query = query.Where("`exam_q_type` = ?", input.Type)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ExamQuestion
	if err := query.Order("`add_time` DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func CreateExamQuestionContext(ctx context.Context, input QuestionBankInput) (model.ExamQuestion, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	deptID, err := firstAdminDeptID(db, input.AdminID)
	if err != nil {
		return model.ExamQuestion{}, err
	}
	now := database.Now()
	q := model.ExamQuestion{
		Title:        input.Title,
		Type:         input.Type,
		Schema:       input.Schema,
		Category:     input.Category,
		Tags:         input.Tags,
		Status:       1,
		DeptID:       deptID,
		CreateBy:     input.AdminID,
		UpdateBy:     input.AdminID,
		UpdateDeptID: deptID,
		AddTime:      now,
		EditTime:     now,
	}
	if err := db.Create(&q).Error; err != nil {
		return model.ExamQuestion{}, err
	}
	return q, nil
}

func UpdateExamQuestionContext(ctx context.Context, input QuestionBankInput) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.ExamQuestion{}).Where("`exam_q_id` = ?", input.ID).Updates(map[string]interface{}{
		"exam_q_title":    input.Title,
		"exam_q_type":     input.Type,
		"exam_q_schema":   input.Schema,
		"exam_q_category": input.Category,
		"exam_q_tags":     input.Tags,
		"edit_time":       database.Now(),
	}).Error
}

func UpdateExamQuestionForAdminContext(ctx context.Context, input QuestionBankInput, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.ExamQuestion{}, access.ExamQuestionAuditFields)
	if err != nil {
		return err
	}
	deptID, err := firstAdminDeptID(db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(query.Where("`exam_q_id` = ?", input.ID).Updates(map[string]interface{}{
		"exam_q_title":    input.Title,
		"exam_q_type":     input.Type,
		"exam_q_schema":   input.Schema,
		"exam_q_category": input.Category,
		"exam_q_tags":     input.Tags,
		"update_by":       adminID,
		"update_dept_id":  deptID,
		"edit_time":       database.Now(),
	}))
}

func DeleteExamQuestionContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`exam_q_id` = ?", id).Delete(&model.ExamQuestion{}).Error
}

func DeleteExamQuestionForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.ExamQuestion{}, access.ExamQuestionAuditFields)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(query.Where("`exam_q_id` = ?", id).Delete(&model.ExamQuestion{}))
}

func ExamQuestionCategoriesContext(ctx context.Context) ([]string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var categories []string
	err := db.Model(&model.ExamQuestion{}).
		Where("`exam_q_category` != '' AND `exam_q_category` IS NOT NULL").
		Select("DISTINCT `exam_q_category`").
		Order("`exam_q_category` ASC").
		Pluck("`exam_q_category`", &categories).Error
	return categories, err
}

func ExamQuestionCategoriesForAdminContext(ctx context.Context, adminID uint) ([]string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query, err := access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.ExamQuestion{}, access.ExamQuestionAuditFields)
	if err != nil {
		return nil, err
	}
	var categories []string
	err = query.Where("`exam_q_category` != '' AND `exam_q_category` IS NOT NULL").
		Select("DISTINCT `exam_q_category`").
		Order("`exam_q_category` ASC").
		Pluck("`exam_q_category`", &categories).Error
	return categories, err
}
