package formkitadmin

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

type ResourceInput struct {
	OwnerID  uint
	Type     string
	URL      string
	Filename string
	Path     string
	Domain   string
	AddTime  int64
}

type ResourceResult struct {
	ID       uint   `json:"id"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Type     string `json:"type"`
}

func CreateSurveyResourceContext(ctx context.Context, input ResourceInput) (ResourceResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	res := model.SurveyResource{
		SurveyID: input.OwnerID,
		Type:     input.Type,
		URL:      input.URL,
		Filename: input.Filename,
		Path:     input.Path,
		Domain:   input.Domain,
		AddTime:  input.AddTime,
	}
	if err := db.Create(&res).Error; err != nil {
		return ResourceResult{}, err
	}
	return surveyResourceResult(res), nil
}

func ListSurveyResourcesContext(ctx context.Context, surveyID uint, resType string) ([]model.SurveyResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Where("`survey_res_survey_id` = ?", surveyID)
	if resType != "" {
		query = query.Where("`survey_res_type` = ?", resType)
	}
	var list []model.SurveyResource
	if err := query.Order("`survey_res_add_time` DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteSurveyResourceContext(ctx context.Context, id uint) (model.SurveyResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var res model.SurveyResource
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, id).Error; err != nil {
			return err
		}
		return tx.Delete(&res).Error
	})
	return res, err
}

func DeleteSurveyResourceForAdminContext(ctx context.Context, id uint, adminID uint) (model.SurveyResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var res model.SurveyResource
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, id).Error; err != nil {
			return err
		}
		queryBuilder, err := access.ScopedResourceQueryByFieldsContext(ctx, tx, adminID, &model.Survey{}, access.SurveyAuditFields)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`survey_id` = ?", res.SurveyID).First(&model.Survey{}).Error; err != nil {
			return err
		}
		return tx.Delete(&res).Error
	})
	return res, err
}

func CreateExamResourceContext(ctx context.Context, input ResourceInput) (ResourceResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	res := model.ExamResource{
		ExamID:   input.OwnerID,
		Type:     input.Type,
		URL:      input.URL,
		Filename: input.Filename,
		Path:     input.Path,
		Domain:   input.Domain,
		AddTime:  input.AddTime,
	}
	if err := db.Create(&res).Error; err != nil {
		return ResourceResult{}, err
	}
	return examResourceResult(res), nil
}

func ListExamResourcesContext(ctx context.Context, examID uint, resType string) ([]model.ExamResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	query := db.Where("`exam_res_exam_id` = ?", examID)
	if resType != "" {
		query = query.Where("`exam_res_type` = ?", resType)
	}
	var list []model.ExamResource
	if err := query.Order("`exam_res_add_time` DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteExamResourceContext(ctx context.Context, id uint) (model.ExamResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var res model.ExamResource
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, id).Error; err != nil {
			return err
		}
		return tx.Delete(&res).Error
	})
	return res, err
}

func DeleteExamResourceForAdminContext(ctx context.Context, id uint, adminID uint) (model.ExamResource, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var res model.ExamResource
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, id).Error; err != nil {
			return err
		}
		queryBuilder, err := access.ScopedResourceQueryByFieldsContext(ctx, tx, adminID, &model.Exam{}, access.ExamAuditFields)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`exam_id` = ?", res.ExamID).First(&model.Exam{}).Error; err != nil {
			return err
		}
		return tx.Delete(&res).Error
	})
	return res, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func surveyResourceResult(res model.SurveyResource) ResourceResult {
	return ResourceResult{
		ID:       res.ID,
		URL:      res.URL,
		Filename: res.Filename,
		Path:     res.Path,
		Domain:   res.Domain,
		Type:     res.Type,
	}
}

func examResourceResult(res model.ExamResource) ResourceResult {
	return ResourceResult{
		ID:       res.ID,
		URL:      res.URL,
		Filename: res.Filename,
		Path:     res.Path,
		Domain:   res.Domain,
		Type:     res.Type,
	}
}
