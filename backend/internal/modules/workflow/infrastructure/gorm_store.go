package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowcore "wecheckin/backend/internal/workflow"
	"wecheckin/backend/pkg/database"
)

var (
	ErrDefinitionNotPublished = errors.New("流程定义尚未发布")
	ErrInstanceNotFound       = errors.New("流程实例不存在")
	ErrTaskNotFound           = errors.New("流程任务不存在")
)

type GormStore struct {
	db    *gorm.DB
	txCtx context.Context
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (store *GormStore) InTransaction(ctx context.Context, fn func(application.TransactionStore) error) error {
	if store == nil || store.db == nil {
		return errors.New("工作流数据库未初始化")
	}
	queryCtx, cancel := database.QueryContext(ctx)
	defer cancel()
	return store.db.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormStore{db: tx, txCtx: queryCtx})
	})
}

func (store *GormStore) LoadPublishedDefinition(ctx context.Context, definitionID uint, version int) (workflowcore.Definition, int, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return workflowcore.Definition{}, 0, err
	}
	defer cancel()

	var definitionModel workflowmodel.Definition
	if err := db.Clauses(clause.Locking{Strength: "SHARE"}).First(&definitionModel, definitionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, 0, fmt.Errorf("流程定义不存在: %w", err)
		}
		return workflowcore.Definition{}, 0, err
	}
	if definitionModel.Status != workflowmodel.DefinitionStatusPublished || definitionModel.CurrentVersion < 1 {
		return workflowcore.Definition{}, 0, ErrDefinitionNotPublished
	}
	if version <= 0 {
		version = definitionModel.CurrentVersion
	}
	return loadDefinitionVersion(db, definitionID, version)
}

func (store *GormStore) ListPublishedDefinitions(ctx context.Context) ([]application.PublishedDefinition, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var rows []workflowmodel.Definition
	if err := db.Where("definition_status = ? AND definition_current_version > 0", workflowmodel.DefinitionStatusPublished).
		Order("definition_category ASC").Order("definition_name ASC").Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]application.PublishedDefinition, 0, len(rows))
	for _, row := range rows {
		definition, version, err := loadDefinitionVersion(db, row.ID, row.CurrentVersion)
		if err != nil {
			return nil, err
		}
		result = append(result, publishedDefinition(row, definition, version))
	}
	return result, nil
}

func (store *GormStore) GetPublishedDefinition(ctx context.Context, definitionID uint) (*application.PublishedDefinition, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row workflowmodel.Definition
	if err := db.First(&row, "id = ? AND definition_status = ? AND definition_current_version > 0", definitionID, workflowmodel.DefinitionStatusPublished).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDefinitionNotPublished
		}
		return nil, err
	}
	definition, version, err := loadDefinitionVersion(db, row.ID, row.CurrentVersion)
	if err != nil {
		return nil, err
	}
	result := publishedDefinition(row, definition, version)
	return &result, nil
}

func (store *GormStore) CreateState(ctx context.Context, state *workflowdomain.State) error {
	if state == nil {
		return errors.New("流程运行状态不能为空")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	now := database.Now()

	instance, err := instanceToModel(state.Instance, state.FormData, now)
	if err != nil {
		return err
	}
	if err := db.Create(&instance).Error; err != nil {
		return err
	}
	if err := createTokens(db, state.Instance.ID, state.Tokens, now); err != nil {
		return err
	}
	if err := createTasks(db, state.Instance.ID, state.Tasks, now, handledActors(state.History)); err != nil {
		return err
	}
	if err := createVariables(db, state.Instance.ID, state.Variables); err != nil {
		return err
	}
	return createHistory(db, state.Instance.ID, state.History, now)
}

func (store *GormStore) LoadStateByTaskForUpdate(ctx context.Context, taskID string) (workflowcore.Definition, *workflowdomain.State, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return workflowcore.Definition{}, nil, err
	}
	defer cancel()

	var task workflowmodel.ProcessTask
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, nil, ErrTaskNotFound
		}
		return workflowcore.Definition{}, nil, err
	}
	var instance workflowmodel.ProcessInstance
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&instance, "id = ?", task.InstanceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, nil, ErrInstanceNotFound
		}
		return workflowcore.Definition{}, nil, err
	}
	definition, _, err := loadDefinitionVersion(db, instance.DefinitionID, instance.DefinitionVersion)
	if err != nil {
		return workflowcore.Definition{}, nil, err
	}
	state, err := loadStateRecords(db, instance)
	return definition, state, err
}

