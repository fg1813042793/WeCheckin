package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/support/access"
	"wecheckin/backend/internal/workflowcore"
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

func (store *GormStore) GetStartDraft(ctx context.Context, definitionID uint, starterID string) (*application.StartDraft, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row workflowmodel.StartDraft
	if err := db.First(&row, "definition_id = ? AND starter_id = ?", definitionID, strings.TrimSpace(starterID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	formData := make(map[string]interface{})
	if strings.TrimSpace(row.FormDataJSON) != "" {
		if err := json.Unmarshal([]byte(row.FormDataJSON), &formData); err != nil {
			return nil, fmt.Errorf("解析流程草稿失败: %w", err)
		}
	}
	return &application.StartDraft{
		DefinitionID: row.DefinitionID, DefinitionVersion: row.DefinitionVersion,
		StarterID: row.StarterID, FormData: formData, UpdatedAt: row.EditTime,
	}, nil
}

func (store *GormStore) SaveStartDraft(ctx context.Context, draft application.StartDraft) (*application.StartDraft, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	encoded, err := json.Marshal(draft.FormData)
	if err != nil {
		return nil, fmt.Errorf("序列化流程草稿失败: %w", err)
	}
	row := workflowmodel.StartDraft{
		DefinitionID: draft.DefinitionID, DefinitionVersion: draft.DefinitionVersion,
		StarterID: strings.TrimSpace(draft.StarterID), FormDataJSON: string(encoded), EditTime: draft.UpdatedAt,
	}
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "definition_id"}, {Name: "starter_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"definition_version": row.DefinitionVersion,
			"form_data_json":     row.FormDataJSON,
			"edit_time":          row.EditTime,
			"updated_at":         database.Now(),
		}),
	}).Create(&row).Error; err != nil {
		return nil, err
	}
	result := draft
	return &result, nil
}

func (store *GormStore) DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return db.Where("definition_id = ? AND starter_id = ?", definitionID, strings.TrimSpace(starterID)).
		Delete(&workflowmodel.StartDraft{}).Error
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

func (store *GormStore) LoadDefinitionAndStateByInstanceForUpdate(ctx context.Context, instanceID string) (workflowcore.Definition, *workflowdomain.State, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return workflowcore.Definition{}, nil, err
	}
	defer cancel()
	var instance workflowmodel.ProcessInstance
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&instance, "id = ?", instanceID).Error; err != nil {
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

func (store *GormStore) PersistEffects(ctx context.Context, state *workflowdomain.State) ([]string, error) {
	if state == nil {
		return nil, errors.New("流程运行状态不能为空")
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	now := database.Now()
	if len(state.Participants) > 0 {
		rows := make([]workflowmodel.InstanceParticipant, 0, len(state.Participants))
		for _, participant := range state.Participants {
			rows = append(rows, workflowmodel.InstanceParticipant{
				ID: participant.ID, InstanceID: state.Instance.ID, UserID: participant.UserID,
				Role: string(participant.Role), NodeID: participant.NodeID, AddTime: now,
			})
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	if len(state.NotificationIntents) == 0 {
		return nil, nil
	}
	starterName := workflowUserDisplayName(db, state.Instance.StarterID)
	outboxIDs := make([]string, 0)
	for _, intent := range state.NotificationIntents {
		payload := renderNotificationPayload(state, intent, starterName)
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("序列化流程通知失败: %w", err)
		}
		channels := append([]string(nil), intent.Config.Channels...)
		sort.Strings(channels)
		for _, channel := range channels {
			channel = strings.TrimSpace(channel)
			outboxID := intent.ID + "-" + channel
			taskKey := strings.TrimSpace(intent.TaskID)
			if taskKey == "" {
				taskKey = "-"
			}
			row := workflowmodel.NotificationOutbox{
				ID: outboxID, InstanceID: state.Instance.ID, NodeID: intent.NodeID, TaskID: intent.TaskID,
				RecipientUserID: intent.RecipientUserID, Kind: string(intent.Kind), Channel: channel,
				Status:      workflowmodel.NotificationStatusPending,
				DedupeKey:   strings.Join([]string{state.Instance.ID, string(intent.Kind), intent.NodeID, taskKey, intent.RecipientUserID, channel}, ":"),
				PayloadJSON: string(payloadJSON), NextRetryAt: now, AddTime: now, EditTime: now,
			}
			result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
			if result.Error != nil {
				return nil, result.Error
			}
			if result.RowsAffected > 0 {
				outboxIDs = append(outboxIDs, outboxID)
			}
		}
	}
	return outboxIDs, nil
}

func (store *GormStore) HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return false, err
	}
	defer cancel()
	var count int64
	err = db.Model(&workflowmodel.InstanceParticipant{}).
		Where("instance_id = ? AND user_id = ? AND participant_role = ?", strings.TrimSpace(instanceID), strings.TrimSpace(userID), strings.TrimSpace(role)).
		Count(&count).Error
	return count > 0, err
}

func workflowUserDisplayName(db *gorm.DB, userID string) string {
	fallback := strings.TrimSpace(userID)
	id, err := strconv.ParseUint(fallback, 10, 64)
	if err != nil || id == 0 {
		return fallback
	}
	var user model.User
	if err := db.Select("id", "user_name", "user_account").First(&user, uint(id)).Error; err != nil {
		return fallback
	}
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	if account := strings.TrimSpace(user.Account); account != "" {
		return account
	}
	return fallback
}

func renderNotificationPayload(state *workflowdomain.State, intent workflowdomain.NotificationIntent, starterName string) application.NotificationPayload {
	replacements := map[string]string{
		"{{workflowName}}": intent.WorkflowName,
		"{{nodeName}}":     intent.NodeName,
		"{{starterName}}":  starterName,
		"{{instanceId}}":   state.Instance.ID,
		"{{taskId}}":       intent.TaskID,
	}
	render := func(value string, limit int) string {
		for token, replacement := range replacements {
			value = strings.ReplaceAll(value, token, replacement)
		}
		value = strings.TrimSpace(value)
		if utf8.RuneCountInString(value) <= limit {
			return value
		}
		return string([]rune(value)[:limit])
	}
	return application.NotificationPayload{
		Title: render(intent.Config.Title, 64), Content: render(intent.Config.Content, 1000),
		WorkflowName: intent.WorkflowName, NodeName: intent.NodeName,
		StarterID: state.Instance.StarterID, StarterName: starterName,
		InstanceID: state.Instance.ID, TaskID: intent.TaskID,
		RecipientUserID: intent.RecipientUserID, Kind: string(intent.Kind),
	}
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
		StarterID: instance.StarterID, OperatorID: instance.OperatorID,
		Status: string(instance.Status), FormDataJSON: formDataJSON,
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
		if handledAt == 0 && (task.Status == workflowdomain.TaskStatusCompleted || task.Status == workflowdomain.TaskStatusApproved || task.Status == workflowdomain.TaskStatusRejected) {
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
		NodeTypes:        definitionNodeTypes(definition),
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
		StarterID: row.StarterID, OperatorID: row.OperatorID,
		Status: row.Status, StartTime: row.StartTime, EndTime: row.EndTime,
	}
}

func publishedDefinition(row workflowmodel.Definition, definition workflowcore.Definition, version int) application.PublishedDefinition {
	return application.PublishedDefinition{
		ID: row.ID, Key: row.Key, Name: row.Name, Description: row.Description,
		Category: row.Category, Version: version, Form: cloneDefinitionForm(definition),
		FieldPermissions: definitionFieldPermissions(definition), StartNodeID: definitionStartNodeID(definition),
		Initiator: definitionInitiator(definition), Availability: definitionStartAvailability(definition),
	}
}

func definitionInitiator(definition workflowcore.Definition) workflowcore.InitiatorConfig {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart && node.Initiator != nil {
			return workflowcore.InitiatorConfig{
				Scope:         node.Initiator.Scope,
				UserIDs:       append([]uint(nil), node.Initiator.UserIDs...),
				DepartmentIDs: append([]uint(nil), node.Initiator.DepartmentIDs...),
			}
		}
	}
	return workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}
}

