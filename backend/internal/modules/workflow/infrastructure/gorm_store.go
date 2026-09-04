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
	"wecheckin/backend/internal/support/media"
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
		row.LogoURL = media.FullURLWithStaticDomainContext(ctx, row.LogoURL)
		definition, version, err := loadDefinitionVersion(db, row.ID, row.CurrentVersion)
		if err != nil {
			return nil, err
		}
		result = append(result, publishedDefinition(row, definition, version, publishedAssigneeLabels{}, false))
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
	row.LogoURL = media.FullURLWithStaticDomainContext(ctx, row.LogoURL)
	labels, err := loadPublishedAssigneeLabels(db, definition.Nodes)
	if err != nil {
		return nil, err
	}
	result := publishedDefinition(row, definition, version, labels, true)
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
			row := workflowmodel.NotificationOutbox{
				ID: outboxID, InstanceID: state.Instance.ID, NodeID: intent.NodeID, TaskID: intent.TaskID,
				RecipientUserID: intent.RecipientUserID, Kind: string(intent.Kind), Channel: channel,
				Status:      workflowmodel.NotificationStatusPending,
				DedupeKey:   notificationDedupeKey(state.Instance.ID, intent, channel),
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

func notificationDedupeKey(instanceID string, intent workflowdomain.NotificationIntent, channel string) string {
	taskKey := strings.TrimSpace(intent.TaskID)
	if taskKey == "" {
		taskKey = "-"
	}
	parts := []string{instanceID, string(intent.Kind), intent.NodeID, taskKey, intent.RecipientUserID, channel}
	if suffix := strings.TrimSpace(intent.DedupeKeySuffix); suffix != "" {
		parts = append(parts, suffix)
	}
	return strings.Join(parts, ":")
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
	if name := workflowUserName(user); name != "" {
		return name
	}
	return fallback
}

func workflowUserName(user model.User) string {
	if name := strings.TrimSpace(user.Name); name != "" {
		return name
	}
	return strings.TrimSpace(user.Account)
}

func renderNotificationPayload(state *workflowdomain.State, intent workflowdomain.NotificationIntent, starterName string) application.NotificationPayload {
	result := ""
	switch intent.Kind {
	case workflowdomain.NotificationKindApprovalResultApproved:
		result = "已通过"
	case workflowdomain.NotificationKindApprovalResultRejected:
		result = "已驳回"
	case workflowdomain.NotificationKindApprovalResultReturned:
		result = "已退回"
		if targetNodeName := strings.TrimSpace(intent.TargetNodeName); targetNodeName != "" {
			result += fmt.Sprintf("至“%s”", targetNodeName)
		}
	}
	replacements := map[string]string{
		"{{workflowName}}": intent.WorkflowName,
		"{{nodeName}}":     intent.NodeName,
		"{{starterName}}":  starterName,
		"{{instanceId}}":   state.Instance.ID,
		"{{taskId}}":       intent.TaskID,
		"{{result}}":       result,
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
	payload := application.NotificationPayload{
		Title: render(intent.Config.Title, 64), Content: render(intent.Config.Content, 1000),
		WorkflowName: intent.WorkflowName, NodeName: intent.NodeName,
		StarterID: state.Instance.StarterID, StarterName: starterName,
		InstanceID: state.Instance.ID, TaskID: intent.TaskID,
		RecipientUserID: intent.RecipientUserID, Kind: string(intent.Kind),
	}
	if intent.Kind == workflowdomain.NotificationKindInstanceCommented {
		payload.MessageType = application.NotificationMessageTypeActionCard
	}
	return payload
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
		StartTime: startTime, EndTime: endTime,
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
	labels, err := loadPublishedAssigneeLabels(db, definition.Nodes)
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
	instances, err := loadInstanceSummaries(db, []workflowmodel.ProcessInstance{instance})
	if err != nil {
		return nil, err
	}
	taskList, err := loadTaskSummaries(db, tasks)
	if err != nil {
		return nil, err
	}
	historyList, err := loadHistorySummaries(db, history)
	if err != nil {
		return nil, err
	}
	detail := &application.InstanceDetail{
		Instance: instances[0], Variables: make(map[string]interface{}, len(variables)),
		Form:             cloneDefinitionForm(definition),
		FormData:         make(map[string]interface{}),
		FieldPermissions: definitionFieldPermissions(definition),
		StartNodeID:      definitionStartNodeID(definition),
		NodeTypes:        definitionNodeTypes(definition),
		Tokens:           make([]application.TokenSummary, 0, len(tokens)),
		Tasks:            taskList,
		History:          historyList,
	}
	formData, err := decodeFormData(instance.FormDataJSON)
	if err != nil {
		return nil, err
	}
	detail.FormData = formData
	users, err := loadWorkflowUsers(db, collectInstanceUserIDs(instance, tasks, history, definition.Form, formData))
	if err != nil {
		return nil, err
	}
	detail.UserNames = workflowUserNames(users)
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
	populateInstanceGraph(detail, definition, labels)
	populateInstanceNodeProgress(detail, definition)
	return detail, nil
}

func populateInstanceGraph(detail *application.InstanceDetail, definition workflowcore.Definition, labels publishedAssigneeLabels) {
	detail.Nodes, detail.Edges = buildPublishedWorkflowGraph(definition, labels)
}

func populateInstanceNodeProgress(detail *application.InstanceDetail, definition workflowcore.Definition) {
	detail.NodeProgress = application.BuildNodeProgress(
		definition,
		detail.Instance,
		detail.Tokens,
		detail.Tasks,
		detail.History,
	)
}

func instanceSummary(row workflowmodel.ProcessInstance) application.InstanceSummary {
	return application.InstanceSummary{
		ID: row.ID, DefinitionID: row.DefinitionID, DefinitionVersion: row.DefinitionVersion,
		DefinitionKey: row.DefinitionKey, BusinessType: row.BusinessType, BusinessKey: row.BusinessKey,
		StarterID: row.StarterID, OperatorID: row.OperatorID,
		CurrentNodeNames: []string{}, CurrentAssigneeNames: []string{},
		Status: row.Status, StartTime: row.StartTime, EndTime: row.EndTime,
	}
}

func loadInstanceSummaries(db *gorm.DB, rows []workflowmodel.ProcessInstance) ([]application.InstanceSummary, error) {
	currentTasks, err := loadCurrentInstanceTasks(db, instanceRowIDs(rows))
	if err != nil {
		return nil, err
	}
	users, err := loadWorkflowUsers(db, instanceSummaryUserIDs(rows, currentTasks))
	if err != nil {
		return nil, err
	}
	definitionNames, err := loadWorkflowDefinitionNames(db, instanceDefinitionIDs(rows))
	if err != nil {
		return nil, err
	}
	return instanceSummariesWithCurrentTasks(rows, users, definitionNames, currentTasks), nil
}

func instanceRowIDs(rows []workflowmodel.ProcessInstance) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if id := strings.TrimSpace(row.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func loadCurrentInstanceTasks(db *gorm.DB, instanceIDs []string) ([]workflowmodel.ProcessTask, error) {
	tasks := make([]workflowmodel.ProcessTask, 0)
	if len(instanceIDs) == 0 {
		return tasks, nil
	}
	err := db.Select("id", "instance_id", "node_id", "node_name", "task_assignee_id", "task_status", "task_sequence", "created_at").
		Where("instance_id IN ? AND task_status = ?", instanceIDs, workflowmodel.TaskStatusPending).
		Order("created_at ASC").Order("task_sequence ASC").Order("id ASC").Find(&tasks).Error
	return tasks, err
}

func instanceSummaryUserIDs(rows []workflowmodel.ProcessInstance, tasks []workflowmodel.ProcessTask) []uint {
	ids := instanceUserIDs(rows)
	seen := make(map[uint]struct{}, len(ids)+len(tasks))
	for _, id := range ids {
		seen[id] = struct{}{}
	}
	for _, task := range tasks {
		appendWorkflowUserID(&ids, seen, task.AssigneeID)
	}
	return ids
}

func instanceDefinitionIDs(rows []workflowmodel.ProcessInstance) []uint {
	ids := make([]uint, 0, len(rows))
	seen := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		if row.DefinitionID == 0 {
			continue
		}
		if _, exists := seen[row.DefinitionID]; exists {
			continue
		}
		seen[row.DefinitionID] = struct{}{}
		ids = append(ids, row.DefinitionID)
	}
	return ids
}

func loadWorkflowDefinitionNames(db *gorm.DB, definitionIDs []uint) (map[uint]string, error) {
	names := make(map[uint]string, len(definitionIDs))
	if len(definitionIDs) == 0 {
		return names, nil
	}
	var definitions []workflowmodel.Definition
	if err := db.Select("id", "definition_name").Where("id IN ?", definitionIDs).Find(&definitions).Error; err != nil {
		return nil, err
	}
	for _, definition := range definitions {
		names[definition.ID] = strings.TrimSpace(definition.Name)
	}
	return names, nil
}

func instanceUserIDs(rows []workflowmodel.ProcessInstance) []uint {
	ids := make([]uint, 0, len(rows)*2)
	seen := make(map[uint]struct{}, len(rows)*2)
	for _, row := range rows {
		appendWorkflowUserID(&ids, seen, row.StarterID)
		appendWorkflowUserID(&ids, seen, row.OperatorID)
	}
	return ids
}

func instanceStarterUserIDs(rows []workflowmodel.ProcessInstance) []uint {
	ids := make([]uint, 0, len(rows))
	seen := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		id, err := strconv.ParseUint(strings.TrimSpace(row.StarterID), 10, 64)
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
	return ids
}

func instanceSummaries(rows []workflowmodel.ProcessInstance, users []model.User) []application.InstanceSummary {
	return instanceSummariesWithDefinitionNames(rows, users, nil)
}

func instanceSummariesWithDefinitionNames(rows []workflowmodel.ProcessInstance, users []model.User, definitionNames map[uint]string) []application.InstanceSummary {
	return instanceSummariesWithCurrentTasks(rows, users, definitionNames, nil)
}

func instanceSummariesWithCurrentTasks(rows []workflowmodel.ProcessInstance, users []model.User, definitionNames map[uint]string, tasks []workflowmodel.ProcessTask) []application.InstanceSummary {
	names := workflowUserNames(users)
	nodeNamesByInstance := make(map[string][]string)
	assigneeNamesByInstance := make(map[string][]string)
	nodeIDsByInstance := make(map[string]map[string]struct{})
	assigneeIDsByInstance := make(map[string]map[string]struct{})
	for _, task := range tasks {
		if task.Status != workflowmodel.TaskStatusPending {
			continue
		}
		instanceID := strings.TrimSpace(task.InstanceID)
		if instanceID == "" {
			continue
		}
		if nodeIDsByInstance[instanceID] == nil {
			nodeIDsByInstance[instanceID] = make(map[string]struct{})
		}
		nodeID := strings.TrimSpace(task.NodeID)
		if _, exists := nodeIDsByInstance[instanceID][nodeID]; !exists {
			nodeName := strings.TrimSpace(task.NodeName)
			if nodeName == "" {
				nodeName = nodeID
			}
			if nodeName != "" {
				nodeIDsByInstance[instanceID][nodeID] = struct{}{}
				nodeNamesByInstance[instanceID] = append(nodeNamesByInstance[instanceID], nodeName)
			}
		}
		assigneeID := strings.TrimSpace(task.AssigneeID)
		assigneeName := names[assigneeID]
		if assigneeName == "" {
			continue
		}
		if assigneeIDsByInstance[instanceID] == nil {
			assigneeIDsByInstance[instanceID] = make(map[string]struct{})
		}
		if _, exists := assigneeIDsByInstance[instanceID][assigneeID]; exists {
			continue
		}
		assigneeIDsByInstance[instanceID][assigneeID] = struct{}{}
		assigneeNamesByInstance[instanceID] = append(assigneeNamesByInstance[instanceID], assigneeName)
	}

	list := make([]application.InstanceSummary, 0, len(rows))
	for _, row := range rows {
		summary := instanceSummary(row)
		summary.DefinitionName = definitionNames[summary.DefinitionID]
		summary.StarterName = names[summary.StarterID]
		summary.OperatorName = names[summary.OperatorID]
		if summary.Status == workflowmodel.InstanceStatusRunning {
			summary.CurrentNodeNames = append(summary.CurrentNodeNames, nodeNamesByInstance[summary.ID]...)
			summary.CurrentAssigneeNames = append(summary.CurrentAssigneeNames, assigneeNamesByInstance[summary.ID]...)
		}
		list = append(list, summary)
	}
	return list
}

func publishedDefinition(row workflowmodel.Definition, definition workflowcore.Definition, version int, labels publishedAssigneeLabels, includeGraph bool) application.PublishedDefinition {
	result := application.PublishedDefinition{
		ID: row.ID, Key: row.Key, Name: row.Name, Description: row.Description,
		Category: row.Category, LogoURL: row.LogoURL, Version: version, Form: cloneDefinitionForm(definition),
		FieldPermissions: definitionFieldPermissions(definition), StartNodeID: definitionStartNodeID(definition),
		Initiator: definitionInitiator(definition), Availability: definitionStartAvailability(definition),
		StartLimit: definitionStartLimit(definition), StartLimitStatus: application.StartLimitStatus{Allowed: true},
	}
	if includeGraph {
		result.Nodes, result.Edges = buildPublishedWorkflowGraph(definition, labels)
	}
	return result
}

func definitionInitiator(definition workflowcore.Definition) workflowcore.InitiatorConfig {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart && node.Initiator != nil {
			return workflowcore.InitiatorConfig{
				Scope:           node.Initiator.Scope,
				UserIDs:         append([]uint(nil), node.Initiator.UserIDs...),
				DepartmentIDs:   append([]uint(nil), node.Initiator.DepartmentIDs...),
				ExcludedUserIDs: append([]uint(nil), node.Initiator.ExcludedUserIDs...),
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

func definitionStartLimit(definition workflowcore.Definition) workflowcore.StartLimitConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartLimit(definition.Nodes[index].StartLimit)
		}
	}
	return workflowcore.DefaultStartLimit()
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

func collectInstanceUserIDs(
	instance workflowmodel.ProcessInstance,
	tasks []workflowmodel.ProcessTask,
	history []workflowmodel.ProcessHistory,
	fields []workflowcore.FormField,
	formData map[string]interface{},
) []uint {
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})
	appendWorkflowUserID(&ids, seen, instance.StarterID)
	appendWorkflowUserID(&ids, seen, instance.OperatorID)
	for _, task := range tasks {
		appendWorkflowUserID(&ids, seen, task.AssigneeID)
		appendWorkflowUserID(&ids, seen, task.HandledBy)
	}
	for _, event := range history {
		appendWorkflowUserID(&ids, seen, event.ActorID)
	}
	collectFormUserIDs(&ids, seen, fields, formData)
	return ids
}

func collectFormUserIDs(ids *[]uint, seen map[uint]struct{}, fields []workflowcore.FormField, data map[string]interface{}) {
	for _, field := range fields {
		if field.Type == workflowcore.FormFieldTypeGroup {
			collectFormUserIDs(ids, seen, field.Fields, data)
			continue
		}
		value := data[field.Key]
		switch field.Type {
		case workflowcore.FormFieldTypeUser:
			appendWorkflowUserID(ids, seen, stringFormValue(value))
		case workflowcore.FormFieldTypeUserMulti:
			for _, item := range stringSliceFormValue(value) {
				appendWorkflowUserID(ids, seen, item)
			}
		case workflowcore.FormFieldTypeDetailList:
			for _, row := range detailFormRows(value) {
				collectFormUserIDs(ids, seen, field.Columns, row)
			}
		}
	}
}

func appendWorkflowUserID(ids *[]uint, seen map[uint]struct{}, value string) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return
	}
	id := uint(parsed)
	if _, exists := seen[id]; exists {
		return
	}
	seen[id] = struct{}{}
	*ids = append(*ids, id)
}