func (store *GormStore) LoadStateByInstanceForUpdate(ctx context.Context, instanceID string) (*workflowdomain.State, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var instance workflowmodel.ProcessInstance
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&instance, "id = ?", instanceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInstanceNotFound
		}
		return nil, err
	}
	return loadStateRecords(db, instance)
}

func (store *GormStore) SaveState(ctx context.Context, state *workflowdomain.State) error {
	if state == nil {
		return errors.New("流程运行状态不能为空")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	now := database.Now()

	formDataJSON, err := encodeFormData(state.FormData)
	if err != nil {
		return err
	}
	instanceUpdates := map[string]interface{}{
		"instance_status": string(state.Instance.Status),
		"form_data_json":  formDataJSON,
	}
	if state.Instance.Status != workflowdomain.InstanceStatusRunning {
		instanceUpdates["end_time"] = now
	}
	if result := db.Model(&workflowmodel.ProcessInstance{}).Where("id = ?", state.Instance.ID).Updates(instanceUpdates); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return ErrInstanceNotFound
	}

	if err := upsertTokens(db, state.Instance.ID, state.Tokens, now); err != nil {
		return err
	}
	if err := upsertTasks(db, state.Instance.ID, state.Tasks, state.History, now); err != nil {
		return err
	}
	if err := upsertVariables(db, state.Instance.ID, state.Variables); err != nil {
		return err
	}
	return createHistory(db, state.Instance.ID, state.History, now)
}

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
	list := make([]application.InstanceSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, instanceSummary(row))
	}
	return &application.InstanceList{List: list, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (store *GormStore) GetInstance(ctx context.Context, instanceID string) (*application.InstanceDetail, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var instance workflowmodel.ProcessInstance
	if err := db.First(&instance, "id = ?", instanceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInstanceNotFound
		}
		return nil, err
	}
	return loadInstanceDetail(db, instance)
}

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
	list := make([]application.TaskSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, taskSummary(row))
	}
	return &application.TaskList{List: list, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func (store *GormStore) contextDB(ctx context.Context) (*gorm.DB, context.CancelFunc, error) {
	if store == nil || store.db == nil {
		return nil, func() {}, errors.New("工作流数据库未初始化")
	}
	if store.txCtx != nil {
		return store.db.WithContext(store.txCtx), func() {}, nil
	}
	queryCtx, cancel := database.QueryContext(ctx)
	return store.db.WithContext(queryCtx), cancel, nil
}

func loadDefinitionVersion(db *gorm.DB, definitionID uint, version int) (workflowcore.Definition, int, error) {
	var versionModel workflowmodel.DefinitionVersion
	if err := db.First(&versionModel, "definition_id = ? AND definition_version = ?", definitionID, version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, 0, fmt.Errorf("流程定义版本不存在: %w", err)
		}
		return workflowcore.Definition{}, 0, err
	}
	var definition workflowcore.Definition
	if err := json.Unmarshal([]byte(versionModel.SourceJSON), &definition); err != nil {
		return workflowcore.Definition{}, 0, fmt.Errorf("解析流程定义版本失败: %w", err)
	}
	if validationErrors := workflowcore.ValidateDefinition(definition); len(validationErrors) > 0 {
		return workflowcore.Definition{}, 0, workflowcore.ValidationErrors(validationErrors)
	}
	return definition, versionModel.Version, nil
}

