// Package survey 提供问卷和答卷业务层。
//
// 设计原则：
//   - 复用 formkit/question、formkit/schema、formkit/calc 作为底层引擎
//   - 不依赖 enroll/event 业务表，完全独立
//   - 调用方只需关心 survey（业务）+ response（答卷）+ statistic（统计）+ channel（渠道）
package survey

import (
	"context"
	"errors"
	"strconv"
	"time"

	"wecheckin/backend/internal/app/formkit/calc"
	"wecheckin/backend/internal/app/formkit/schema"
	"wecheckin/backend/internal/app/support/access"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"

	"gorm.io/gorm"
)

// SurveyService 问卷/表单业务
type SurveyService struct{}

// NewSurveyService ...
func NewSurveyService() *SurveyService { return &SurveyService{} }

var adminSurveyListColumns = []string{
	"survey_id",
	"survey_title",
	"survey_desc",
	"survey_category",
	"survey_tags",
	"survey_cover",
	"survey_visibility",
	"survey_allow_multi",
	"survey_start_time",
	"survey_end_time",
	"survey_max_response",
	"survey_show_result",
	"survey_anonymous",
	"survey_dept_ids",
	"survey_qr",
	"survey_status",
	"survey_mode",
	"survey_order",
	"create_dept_id",
	"create_by",
	"update_by",
	"update_dept_id",
	"add_time",
	"edit_time",
}

func scopedSurveyQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error) {
	return access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Survey{}, access.SurveyAuditFields)
}

func ensureSurveyVisibleContext(ctx context.Context, db *gorm.DB, surveyID uint, adminID uint) error {
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return queryBuilder.Where("`survey_id` = ?", surveyID).First(&model.Survey{}).Error
}

type DetailResult struct {
	Survey        *model.Survey `json:"survey"`
	ResponseCount int64         `json:"responseCount"`
	Schema        string        `json:"schema"`
}

type LogicApplyResult struct {
	Answers map[string]interface{} `json:"answers"`
	States  map[string]string      `json:"states,omitempty"`
}

// Create 创建问卷
func (s *SurveyService) Create(sv *model.Survey) error {
	return s.CreateContext(context.Background(), sv)
}

