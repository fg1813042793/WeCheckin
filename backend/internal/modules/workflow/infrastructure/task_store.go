package infrastructure

import (
	"context"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

func (store *GormStore) ListTasks(ctx context.Context, query application.TaskQuery) (*application.TaskList, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	base := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), query)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []workflowmodel.ProcessTask
	offset := (query.Page - 1) * query.PageSize
	if err := applyTaskFilters(db.Model(&workflowmodel.ProcessTask{}), query).
		Order("created_at DESC").Order("id DESC").Offset(offset).Limit(query.PageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	list, err := loadTaskSummaries(db, rows)
	if err != nil {
		return nil, err
	}
	return &application.TaskList{List: list, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func taskSummary(row workflowmodel.ProcessTask) (application.TaskSummary, error) {
	images, err := decodeWorkflowImages(row.ImagesJSON)
	if err != nil {
		return application.TaskSummary{}, err
	}
	return application.TaskSummary{
		ID: row.ID, InstanceID: row.InstanceID, NodeID: row.NodeID, NodeName: row.NodeName,
		AssigneeID: row.AssigneeID, ApprovalMode: row.ApprovalMode, CompletionRate: row.CompletionRate,
		Sequence: row.Sequence, Total: row.Total, Status: row.Status, Action: row.Action,
		ApprovalChainKey: row.ApprovalChainKey, ApprovalLayer: row.ApprovalLayer,
		ApprovalLayerTotal: row.ApprovalLayerTotal, SourceDepartmentID: row.SourceDepartmentID,
		SourceDepartmentName: row.SourceDepartmentName,
		Comment:              row.Comment, Images: images, HandledBy: row.HandledBy, HandledAt: row.HandledAt,
	}, nil
}

func loadTaskSummaries(db *gorm.DB, rows []workflowmodel.ProcessTask) ([]application.TaskSummary, error) {
	instanceIDs := taskInstanceIDs(rows)
	instances := make([]workflowmodel.ProcessInstance, 0, len(instanceIDs))
	if len(instanceIDs) > 0 {
		if err := db.Select("id", "definition_id", "starter_id").Where("id IN ?", instanceIDs).Find(&instances).Error; err != nil {
			return nil, err
		}
	}

	userIDs := taskUserIDs(rows)
	seenUserIDs := make(map[uint]struct{}, len(userIDs)+len(instances))
	for _, userID := range userIDs {
		seenUserIDs[userID] = struct{}{}
	}
	for _, instance := range instances {
		appendWorkflowUserID(&userIDs, seenUserIDs, instance.StarterID)
	}
	users, err := loadWorkflowUsers(db, userIDs)
	if err != nil {
		return nil, err
	}
	definitionNames, err := loadWorkflowDefinitionNames(db, instanceDefinitionIDs(instances))
	if err != nil {
		return nil, err
	}
	return taskSummaries(rows, users, instances, definitionNames)
}

func taskInstanceIDs(rows []workflowmodel.ProcessTask) []string {
	ids := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		instanceID := strings.TrimSpace(row.InstanceID)
		if instanceID == "" {
			continue
		}
		if _, exists := seen[instanceID]; exists {
			continue
		}
		seen[instanceID] = struct{}{}
		ids = append(ids, instanceID)
	}
	return ids
}

func taskUserIDs(rows []workflowmodel.ProcessTask) []uint {
	ids := make([]uint, 0, len(rows))
	seen := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		for _, value := range []string{row.AssigneeID, row.HandledBy} {
			id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
			if err != nil || id == 0 {
				continue
			}
			userID := uint(id)
			if _, exists := seen[userID]; exists {
				continue
			}
			seen[userID] = struct{}{}
			ids = append(ids, userID)
		}
	}
	return ids
}

func taskSummaries(
	rows []workflowmodel.ProcessTask,
	users []model.User,
	instances []workflowmodel.ProcessInstance,
	definitionNames map[uint]string,
) ([]application.TaskSummary, error) {
	names := workflowUserNames(users)
	instanceByID := make(map[string]workflowmodel.ProcessInstance, len(instances))
	for _, instance := range instances {
		instanceByID[instance.ID] = instance
	}

	list := make([]application.TaskSummary, 0, len(rows))
	for _, row := range rows {
		summary, err := taskSummary(row)
		if err != nil {
			return nil, err
		}
		summary.AssigneeName = names[strings.TrimSpace(summary.AssigneeID)]
		summary.HandledByName = names[strings.TrimSpace(summary.HandledBy)]
		if instance, exists := instanceByID[summary.InstanceID]; exists {
			summary.DefinitionName = definitionNames[instance.DefinitionID]
			summary.StarterID = instance.StarterID
			summary.StarterName = names[strings.TrimSpace(instance.StarterID)]
		}
		list = append(list, summary)
	}
	return list, nil
}

func loadHistorySummaries(db *gorm.DB, rows []workflowmodel.ProcessHistory) ([]application.HistorySummary, error) {
	ids := make([]uint, 0, len(rows))
	seen := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		appendWorkflowUserID(&ids, seen, row.ActorID)
	}
	users, err := loadWorkflowUsers(db, ids)
	if err != nil {
		return nil, err
	}
	return historySummaries(rows, users)
}

func historySummaries(rows []workflowmodel.ProcessHistory, users []model.User) ([]application.HistorySummary, error) {
	names := workflowUserNames(users)
	result := make([]application.HistorySummary, 0, len(rows))
	for _, row := range rows {
		images, err := decodeWorkflowImages(row.ImagesJSON)
		if err != nil {
			return nil, err
		}
		result = append(result, application.HistorySummary{
			ID: row.ID, EventType: row.EventType, NodeID: row.NodeID, TaskID: row.TaskID,
			ActorID: row.ActorID, ActorName: names[strings.TrimSpace(row.ActorID)],
			Message: row.Message, Images: images, EventTime: row.EventTime,
		})
	}
	return result, nil
}

func loadWorkflowUsers(db *gorm.DB, userIDs []uint) ([]model.User, error) {
	users := make([]model.User, 0, len(userIDs))
	if len(userIDs) == 0 {
		return users, nil
	}
	if err := db.Select("id", "user_name", "user_account").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func workflowUserNames(users []model.User) map[string]string {
	names := make(map[string]string, len(users))
	for _, user := range users {
		if name := workflowUserName(user); name != "" {
			names[strconv.FormatUint(uint64(user.ID), 10)] = name
		}
	}
	return names
}
