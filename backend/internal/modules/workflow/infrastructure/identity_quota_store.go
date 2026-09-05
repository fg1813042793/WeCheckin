package infrastructure

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/workflowcore"
)

func (store *GormStore) IsActiveUser(ctx context.Context, userID string) (bool, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(userID), 10, 64)
	if err != nil || id == 0 {
		return false, nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	var count int64
	if err := db.Table("users").Where("id = ? AND user_status = 1", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (store *GormStore) UserDepartmentIDs(ctx context.Context, rawUserID string) ([]uint, error) {
	userID := parsePositiveUint(rawUserID)
	if userID == 0 {
		return nil, nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var departmentIDs []uint
	if err := db.Table("user_depts").
		Where("user_dept_user_id = ?", userID).
		Order("user_dept_dept_id ASC").
		Pluck("user_dept_dept_id", &departmentIDs).Error; err != nil {
		return nil, err
	}
	return normalizeUintIDs(departmentIDs), nil
}

func (store *GormStore) CanOperatorStartFor(ctx context.Context, operatorID, starterID string) (bool, error) {
	operator, err := strconv.ParseUint(strings.TrimSpace(operatorID), 10, 64)
	if err != nil || operator == 0 {
		return false, nil
	}
	starter, err := strconv.ParseUint(strings.TrimSpace(starterID), 10, 64)
	if err != nil || starter == 0 {
		return false, nil
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	var admin model.Admin
	if err := db.First(&admin, uint(operator)).Error; err != nil {
		return false, err
	}
	query := db.Model(&model.User{}).Where("id = ? AND user_status = 1", starter)
	if where, args := access.UserDataScopeFilterWithDBContext(ctx, db, &admin); where != "" {
		query = query.Where(where, args...)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (store *GormStore) CountStartQuotaUsage(
	ctx context.Context,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
) (int, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	return countStartQuotaUsage(db, definitionID, starterID, window)
}

func (store *GormStore) ConsumeStartQuota(
	ctx context.Context,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
	maxCount int,
) (int, bool, error) {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 || starterID == "" || strings.TrimSpace(window.PeriodKey) == "" || maxCount < 1 {
		return 0, false, errors.New("流程发起额度参数无效")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, false, err
	}
	defer cancel()

	usage := workflowmodel.StartQuotaUsage{
		DefinitionID: definitionID, StarterID: starterID, PeriodKey: window.PeriodKey,
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "definition_id"}, {Name: "starter_id"}, {Name: "period_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"id": gorm.Expr("id")}),
	}).Create(&usage).Error; err != nil {
		return 0, false, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("definition_id = ? AND starter_id = ? AND period_key = ?", definitionID, starterID, window.PeriodKey).
		First(&usage).Error; err != nil {
		return 0, false, err
	}

	usedCount, err := countStartQuotaUsage(db, definitionID, starterID, window)
	if err != nil {
		return 0, false, err
	}
	if usedCount >= maxCount {
		return usedCount, false, nil
	}
	nextCount := usedCount + 1
	if err := db.Model(&workflowmodel.StartQuotaUsage{}).Where("id = ?", usage.ID).
		Update("used_count", nextCount).Error; err != nil {
		return 0, false, err
	}
	return nextCount, true, nil
}

func countStartQuotaUsage(
	db *gorm.DB,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
) (int, error) {
	query := db.Model(&workflowmodel.ProcessInstance{}).
		Where("definition_id = ? AND starter_id = ?", definitionID, strings.TrimSpace(starterID))
	if window.StartsAt > 0 {
		query = query.Where("start_time >= ?", window.StartsAt)
	}
	if window.EndsAt > 0 {
		query = query.Where("start_time < ?", window.EndsAt)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
