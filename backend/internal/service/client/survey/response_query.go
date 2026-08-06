package survey

import (
	"context"
	"encoding/json"

	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ParsedAnswers struct {
	Answers map[string]interface{} `json:"answers"`
}

// Get 读取答卷（管理员用，含答案）
func (r *ResponseService) Get(id uint) (*model.SurveyResponse, error) {
	return r.GetContext(context.Background(), id)
}

func (r *ResponseService) GetContext(ctx context.Context, id uint) (*model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var resp model.SurveyResponse
	if err := db.Where("`survey_resp_id` = ?", id).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *ResponseService) GetForAdminContext(ctx context.Context, id uint, adminID uint) (*model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var resp model.SurveyResponse
	if err := db.Where("`survey_resp_id` = ?", id).First(&resp).Error; err != nil {
		return nil, err
	}
	if err := ensureSurveyVisibleContext(ctx, db, resp.SurveyID, adminID); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetByUser 取某用户对某问卷的答卷
func (r *ResponseService) GetByUser(surveyID uint, userID uint) (*model.SurveyResponse, error) {
	return r.GetByUserContext(context.Background(), surveyID, userID)
}

func (r *ResponseService) GetByUserContext(ctx context.Context, surveyID uint, userID uint) (*model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var resp model.SurveyResponse
	uidStr := userIDToStr(userID, false)
	if err := db.Where("`survey_resp_survey_id` = ? AND `survey_resp_user_id` = ?", surveyID, uidStr).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

// List 答卷列表（admin）
func (r *ResponseService) List(surveyID uint, page, pageSize int, keyword string) ([]model.SurveyResponse, int64, error) {
	return r.ListContext(context.Background(), surveyID, page, pageSize, keyword)
}

func (r *ResponseService) ListContext(ctx context.Context, surveyID uint, page, pageSize int, keyword string) ([]model.SurveyResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", surveyID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("`survey_resp_nickname` LIKE ? OR `survey_resp_user_id` LIKE ? OR `survey_resp_device` LIKE ? OR CAST(`survey_resp_id` AS CHAR) LIKE ?", like, like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SurveyResponse
	err := q.Order("`survey_resp_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *ResponseService) ListForAdminContext(ctx context.Context, surveyID uint, page, pageSize int, keyword string, adminID uint) ([]model.SurveyResponse, int64, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureSurveyVisibleContext(ctx, db, surveyID, adminID); err != nil {
		return nil, 0, err
	}
	return r.ListContext(ctx, surveyID, page, pageSize, keyword)
}

func (r *ResponseService) ListAllBySurvey(surveyID uint) ([]model.SurveyResponse, error) {
	return r.ListAllBySurveyContext(context.Background(), surveyID)
}

func (r *ResponseService) ListAllBySurveyContext(ctx context.Context, surveyID uint) ([]model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var list []model.SurveyResponse
	err := db.Where("`survey_resp_survey_id` = ?", surveyID).Order("`survey_resp_id` ASC").Find(&list).Error
	return list, err
}

func (r *ResponseService) ListAllBySurveyForAdminContext(ctx context.Context, surveyID uint, adminID uint) ([]model.SurveyResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := ensureSurveyVisibleContext(ctx, db, surveyID, adminID); err != nil {
		return nil, err
	}
	return r.ListAllBySurveyContext(ctx, surveyID)
}

func (r *ResponseService) Delete(id uint) error {
	return r.DeleteContext(context.Background(), id)
}

func (r *ResponseService) DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`survey_resp_id` = ?", id).Delete(&model.SurveyResponse{}).Error
}

func (r *ResponseService) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var resp model.SurveyResponse
	if err := db.Where("`survey_resp_id` = ?", id).First(&resp).Error; err != nil {
		return err
	}
	if err := ensureSurveyVisibleContext(ctx, db, resp.SurveyID, adminID); err != nil {
		return err
	}
	return access.RequireRowsAffected(db.Where("`survey_resp_id` = ?", id).Delete(&model.SurveyResponse{}))
}

func (r *ResponseService) BatchDelete(ids []int) error {
	return r.BatchDeleteContext(context.Background(), ids)
}

func (r *ResponseService) BatchDeleteContext(ctx context.Context, ids []int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`survey_resp_id` IN ?", ids).Delete(&model.SurveyResponse{}).Error
}

func (r *ResponseService) BatchDeleteForAdminContext(ctx context.Context, ids []int, adminID uint) error {
	for _, id := range ids {
		if err := r.DeleteForAdminContext(ctx, uint(id), adminID); err != nil {
			return err
		}
	}
	return nil
}

// ParseAnswers 解析答卷答案。
func (r *ResponseService) ParseAnswers(resp *model.SurveyResponse) ParsedAnswers {
	var m map[string]interface{}
	if resp.Answers == "" {
		return ParsedAnswers{Answers: m}
	}
	_ = json.Unmarshal([]byte(resp.Answers), &m)
	return ParsedAnswers{Answers: m}
}
