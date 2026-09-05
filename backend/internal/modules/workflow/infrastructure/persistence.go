package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func instanceToModel(instance workflowdomain.ProcessInstance, formData map[string]interface{}, now int64) (workflowmodel.ProcessInstance, error) {
	formDataJSON, err := encodeFormData(formData)
	if err != nil {
		return workflowmodel.ProcessInstance{}, err
	}
	endTime := int64(0)
	if instance.Status != workflowdomain.InstanceStatusRunning {
		endTime = now
	}
	startTime := instance.StartTime
	if startTime <= 0 {
		startTime = now
	}
	return workflowmodel.ProcessInstance{
		ID: instance.ID, DefinitionID: instance.DefinitionID,
		DefinitionVersion: instance.DefinitionVersion, DefinitionKey: instance.DefinitionKey,
		BusinessType: instance.BusinessType, BusinessKey: instance.BusinessKey,
		StarterID: instance.StarterID, OperatorID: instance.OperatorID,
		Status: string(instance.Status), FormDataJSON: formDataJSON,
		FormRevision: instance.FormRevision,
		StartTime:    startTime, EndTime: endTime,
	}, nil
}

func createTokens(db *gorm.DB, instanceID string, tokens []workflowdomain.Token, now int64) error {
	if len(tokens) == 0 {
		return nil
	}
	rows := make([]workflowmodel.ProcessToken, 0, len(tokens))
	for _, token := range tokens {
		completedAt := int64(0)
		if token.Status == workflowdomain.TokenStatusCompleted || token.Status == workflowdomain.TokenStatusCancelled {
			completedAt = now
		}
		rows = append(rows, workflowmodel.ProcessToken{
			ID: token.ID, InstanceID: instanceID, NodeID: token.NodeID, Status: string(token.Status),
			BranchGroup: token.BranchGroup, BranchTotal: token.BranchTotal, ArrivedAt: now, CompletedAt: completedAt,
		})
	}
	return db.Create(&rows).Error
}

func createTasks(db *gorm.DB, instanceID string, tasks []workflowdomain.Task, now int64, actors map[string]string) error {
	if len(tasks) == 0 {
		return nil
	}
	rows := make([]workflowmodel.ProcessTask, 0, len(tasks))
	for _, task := range tasks {
		rows = append(rows, taskToModel(instanceID, task, now, actors))
	}
	return db.Create(&rows).Error
}

func createVariables(db *gorm.DB, instanceID string, variables map[string]interface{}) error {
	if len(variables) == 0 {
		return nil
	}
	rows, err := variableModels(instanceID, variables)
	if err != nil {
		return err
	}
	return db.Create(&rows).Error
}

func createHistory(db *gorm.DB, instanceID string, history []workflowdomain.HistoryEvent, now int64) error {
	if len(history) == 0 {
		return nil
	}
	rows := make([]workflowmodel.ProcessHistory, 0, len(history))
	for _, event := range history {
		eventTime := event.EventTime
		if eventTime <= 0 {
			eventTime = now
		}
		rows = append(rows, workflowmodel.ProcessHistory{
			ID: event.ID, InstanceID: instanceID, EventType: string(event.Type), NodeID: event.NodeID,
			TaskID: event.TaskID, ActorID: event.ActorID, Message: event.Message,
			ImagesJSON: encodeWorkflowImages(event.Images), EventTime: eventTime,
		})
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (store *GormStore) AppendInstanceHistory(ctx context.Context, instanceID string, event workflowdomain.HistoryEvent, eventTime int64) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Create(&workflowmodel.ProcessHistory{
		ID: event.ID, InstanceID: instanceID, EventType: string(event.Type), NodeID: event.NodeID,
		TaskID: event.TaskID, ActorID: event.ActorID, Message: event.Message,
		ImagesJSON: encodeWorkflowImages(event.Images), EventTime: eventTime,
	}).Error
}

func upsertTokens(db *gorm.DB, instanceID string, tokens []workflowdomain.Token, now int64) error {
	if len(tokens) == 0 {
		return nil
	}
	rows := make([]workflowmodel.ProcessToken, 0, len(tokens))
	for _, token := range tokens {
		completedAt := int64(0)
		if token.Status == workflowdomain.TokenStatusCompleted || token.Status == workflowdomain.TokenStatusCancelled {
			completedAt = now
		}
		rows = append(rows, workflowmodel.ProcessToken{
			ID: token.ID, InstanceID: instanceID, NodeID: token.NodeID, Status: string(token.Status),
			BranchGroup: token.BranchGroup, BranchTotal: token.BranchTotal, ArrivedAt: now, CompletedAt: completedAt,
		})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"node_id", "token_status", "branch_group", "branch_total", "completed_at", "updated_at"}),
	}).Create(&rows).Error
}

func upsertTasks(db *gorm.DB, instanceID string, tasks []workflowdomain.Task, history []workflowdomain.HistoryEvent, now int64) error {
	if len(tasks) == 0 {
		return nil
	}
	var existing []workflowmodel.ProcessTask
	if err := db.Where("instance_id = ?", instanceID).Find(&existing).Error; err != nil {
		return err
	}
	existingHandledAt := make(map[string]int64, len(existing))
	for _, task := range existing {
		existingHandledAt[task.ID] = task.HandledAt
	}
	actors := handledActors(history)
	rows := make([]workflowmodel.ProcessTask, 0, len(tasks))
	for _, task := range tasks {
		handledAt := existingHandledAt[task.ID]
		if handledAt == 0 && (task.Status == workflowdomain.TaskStatusCompleted || task.Status == workflowdomain.TaskStatusApproved || task.Status == workflowdomain.TaskStatusRejected || task.Status == workflowdomain.TaskStatusReturned) {
			handledAt = now
		}
		rows = append(rows, taskToModel(instanceID, task, handledAt, actors))
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"task_status", "task_action", "task_comment", "task_images_json", "handled_by", "handled_at", "updated_at",
		}),
	}).Create(&rows).Error
}

func upsertVariables(db *gorm.DB, instanceID string, variables map[string]interface{}) error {
	if len(variables) == 0 {
		return nil
	}
	rows, err := variableModels(instanceID, variables)
	if err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "instance_id"}, {Name: "variable_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"variable_value_json", "updated_at"}),
	}).Create(&rows).Error
}

func variableModels(instanceID string, variables map[string]interface{}) ([]workflowmodel.ProcessVariable, error) {
	rows := make([]workflowmodel.ProcessVariable, 0, len(variables))
	for key, value := range variables {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("序列化流程变量 %s 失败: %w", key, err)
		}
		rows = append(rows, workflowmodel.ProcessVariable{InstanceID: instanceID, Key: key, ValueJSON: string(encoded)})
	}
	return rows, nil
}