func instanceToModel(instance workflowdomain.ProcessInstance, formData map[string]interface{}, now int64) (workflowmodel.ProcessInstance, error) {
	formDataJSON, err := encodeFormData(formData)
	if err != nil {
		return workflowmodel.ProcessInstance{}, err
	}
	endTime := int64(0)
	if instance.Status != workflowdomain.InstanceStatusRunning {
		endTime = now
	}
	return workflowmodel.ProcessInstance{
		ID: instance.ID, DefinitionID: instance.DefinitionID,
		DefinitionVersion: instance.DefinitionVersion, DefinitionKey: instance.DefinitionKey,
		BusinessType: instance.BusinessType, BusinessKey: instance.BusinessKey,
		StarterID: instance.StarterID, Status: string(instance.Status), FormDataJSON: formDataJSON,
		StartTime: now, EndTime: endTime,
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
		rows = append(rows, workflowmodel.ProcessHistory{
			ID: event.ID, InstanceID: instanceID, EventType: string(event.Type), NodeID: event.NodeID,
			TaskID: event.TaskID, ActorID: event.ActorID, Message: event.Message, EventTime: now,
		})
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
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
		if handledAt == 0 && (task.Status == workflowdomain.TaskStatusApproved || task.Status == workflowdomain.TaskStatusRejected) {
			handledAt = now
		}
		rows = append(rows, taskToModel(instanceID, task, handledAt, actors))
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"task_status", "task_action", "task_comment", "handled_by", "handled_at", "updated_at",
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

func loadStateRecords(db *gorm.DB, instance workflowmodel.ProcessInstance) (*workflowdomain.State, error) {
	var tokens []workflowmodel.ProcessToken
	var tasks []workflowmodel.ProcessTask
	var variables []workflowmodel.ProcessVariable
	var history []workflowmodel.ProcessHistory
	if err := db.Where("instance_id = ?", instance.ID).Order("created_at ASC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("created_at ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("id ASC").Find(&variables).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("event_time ASC").Order("id ASC").Find(&history).Error; err != nil {
		return nil, err
	}
	return stateFromModels(instance, tokens, tasks, variables, history)
}

func loadInstanceDetail(db *gorm.DB, instance workflowmodel.ProcessInstance) (*application.InstanceDetail, error) {
	var tokens []workflowmodel.ProcessToken
	var tasks []workflowmodel.ProcessTask
	var variables []workflowmodel.ProcessVariable
	var history []workflowmodel.ProcessHistory
	definition, _, err := loadDefinitionVersion(db, instance.DefinitionID, instance.DefinitionVersion)
	if err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("created_at ASC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("created_at ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("id ASC").Find(&variables).Error; err != nil {
		return nil, err
	}
	if err := db.Where("instance_id = ?", instance.ID).Order("event_time ASC").Order("id ASC").Find(&history).Error; err != nil {
		return nil, err
	}
	detail := &application.InstanceDetail{
		Instance: instanceSummary(instance), Variables: make(map[string]interface{}, len(variables)),
		Form:             cloneDefinitionForm(definition),
		FormData:         make(map[string]interface{}),
		FieldPermissions: definitionFieldPermissions(definition),
		StartNodeID:      definitionStartNodeID(definition),
		Tokens:           make([]application.TokenSummary, 0, len(tokens)),
		Tasks:            make([]application.TaskSummary, 0, len(tasks)),
		History:          make([]application.HistorySummary, 0, len(history)),
	}
	formData, err := decodeFormData(instance.FormDataJSON)
	if err != nil {
		return nil, err
	}
	detail.FormData = formData
	for _, row := range variables {
		value, err := decodeVariable(row.ValueJSON)
		if err != nil {
			return nil, err
		}
		detail.Variables[row.Key] = value
	}
	for _, row := range tokens {
		detail.Tokens = append(detail.Tokens, application.TokenSummary{
			ID: row.ID, NodeID: row.NodeID, Status: row.Status,
			BranchGroup: row.BranchGroup, BranchTotal: row.BranchTotal,
		})
	}
	for _, row := range tasks {
		detail.Tasks = append(detail.Tasks, taskSummary(row))
	}
	for _, row := range history {
		detail.History = append(detail.History, application.HistorySummary{
			ID: row.ID, EventType: row.EventType, NodeID: row.NodeID, TaskID: row.TaskID,
			ActorID: row.ActorID, Message: row.Message, EventTime: row.EventTime,
		})
	}
	return detail, nil
}

func instanceSummary(row workflowmodel.ProcessInstance) application.InstanceSummary {
	return application.InstanceSummary{
		ID: row.ID, DefinitionID: row.DefinitionID, DefinitionVersion: row.DefinitionVersion,
		DefinitionKey: row.DefinitionKey, BusinessType: row.BusinessType, BusinessKey: row.BusinessKey,
		StarterID: row.StarterID, Status: row.Status, StartTime: row.StartTime, EndTime: row.EndTime,
	}
}

func publishedDefinition(row workflowmodel.Definition, definition workflowcore.Definition, version int) application.PublishedDefinition {
	return application.PublishedDefinition{
		ID: row.ID, Key: row.Key, Name: row.Name, Description: row.Description,
		Category: row.Category, Version: version, Form: cloneDefinitionForm(definition),
		FieldPermissions: definitionFieldPermissions(definition), StartNodeID: definitionStartNodeID(definition),
	}
}

func cloneDefinitionForm(definition workflowcore.Definition) []workflowcore.FormField {
	return append([]workflowcore.FormField(nil), definition.Form...)
}

func definitionFieldPermissions(definition workflowcore.Definition) map[string][]workflowcore.FieldPermission {
	result := make(map[string][]workflowcore.FieldPermission)
	for _, node := range definition.Nodes {
		if len(node.FormPermissions) == 0 {
			continue
		}
		result[node.ID] = append([]workflowcore.FieldPermission(nil), node.FormPermissions...)
	}
	if result == nil {
		return map[string][]workflowcore.FieldPermission{}
	}
	return result
}

func definitionStartNodeID(definition workflowcore.Definition) string {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart {
			return node.ID
		}
	}
	return ""
}

func taskSummary(row workflowmodel.ProcessTask) application.TaskSummary {
	return application.TaskSummary{
		ID: row.ID, InstanceID: row.InstanceID, NodeID: row.NodeID, NodeName: row.NodeName,
		AssigneeID: row.AssigneeID, ApprovalMode: row.ApprovalMode, CompletionRate: row.CompletionRate,
		Sequence: row.Sequence, Total: row.Total, Status: row.Status, Action: row.Action,
		Comment: row.Comment, HandledBy: row.HandledBy, HandledAt: row.HandledAt,
	}
}

func applyInstanceFilters(db *gorm.DB, query application.InstanceQuery) *gorm.DB {
	if query.DefinitionID > 0 {
		db = db.Where("definition_id = ?", query.DefinitionID)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		db = db.Where("instance_status = ?", value)
	}
	if value := strings.TrimSpace(query.BusinessType); value != "" {
		db = db.Where("business_type = ?", value)
	}
	if value := strings.TrimSpace(query.BusinessKey); value != "" {
		db = db.Where("business_key = ?", value)
	}
	if value := strings.TrimSpace(query.StarterID); value != "" {
		db = db.Where("starter_id = ?", value)
	}
	return db
}

func applyTaskFilters(db *gorm.DB, query application.TaskQuery) *gorm.DB {
	if value := strings.TrimSpace(query.InstanceID); value != "" {
		db = db.Where("instance_id = ?", value)
	}
	if value := strings.TrimSpace(query.AssigneeID); value != "" {
		db = db.Where("task_assignee_id = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		db = db.Where("task_status = ?", value)
	}
	return db
}