func (s *SurveyService) CreateContext(ctx context.Context, sv *model.Survey) error {
	if sv.Title == "" {
		return errors.New("标题必填")
	}
	if sv.Schema == "" {
		sv.Schema = `{"version":"2.0","questions":[],"setting":{}}`
	}
	if _, err := schema.Parse(sv.Schema); err != nil {
		return errors.New("schema 无效: " + err.Error())
	}
	now := time.Now().UnixMilli()
	sv.AddTime = now
	sv.EditTime = now
	if sv.UpdateBy == 0 {
		sv.UpdateBy = sv.CreateBy
	}
	if sv.UpdateDeptID == 0 {
		sv.UpdateDeptID = sv.DeptID
	}
	if sv.Status == 0 {
		sv.Status = 1
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Create(sv).Error
}

// Update 更新问卷
func (s *SurveyService) Update(sv *model.Survey) error {
	return s.UpdateContext(context.Background(), sv)
}

func (s *SurveyService) UpdateContext(ctx context.Context, sv *model.Survey) error {
	if sv.ID == 0 {
		return errors.New("ID 必填")
	}
	if _, err := schema.Parse(sv.Schema); err != nil {
		return errors.New("schema 无效: " + err.Error())
	}
	sv.EditTime = time.Now().UnixMilli()
	updates := map[string]interface{}{
		"survey_title":        sv.Title,
		"survey_desc":         sv.Desc,
		"survey_schema":       sv.Schema,
		"survey_category":     sv.Category,
		"survey_tags":         sv.Tags,
		"survey_cover":        sv.Cover,
		"survey_visibility":   sv.Visibility,
		"survey_allow_multi":  sv.AllowMulti,
		"survey_start_time":   sv.StartTime,
		"survey_end_time":     sv.EndTime,
		"survey_max_response": sv.MaxResponse,
		"survey_show_result":  sv.ShowResult,
		"survey_anonymous":    sv.Anonymous,
		"survey_dept_ids":     sv.DeptIDs,
		"survey_status":       sv.Status,
		"survey_mode":         sv.Mode,
		"survey_settings":     sv.Settings,
		"survey_order":        sv.Order,
		"create_dept_id":      sv.DeptID,
		"update_dept_id":      sv.DeptID,
		"edit_time":           sv.EditTime,
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Survey{}).Where("`survey_id` = ?", sv.ID).Updates(updates).Error
}

func (s *SurveyService) UpdateForAdminContext(ctx context.Context, sv *model.Survey, adminID uint) error {
	if sv.ID == 0 {
		return errors.New("ID 必填")
	}
	if _, err := schema.Parse(sv.Schema); err != nil {
		return errors.New("schema 无效: " + err.Error())
	}
	sv.EditTime = time.Now().UnixMilli()
	updates := map[string]interface{}{
		"survey_title":        sv.Title,
		"survey_desc":         sv.Desc,
		"survey_schema":       sv.Schema,
		"survey_category":     sv.Category,
		"survey_tags":         sv.Tags,
		"survey_cover":        sv.Cover,
		"survey_visibility":   sv.Visibility,
		"survey_allow_multi":  sv.AllowMulti,
		"survey_start_time":   sv.StartTime,
		"survey_end_time":     sv.EndTime,
		"survey_max_response": sv.MaxResponse,
		"survey_show_result":  sv.ShowResult,
		"survey_anonymous":    sv.Anonymous,
		"survey_dept_ids":     sv.DeptIDs,
		"survey_status":       sv.Status,
		"survey_mode":         sv.Mode,
		"survey_settings":     sv.Settings,
		"survey_order":        sv.Order,
		"create_dept_id":      sv.DeptID,
		"update_by":           adminID,
		"update_dept_id":      sv.DeptID,
		"edit_time":           sv.EditTime,
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`survey_id` = ?", sv.ID).Updates(updates))
}

// Get 读取问卷（含 schema）
func (s *SurveyService) Get(id uint) (*model.Survey, error) {
	return s.GetContext(context.Background(), id)
}

func (s *SurveyService) GetContext(ctx context.Context, id uint) (*model.Survey, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var sv model.Survey
	if err := db.Where("`survey_id` = ?", id).First(&sv).Error; err != nil {
		return nil, err
	}
	return &sv, nil
}

func (s *SurveyService) GetForAdminContext(ctx context.Context, id uint, adminID uint) (*model.Survey, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var sv model.Survey
	if err := queryBuilder.Where("`survey_id` = ?", id).First(&sv).Error; err != nil {
		return nil, err
	}
	return &sv, nil
}

func (s *SurveyService) Detail(id uint) (*DetailResult, error) {
	return s.DetailContext(context.Background(), id)
}

func (s *SurveyService) DetailContext(ctx context.Context, id uint) (*DetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var sv model.Survey
	if err := db.Where("`survey_id` = ?", id).First(&sv).Error; err != nil {
		return nil, err
	}
	var responseCount int64
	if err := db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", id).Count(&responseCount).Error; err != nil {
		return nil, err
	}
	return &DetailResult{
		Survey:        &sv,
		ResponseCount: responseCount,
		Schema:        sv.Schema,
	}, nil
}

func (s *SurveyService) DetailForAdminContext(ctx context.Context, id uint, adminID uint) (*DetailResult, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var sv model.Survey
	if err := queryBuilder.Where("`survey_id` = ?", id).First(&sv).Error; err != nil {
		return nil, err
	}
	var responseCount int64
	if err := db.Model(&model.SurveyResponse{}).Where("`survey_resp_survey_id` = ?", id).Count(&responseCount).Error; err != nil {
		return nil, err
	}
	return &DetailResult{
		Survey:        &sv,
		ResponseCount: responseCount,
		Schema:        sv.Schema,
	}, nil
}

// List 列表（支持 keyword/category/status 过滤）
func (s *SurveyService) List(keyword, category string, status, page, pageSize int) ([]model.Survey, int64, error) {
	return s.ListContext(context.Background(), keyword, category, status, page, pageSize)
}

func (s *SurveyService) ListContext(ctx context.Context, keyword, category string, status, page, pageSize int) ([]model.Survey, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q := db.Model(&model.Survey{})
	if keyword != "" {
		q = q.Where("`survey_title` LIKE ?", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`survey_category` = ?", category)
	}
	if status >= 0 {
		q = q.Where("`survey_status` = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Survey
	err := q.Order("`survey_order` DESC, `survey_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *SurveyService) ListForAdminContext(ctx context.Context, keyword, category string, status, page, pageSize int, adminID uint) ([]model.Survey, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	q, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, 0, err
	}
	if keyword != "" {
		q = q.Where("`survey_title` LIKE ?", "%"+keyword+"%")
	}
	if category != "" {
		q = q.Where("`survey_category` = ?", category)
	}
	if status >= 0 {
		q = q.Where("`survey_status` = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Survey
	err = q.Select(adminSurveyListColumns).Order("`survey_order` DESC, `survey_id` DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *SurveyService) ResponseCounts(ctx context.Context, surveyIDs []uint) (map[uint]int, error) {
	return s.ResponseCountsContext(ctx, surveyIDs)
}

func (s *SurveyService) ResponseCountsContext(ctx context.Context, surveyIDs []uint) (map[uint]int, error) {
	countMap := make(map[uint]int, len(surveyIDs))
	if len(surveyIDs) == 0 {
		return countMap, nil
	}
	type responseCountRow struct {
		SurveyID uint `gorm:"column:survey_resp_survey_id"`
		Count    int  `gorm:"column:cnt"`
	}
	var rows []responseCountRow
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if err := db.Model(&model.SurveyResponse{}).
		Select("`survey_resp_survey_id`, COUNT(*) AS cnt").
		Where("`survey_resp_survey_id` IN ?", surveyIDs).
		Group("`survey_resp_survey_id`").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		countMap[row.SurveyID] = row.Count
	}
	return countMap, nil
}

func (s *SurveyService) SetStatus(id uint, status int) error {
	return s.SetStatusContext(context.Background(), id, status)
}

func (s *SurveyService) SetStatusContext(ctx context.Context, id uint, status int) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.Survey{}).Where("`survey_id` = ?", id).Update("survey_status", status).Error
}

func (s *SurveyService) SetStatusForAdminContext(ctx context.Context, id uint, status int, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return err
	}
	return access.RequireRowsAffected(queryBuilder.Where("`survey_id` = ?", id).Update("survey_status", status))
}

func (s *SurveyService) Copy(id, createBy, deptID uint) (*model.Survey, error) {
	return s.CopyContext(context.Background(), id, createBy, deptID)
}

func (s *SurveyService) CopyContext(ctx context.Context, id, createBy, deptID uint) (*model.Survey, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var source model.Survey
	if err := db.Where("`survey_id` = ?", id).First(&source).Error; err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	copySurvey := source
	copySurvey.ID = 0
	copySurvey.Title = source.Title + " (副本)"
	copySurvey.Status = 0
	copySurvey.AddTime = now
	copySurvey.EditTime = now
	copySurvey.CreateBy = createBy
	copySurvey.DeptID = deptID
	if err := db.Create(&copySurvey).Error; err != nil {
		return nil, err
	}
	return &copySurvey, nil
}

func (s *SurveyService) CopyForAdminContext(ctx context.Context, id, createBy, deptID, adminID uint) (*model.Survey, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	queryBuilder, err := scopedSurveyQueryContext(ctx, db, adminID)
	if err != nil {
		return nil, err
	}
	var source model.Survey
	if err := queryBuilder.Where("`survey_id` = ?", id).First(&source).Error; err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	copySurvey := source
	copySurvey.ID = 0
	copySurvey.Title = source.Title + " (副本)"
	copySurvey.Status = 0
	copySurvey.AddTime = now
	copySurvey.EditTime = now
	copySurvey.CreateBy = createBy
	copySurvey.DeptID = deptID
	if err := db.Create(&copySurvey).Error; err != nil {
		return nil, err
	}
	return &copySurvey, nil
}

// Delete 删除问卷（同时删答卷和渠道）
func (s *SurveyService) Delete(id uint) error {
	return s.DeleteContext(context.Background(), id)
}

func (s *SurveyService) DeleteContext(ctx context.Context, id uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`survey_id` = ?", id).Delete(&model.Survey{}).Error; err != nil {
			return err
		}
		if err := tx.Where("`survey_resp_survey_id` = ?", id).Delete(&model.SurveyResponse{}).Error; err != nil {
			return err
		}
		return tx.Where("`survey_ch_survey_id` = ?", id).Delete(&model.SurveyChannel{}).Error
	})
}

