package infrastructure

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func (store *GormStore) ListInstances(ctx context.Context, query application.InstanceQuery) (*application.InstanceList, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	base := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []workflowmodel.ProcessInstance
	offset := (query.Page - 1) * query.PageSize
	if err := applyInstanceFilters(db.Model(&workflowmodel.ProcessInstance{}), query).
		Order("start_time DESC").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	list, err := loadInstanceSummaries(db, rows)
	if err != nil {
		return nil, err
	}
	return &application.InstanceList{List: list, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (store *GormStore) HideStartedInstance(ctx context.Context, instanceID, starterID string, deletedAt int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Model(&workflowmodel.ProcessInstance{}).
		Where("id = ? AND starter_id = ?", strings.TrimSpace(instanceID), strings.TrimSpace(starterID)).
		Update("starter_deleted_at", deletedAt).Error
}

func (store *GormStore) LoadInstancesForDelete(ctx context.Context, instanceIDs []string) ([]application.InstanceSummary, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var rows []workflowmodel.ProcessInstance
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "instance_status").
		Where("id IN ? AND admin_deleted_at = 0", instanceIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]application.InstanceSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, application.InstanceSummary{ID: row.ID, Status: row.Status})
	}
	return result, nil
}

func (store *GormStore) SoftDeleteInstances(ctx context.Context, instanceIDs []string, actorID string, deletedAt int64) (int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	result := db.Model(&workflowmodel.ProcessInstance{}).
		Where("id IN ? AND admin_deleted_at = 0 AND instance_status IN ?", instanceIDs, []string{
			workflowmodel.InstanceStatusCompleted,
			workflowmodel.InstanceStatusRejected,
			workflowmodel.InstanceStatusCancelled,
			"withdrawn",
		}).
		Updates(map[string]interface{}{"admin_deleted_at": deletedAt, "admin_deleted_by": strings.TrimSpace(actorID)})
	return result.RowsAffected, result.Error
}

func (store *GormStore) LoadTaskForDelete(ctx context.Context, taskID string) (*application.TaskSummary, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row workflowmodel.ProcessTask
	err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "task_status").
		Where("id = ? AND admin_deleted_at = 0", strings.TrimSpace(taskID)).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &application.TaskSummary{ID: row.ID, Status: row.Status}, nil
}

func (store *GormStore) SoftDeleteTask(ctx context.Context, taskID, actorID string, deletedAt int64) (int64, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()
	result := db.Model(&workflowmodel.ProcessTask{}).
		Where("id = ? AND admin_deleted_at = 0 AND task_status IN ?", strings.TrimSpace(taskID), []string{
			workflowmodel.TaskStatusCompleted,
			workflowmodel.TaskStatusApproved,
			workflowmodel.TaskStatusRejected,
			workflowmodel.TaskStatusReturned,
			workflowmodel.TaskStatusCancelled,
		}).
		Updates(map[string]interface{}{
			"admin_deleted_at": deletedAt,
			"admin_deleted_by": strings.TrimSpace(actorID),
		})
	return result.RowsAffected, result.Error
}

func (store *GormStore) GetInstance(ctx context.Context, instanceID string) (*application.InstanceDetail, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var instance workflowmodel.ProcessInstance
	if err := db.First(&instance, "id = ? AND admin_deleted_at = 0", instanceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInstanceNotFound
		}
		return nil, err
	}
	return loadInstanceDetail(db, instance)
}

func (store *GormStore) FindStateByBusiness(ctx context.Context, businessType, businessKey string) (*workflowdomain.State, bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, false, err
	}
	defer cancel()
	var instance workflowmodel.ProcessInstance
	err = db.Clauses(clause.Locking{Strength: "SHARE"}).
		First(&instance, "business_type = ? AND business_key = ?", strings.TrimSpace(businessType), strings.TrimSpace(businessKey)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	state, err := loadStateRecords(db, instance)
	if err != nil {
		return nil, false, err
	}
	return state, true, nil
}
