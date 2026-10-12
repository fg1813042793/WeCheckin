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

const startSubmissionGuardPeriodKey = "submission_guard"

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

func (store *GormStore) UserDisplayName(ctx context.Context, rawUserID string) (string, error) {
	userID := parsePositiveUint(rawUserID)
	if userID == 0 {
		return "", errors.New("流程发起人无效")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()
	var user model.User
	if err := db.Select("id", "user_name", "user_account").First(&user, userID).Error; err != nil {
		return "", err
	}
	return workflowUserName(user), nil
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
	departmentIDs = normalizeUintIDs(departmentIDs)
	if len(departmentIDs) == 0 {
		return departmentIDs, nil
	}
	var departments []model.Department
	if err := db.Select("id", "dept_parent_id").Find(&departments).Error; err != nil {
		return nil, err
	}
	return departmentIDsWithAncestors(departmentIDs, departments), nil
}

func departmentIDsWithAncestors(departmentIDs []uint, departments []model.Department) []uint {
	parentByID := make(map[uint]uint, len(departments))
	for _, department := range departments {
		parentByID[department.ID] = department.ParentID
	}
	result := normalizeUintIDs(departmentIDs)
	seen := make(map[uint]struct{}, len(result))
	for _, departmentID := range result {
		seen[departmentID] = struct{}{}
	}
	for _, departmentID := range append([]uint(nil), result...) {
		visited := make(map[uint]struct{})
		for parentID := parentByID[departmentID]; parentID > 0; parentID = parentByID[parentID] {
			if _, loop := visited[parentID]; loop {
				break
			}
			visited[parentID] = struct{}{}
			if _, exists := seen[parentID]; exists {
				continue
			}
			seen[parentID] = struct{}{}
			result = append(result, parentID)
		}
	}
	return result
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

func (store *GormStore) CountRunningStartInstances(
	ctx context.Context,
	definitionID uint,
	starterID string,
) (int, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	return countRunningStartInstances(db, definitionID, starterID)
}

func (store *GormStore) LockStartSubmission(
	ctx context.Context,
	definitionID uint,
	starterID string,
) (int, error) {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 || starterID == "" {
		return 0, errors.New("流程发起锁参数无效")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	if _, err := lockStartQuotaUsageRow(db, definitionID, starterID, startSubmissionGuardPeriodKey); err != nil {
		return 0, err
	}
	return countRunningStartInstances(db, definitionID, starterID)
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

	usage, err := lockStartQuotaUsageRow(db, definitionID, starterID, window.PeriodKey)
	if err != nil {
		return 0, false, err
	}

	usedCount, err := countStartQuotaUsage(db, definitionID, starterID, window)
	if err != nil {
		return 0, false, err
	}
	if usedCount >= maxCount {
		return usedCount, false, nil
	}
	if err := db.Model(&workflowmodel.StartQuotaUsage{}).Where("id = ?", usage.ID).
		Update("used_count", usedCount).Error; err != nil {
		return 0, false, err
	}
	return usedCount, true, nil
}

func lockStartQuotaUsageRow(
	db *gorm.DB,
	definitionID uint,
	starterID string,
	periodKey string,
) (workflowmodel.StartQuotaUsage, error) {
	usage := workflowmodel.StartQuotaUsage{
		DefinitionID: definitionID, StarterID: strings.TrimSpace(starterID), PeriodKey: strings.TrimSpace(periodKey),
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "definition_id"}, {Name: "starter_id"}, {Name: "period_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"id": gorm.Expr("id")}),
	}).Create(&usage).Error; err != nil {
		return workflowmodel.StartQuotaUsage{}, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("definition_id = ? AND starter_id = ? AND period_key = ?", definitionID, strings.TrimSpace(starterID), strings.TrimSpace(periodKey)).
		First(&usage).Error; err != nil {
		return workflowmodel.StartQuotaUsage{}, err
	}
	return usage, nil
}

func startQuotaUsageQuery(
	db *gorm.DB,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
) *gorm.DB {
	query := db.Model(&workflowmodel.ProcessInstance{}).
		Where("definition_id = ? AND starter_id = ?", definitionID, strings.TrimSpace(starterID)).
		Where("instance_status = ?", workflowmodel.InstanceStatusCompleted).
		Where("admin_deleted_at = ?", int64(0))
	if window.StartsAt > 0 {
		query = query.Where("start_time >= ?", window.StartsAt)
	}
	if window.EndsAt > 0 {
		query = query.Where("start_time < ?", window.EndsAt)
	}
	return query
}

func runningStartInstanceQuery(db *gorm.DB, definitionID uint, starterID string) *gorm.DB {
	return db.Model(&workflowmodel.ProcessInstance{}).
		Where("definition_id = ? AND starter_id = ?", definitionID, strings.TrimSpace(starterID)).
		Where("instance_status = ?", workflowmodel.InstanceStatusRunning).
		Where("admin_deleted_at = ?", int64(0))
}

func countRunningStartInstances(db *gorm.DB, definitionID uint, starterID string) (int, error) {
	var count int64
	if err := runningStartInstanceQuery(db, definitionID, starterID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func countStartQuotaUsage(
	db *gorm.DB,
	definitionID uint,
	starterID string,
	window workflowcore.StartLimitWindow,
) (int, error) {
	var count int64
	if err := startQuotaUsageQuery(db, definitionID, starterID, window).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