func (s *SurveyService) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		queryBuilder, err := scopedSurveyQueryContext(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if err := queryBuilder.Where("`survey_id` = ?", id).First(&model.Survey{}).Error; err != nil {
			return err
		}
		if err := access.RequireRowsAffected(tx.Where("`survey_id` = ?", id).Delete(&model.Survey{})); err != nil {
			return err
		}
		if err := tx.Where("`survey_resp_survey_id` = ?", id).Delete(&model.SurveyResponse{}).Error; err != nil {
			return err
		}
		return tx.Where("`survey_ch_survey_id` = ?", id).Delete(&model.SurveyChannel{}).Error
	})
}

// Visible 判断 userID+部门 是否有权访问
func (s *SurveyService) Visible(sv *model.Survey, userID uint, deptIDs []uint) bool {
	if sv.Visibility == 0 {
		return true
	}
	if userID == 0 {
		return false
	}
	if sv.Visibility == 1 {
		return true
	}
	if len(deptIDs) == 0 {
		return false
	}
	for _, did := range deptIDs {
		if contains(sv.DeptIDs, did) {
			return true
		}
	}
	return false
}

func contains(s string, id uint) bool {
	idStr := strconv.FormatUint(uint64(id), 10)
	if s == "" {
		return false
	}
	for _, part := range splitComma(s) {
		if part == idStr {
			return true
		}
	}
	return false
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// ApplyLogic 应用 schema 逻辑（公开，给 H5 端调）
func (s *SurveyService) ApplyLogic(sv *model.Survey, answers map[string]interface{}) (LogicApplyResult, error) {
	sch, err := schema.Parse(sv.Schema)
	if err != nil {
		return LogicApplyResult{Answers: answers}, err
	}
	eng := calc.New()
	ans2, err := eng.ApplyCalcValues(sch, answers)
	if err != nil {
		return LogicApplyResult{Answers: answers}, err
	}
	states, _ := eng.ApplyLogic(sch, ans2)
	return LogicApplyResult{Answers: ans2, States: states}, nil
}