func stringFormValue(value interface{}) string {
	text, _ := value.(string)
	return text
}

func stringSliceFormValue(value interface{}) []string {
	switch values := value.(type) {
	case []string:
		return values
	case []interface{}:
		result := make([]string, 0, len(values))
		for _, item := range values {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func detailFormRows(value interface{}) []map[string]interface{} {
	switch rows := value.(type) {
	case []map[string]interface{}:
		return rows
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(rows))
		for _, item := range rows {
			if row, ok := item.(map[string]interface{}); ok {
				result = append(result, row)
			}
		}
		return result
	default:
		return nil
	}
}

func applyInstanceFilters(db *gorm.DB, query application.InstanceQuery) *gorm.DB {
	db = db.Where("admin_deleted_at = 0")
	if query.DefinitionID > 0 {
		db = db.Where("definition_id = ?", query.DefinitionID)
	}
	if query.DefinitionVersion > 0 {
		db = db.Where("definition_version = ?", query.DefinitionVersion)
	}
	if len(query.InstanceIDs) > 0 {
		db = db.Where("id IN ?", query.InstanceIDs)
	}
	if value := strings.TrimSpace(query.DefinitionName); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM workflow_definitions name_definition
			WHERE name_definition.id = workflow_process_instances.definition_id
			AND name_definition.definition_name LIKE ? ESCAPE '!'
		)`, containsLikePattern(value))
	}
	if value := strings.TrimSpace(query.DefinitionCategory); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM workflow_definitions scope_definition
			WHERE scope_definition.id = workflow_process_instances.definition_id
			AND scope_definition.definition_category = ?
		)`, value)
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
	if value := strings.TrimSpace(query.StarterName); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM users starter_user
			WHERE starter_user.id = CAST(workflow_process_instances.starter_id AS UNSIGNED)
			AND starter_user.user_name LIKE ? ESCAPE '!'
		)`, containsLikePattern(value))
	}
	if query.StartTimeFrom > 0 {
		db = db.Where("start_time >= ?", query.StartTimeFrom)
	}
	if query.StartTimeTo > 0 {
		db = db.Where("start_time <= ?", query.StartTimeTo)
	}
	if query.EndTimeFrom > 0 {
		db = db.Where("end_time >= ?", query.EndTimeFrom)
	}
	if query.EndTimeTo > 0 {
		db = db.Where("end_time <= ?", query.EndTimeTo)
	}
	if userID := strings.TrimSpace(query.ScopeUserID); userID != "" {
		switch query.Scope {
		case application.InstanceScopeStarted:
			db = db.Where("starter_id = ? AND starter_deleted_at = ?", userID, int64(0))
		case application.InstanceScopeHandled:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_process_tasks scope_task
				WHERE scope_task.instance_id = workflow_process_instances.id
				AND scope_task.task_status IN (?, ?, ?, ?)
				AND (scope_task.task_assignee_id = ? OR scope_task.handled_by = ?)
			)`,
				workflowmodel.TaskStatusCompleted,
				workflowmodel.TaskStatusApproved,
				workflowmodel.TaskStatusRejected,
				workflowmodel.TaskStatusReturned,
				userID,
				userID,
			)
		case application.InstanceScopeCopied:
			db = db.Where(`EXISTS (
				SELECT 1 FROM workflow_instance_participants scope_participant
				WHERE scope_participant.instance_id = workflow_process_instances.id
				AND scope_participant.user_id = ? AND scope_participant.participant_role = ?
			)`, userID, workflowmodel.ParticipantRoleCC)
		}
	}
	if where, args := instanceVisibilityWhere(query.Visibility); where != "" {
		db = db.Where(where, args...)
	}
	return db
}