func definitionStartAvailability(definition workflowcore.Definition) workflowcore.StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartAvailability(definition.Nodes[index].Availability)
		}
	}
	return workflowcore.DefaultStartAvailability()
}

func cloneDefinitionForm(definition workflowcore.Definition) []workflowcore.FormField {
	return cloneFormFields(definition.Form)
}

func cloneFormFields(fields []workflowcore.FormField) []workflowcore.FormField {
	if len(fields) == 0 {
		return nil
	}
	result := make([]workflowcore.FormField, 0, len(fields))
	for _, field := range fields {
		field.Options = cloneFormOptions(field.Options)
		if field.OptionSource != nil {
			optionSource := *field.OptionSource
			field.OptionSource = &optionSource
		}
		field.Columns = cloneFormFields(field.Columns)
		field.Fields = cloneFormFields(field.Fields)
		if field.Help != nil {
			help := *field.Help
			field.Help = &help
		}
		result = append(result, field)
	}
	return result
}

func cloneFormOptions(options []workflowcore.FormOption) []workflowcore.FormOption {
	if len(options) == 0 {
		return nil
	}
	result := make([]workflowcore.FormOption, 0, len(options))
	for _, option := range options {
		option.Children = cloneFormOptions(option.Children)
		result = append(result, option)
	}
	return result
}

func definitionFieldPermissions(definition workflowcore.Definition) map[string][]workflowcore.FieldPermission {
	result := make(map[string][]workflowcore.FieldPermission)
	for _, node := range definition.Nodes {
		if len(node.FormPermissions) == 0 {
			continue
		}
		permissions := make([]workflowcore.FieldPermission, 0, len(node.FormPermissions))
		for _, permission := range node.FormPermissions {
			permission.Actions = append([]string(nil), permission.Actions...)
			permissions = append(permissions, permission)
		}
		result[node.ID] = permissions
	}
	if result == nil {
		return map[string][]workflowcore.FieldPermission{}
	}
	return result
}

func definitionNodeTypes(definition workflowcore.Definition) map[string]string {
	result := make(map[string]string, len(definition.Nodes))
	for _, node := range definition.Nodes {
		result[node.ID] = node.Type
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
	if userID := strings.TrimSpace(query.ScopeUserID); userID != "" {
		switch query.Scope {
		case application.InstanceScopeStarted:
			db = db.Where("starter_id = ?", userID)
		case application.InstanceScopeHandled:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_process_tasks scope_task
				WHERE scope_task.instance_id = workflow_process_instances.id
				AND (scope_task.task_assignee_id = ? OR scope_task.handled_by = ?)
			)`, userID, userID)
		case application.InstanceScopeCopied:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_instance_participants scope_participant
				WHERE scope_participant.instance_id = workflow_process_instances.id
				AND scope_participant.user_id = ? AND scope_participant.participant_role = ?
			)`, userID, workflowmodel.ParticipantRoleCC)
		}
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
