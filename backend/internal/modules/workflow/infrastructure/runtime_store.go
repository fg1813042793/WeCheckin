package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"strings"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
)

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
		"form_revision":   state.Instance.FormRevision,
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