func instanceVisibilityWhere(visibility *application.InstanceVisibility) (string, []interface{}) {
	if visibility == nil || (visibility.Ready && visibility.All) {
		return "", nil
	}
	if !visibility.Ready {
		return "1 = 0", nil
	}
	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if userIDs := normalizeVisibilityUintIDs(visibility.UserIDs); len(userIDs) > 0 {
		clauses = append(clauses, "CAST(starter_id AS UNSIGNED) IN ?")
		args = append(args, userIDs)
	}
	if departmentIDs := normalizeVisibilityUintIDs(visibility.DepartmentIDs); len(departmentIDs) > 0 {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM user_depts summary_user_dept
			WHERE summary_user_dept.user_dept_user_id = CAST(workflow_process_instances.starter_id AS UNSIGNED)
			AND summary_user_dept.user_dept_dept_id IN ?
		)`)
		args = append(args, departmentIDs)
	}
	if len(clauses) == 0 {
		return "1 = 0", nil
	}
	return "(" + strings.Join(clauses, " OR ") + ")", args
}

func normalizeVisibilityUintIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func applyTaskFilters(db *gorm.DB, query application.TaskQuery) *gorm.DB {
	if query.HideAdminDeleted {
		db = db.Where("workflow_process_tasks.admin_deleted_at = ?", int64(0))
	}
	db = db.Where(`EXISTS (
		SELECT 1 FROM workflow_process_instances active_instance
		WHERE active_instance.id = workflow_process_tasks.instance_id
		AND active_instance.admin_deleted_at = 0
	)`)
	if value := strings.TrimSpace(query.InstanceID); value != "" {
		db = db.Where("instance_id = ?", value)
	}
	if value := strings.TrimSpace(query.AssigneeID); value != "" {
		db = db.Where("task_assignee_id = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		db = db.Where("task_status = ?", value)
	}
	if hasTaskInstanceFilters(query) {
		instances := db.Session(&gorm.Session{NewDB: true}).
			Model(&workflowmodel.ProcessInstance{}).
			Select("id")
		instances = applyInstanceFilters(instances, application.InstanceQuery{
			DefinitionName:     query.DefinitionName,
			DefinitionCategory: query.DefinitionCategory,
			StarterName:        query.StarterName,
			StartTimeFrom:      query.StartTimeFrom,
			StartTimeTo:        query.StartTimeTo,
		})
		db = db.Where("instance_id IN (?)", instances)
	}
	return db
}

func hasTaskInstanceFilters(query application.TaskQuery) bool {
	return strings.TrimSpace(query.DefinitionName) != "" ||
		strings.TrimSpace(query.DefinitionCategory) != "" ||
		strings.TrimSpace(query.StarterName) != "" ||
		query.StartTimeFrom > 0 || query.StartTimeTo > 0
}

func containsLikePattern(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return "%" + replacer.Replace(strings.TrimSpace(value)) + "%"
}
