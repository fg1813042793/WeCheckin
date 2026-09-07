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
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
)

func (store *GormStore) LoadCompletedRevisionSourceForUpdate(ctx context.Context, instanceID string) (workflowcore.Definition, *workflowdomain.State, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return workflowcore.Definition{}, nil, err
	}
	defer cancel()

	var instance workflowmodel.ProcessInstance
	query := lockCompletedRevisionSourceQuery(db, instanceID, &instance)
	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return workflowcore.Definition{}, nil, ErrInstanceNotFound
		}
		return workflowcore.Definition{}, nil, query.Error
	}
	definition, _, err := loadDefinitionVersion(db, instance.DefinitionID, instance.DefinitionVersion)
	if err != nil {
		return workflowcore.Definition{}, nil, err
	}
	state, err := loadStateRecords(db, instance)
	return definition, state, err
}

func lockCompletedRevisionSourceQuery(db *gorm.DB, instanceID string, destinations ...*workflowmodel.ProcessInstance) *gorm.DB {
	row := &workflowmodel.ProcessInstance{}
	if len(destinations) > 0 && destinations[0] != nil {
		row = destinations[0]
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).First(row, "id = ?", strings.TrimSpace(instanceID))
}

func (store *GormStore) LoadActiveFormRevisionForUpdate(ctx context.Context, instanceID string) (*workflowdomain.FormRevisionRequest, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var row workflowmodel.FormRevisionRequest
	query := lockActiveFormRevisionQuery(db, instanceID, &row)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if query.Error != nil {
		return nil, query.Error
	}
	return loadFormRevisionTasks(db, row, true)
}

func lockActiveFormRevisionQuery(db *gorm.DB, instanceID string, destinations ...*workflowmodel.FormRevisionRequest) *gorm.DB {
	row := &workflowmodel.FormRevisionRequest{}
	if len(destinations) > 0 && destinations[0] != nil {
		row = destinations[0]
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE"}).First(row, "active_lock_key = ?", strings.TrimSpace(instanceID))
}

func (store *GormStore) LoadFormRevisionForUpdate(ctx context.Context, revisionID string) (*workflowdomain.FormRevisionRequest, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	return loadFormRevisionForUpdate(db, revisionID)
}

func (store *GormStore) LoadFormRevisionByTaskForUpdate(ctx context.Context, taskID string) (*workflowdomain.FormRevisionRequest, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var task workflowmodel.FormRevisionTask
	if err := db.Select("revision_request_id").First(&task, "id = ?", strings.TrimSpace(taskID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFormRevisionTaskNotFound
		}
		return nil, err
	}
	return loadFormRevisionForUpdate(db, task.RevisionRequestID)
}

func loadFormRevisionForUpdate(db *gorm.DB, revisionID string) (*workflowdomain.FormRevisionRequest, error) {
	var row workflowmodel.FormRevisionRequest
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", strings.TrimSpace(revisionID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFormRevisionNotFound
		}
		return nil, err
	}
	return loadFormRevisionTasks(db, row, true)
}

func loadFormRevisionTasks(db *gorm.DB, row workflowmodel.FormRevisionRequest, lock bool) (*workflowdomain.FormRevisionRequest, error) {
	query := db.Where("revision_request_id = ?", row.ID).Order("stage ASC").Order("node_id ASC").Order("task_sequence ASC").Order("id ASC")
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var tasks []workflowmodel.FormRevisionTask
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return formRevisionFromModels(row, tasks)
}

func (store *GormStore) CreateFormRevision(ctx context.Context, revision *workflowdomain.FormRevisionRequest) error {
	requestRow, taskRows, err := formRevisionModels(revision)
	if err != nil {
		return err
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err := db.Create(&requestRow).Error; err != nil {
		return err
	}
	if len(taskRows) == 0 {
		return nil
	}
	return db.Create(&taskRows).Error
}

func (store *GormStore) SaveFormRevision(ctx context.Context, revision *workflowdomain.FormRevisionRequest) error {
	requestRow, taskRows, err := formRevisionModels(revision)
	if err != nil {
		return err
	}
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	updates := map[string]interface{}{
		"revision_status":       requestRow.Status,
		"active_lock_key":       requestRow.ActiveLockKey,
		"applied_form_revision": requestRow.AppliedFormRevision,
		"edit_time":             requestRow.EditTime,
		"completed_at":          requestRow.CompletedAt,
		"updated_at":            database.Now(),
	}
	result := db.Model(&workflowmodel.FormRevisionRequest{}).Where("id = ?", requestRow.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFormRevisionNotFound
	}
	for _, task := range taskRows {
		updates := map[string]interface{}{
			"task_status": task.Status, "task_action": task.Action, "task_comment": task.Comment,
			"task_images_json": task.ImagesJSON, "handled_by": task.HandledBy, "handled_at": task.HandledAt,
			"edit_time": task.EditTime, "updated_at": database.Now(),
		}
		if err := db.Model(&workflowmodel.FormRevisionTask{}).Where("id = ? AND revision_request_id = ?", task.ID, revision.ID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func (store *GormStore) ListFormRevisions(ctx context.Context, instanceID, actorID string) ([]application.FormRevisionSummary, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	query := db.Where("source_instance_id = ?", strings.TrimSpace(instanceID))
	query = applyFormRevisionViewer(query, strings.TrimSpace(actorID))
	var rows []workflowmodel.FormRevisionRequest
	if err := query.Order("add_time DESC").Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]application.FormRevisionSummary, 0, len(rows))
	for _, row := range rows {
		summary, err := formRevisionSummaryFromModel(row)
		if err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	return result, nil
}

func (store *GormStore) GetFormRevision(ctx context.Context, revisionID, actorID string) (*application.FormRevisionDetail, error) {
	db, cancel, err := store.contextDB(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	query := applyFormRevisionViewer(db, strings.TrimSpace(actorID))
	var row workflowmodel.FormRevisionRequest
	if err := query.First(&row, "id = ?", strings.TrimSpace(revisionID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFormRevisionNotFound
		}
		return nil, err
	}
	revision, err := loadFormRevisionTasks(db, row, false)
	if err != nil {
		return nil, err
	}
	return formRevisionDetail(revision), nil
}

func applyFormRevisionViewer(query *gorm.DB, actorID string) *gorm.DB {
	if actorID == "" {
		return query
	}
	return query.Where(`(
		requester_id = ?
		OR EXISTS (SELECT 1 FROM workflow_form_revision_tasks frt WHERE frt.revision_request_id = workflow_form_revision_requests.id AND (frt.assignee_id = ? OR frt.handled_by = ?))
		OR EXISTS (SELECT 1 FROM workflow_process_instances wpi WHERE wpi.id = workflow_form_revision_requests.source_instance_id AND (wpi.starter_id = ? OR wpi.operator_id = ?))
		OR EXISTS (SELECT 1 FROM workflow_process_tasks wpt WHERE wpt.instance_id = workflow_form_revision_requests.source_instance_id AND (wpt.task_assignee_id = ? OR wpt.handled_by = ?))
		OR EXISTS (SELECT 1 FROM workflow_instance_participants wip WHERE wip.instance_id = workflow_form_revision_requests.source_instance_id AND wip.user_id = ?)
	)`, actorID, actorID, actorID, actorID, actorID, actorID, actorID, actorID)
}

func formRevisionModels(revision *workflowdomain.FormRevisionRequest) (workflowmodel.FormRevisionRequest, []workflowmodel.FormRevisionTask, error) {
	if revision == nil {
		return workflowmodel.FormRevisionRequest{}, nil, errors.New("表单修订请求不能为空")
	}
	beforeJSON, err := encodeRevisionJSON(revision.BeforeFormData, map[string]interface{}{})
	if err != nil {
		return workflowmodel.FormRevisionRequest{}, nil, fmt.Errorf("序列化修订前表单失败: %w", err)
	}
	patchJSON, err := encodeRevisionJSON(revision.Patch, map[string]interface{}{})
	if err != nil {
		return workflowmodel.FormRevisionRequest{}, nil, fmt.Errorf("序列化修订补丁失败: %w", err)
	}
	proposedJSON, err := encodeRevisionJSON(revision.ProposedFormData, map[string]interface{}{})
	if err != nil {
		return workflowmodel.FormRevisionRequest{}, nil, fmt.Errorf("序列化候选表单失败: %w", err)
	}
	labelsJSON, err := encodeRevisionJSON(revision.ChangedFieldLabels, []string{})
	if err != nil {
		return workflowmodel.FormRevisionRequest{}, nil, fmt.Errorf("序列化修改字段名称失败: %w", err)
	}
	row := workflowmodel.FormRevisionRequest{
		ID: revision.ID, SourceInstanceID: revision.SourceInstanceID,
		SourceNodeID: revision.SourceNodeID, SourceNodeName: revision.SourceNodeName,
		RequesterID: revision.RequesterID, BaseFormRevision: revision.BaseFormRevision,
		BeforeFormDataJSON: beforeJSON, PatchJSON: patchJSON, ProposedFormDataJSON: proposedJSON,
		ChangedFieldLabelsJSON: labelsJSON, Reason: revision.Reason, RevisionMode: revision.Mode,
		Status: string(revision.Status), ActiveLockKey: activeFormRevisionLock(revision),
		AppliedFormRevision: revision.AppliedFormRevision, AddTime: revision.CreatedAt,
		EditTime: revisionEditTime(revision), CompletedAt: revision.CompletedAt,
	}
	tasks := make([]workflowmodel.FormRevisionTask, 0, len(revision.Tasks))
	for _, task := range revision.Tasks {
		tasks = append(tasks, workflowmodel.FormRevisionTask{
			ID: task.ID, RevisionRequestID: revision.ID, SourceTaskID: task.SourceTaskID,
			NodeID: task.NodeID, NodeName: task.NodeName, Stage: task.Stage,
			AssigneeID: task.AssigneeID, AssigneeName: task.AssigneeName,
			ApprovalMode: task.ApprovalMode, CompletionRate: task.CompletionRate,
			Sequence: task.Sequence, Total: task.Total, Status: string(task.Status), Action: string(task.Action),
			Comment: task.Comment, ImagesJSON: encodeWorkflowImages(task.Images),
			HandledBy: task.HandledBy, HandledAt: task.HandledAt,
			AddTime: revision.CreatedAt, EditTime: revisionEditTime(revision),
		})
	}
	return row, tasks, nil
}

func formRevisionFromModels(row workflowmodel.FormRevisionRequest, taskRows []workflowmodel.FormRevisionTask) (*workflowdomain.FormRevisionRequest, error) {
	before, err := decodeFormData(row.BeforeFormDataJSON)
	if err != nil {
		return nil, fmt.Errorf("解析修订前表单失败: %w", err)
	}
	patch, err := decodeFormData(row.PatchJSON)
	if err != nil {
		return nil, fmt.Errorf("解析修订补丁失败: %w", err)
	}
	proposed, err := decodeFormData(row.ProposedFormDataJSON)
	if err != nil {
		return nil, fmt.Errorf("解析候选表单失败: %w", err)
	}
	labels := make([]string, 0)
	if strings.TrimSpace(row.ChangedFieldLabelsJSON) != "" {
		if err := json.Unmarshal([]byte(row.ChangedFieldLabelsJSON), &labels); err != nil {
			return nil, fmt.Errorf("解析修改字段名称失败: %w", err)
		}
	}
	revision := &workflowdomain.FormRevisionRequest{
		ID: row.ID, SourceInstanceID: row.SourceInstanceID, SourceNodeID: row.SourceNodeID,
		SourceNodeName: row.SourceNodeName, RequesterID: row.RequesterID,
		BaseFormRevision: row.BaseFormRevision, AppliedFormRevision: row.AppliedFormRevision,
		BeforeFormData: before, Patch: patch, ProposedFormData: proposed,
		ChangedFieldLabels: labels, Reason: row.Reason, Mode: row.RevisionMode,
		Status: workflowdomain.FormRevisionStatus(row.Status), CreatedAt: row.AddTime, CompletedAt: row.CompletedAt,
		Tasks: make([]workflowdomain.FormRevisionTask, 0, len(taskRows)),
	}
	for _, taskRow := range taskRows {
		images, err := decodeWorkflowImages(taskRow.ImagesJSON)
		if err != nil {
			return nil, fmt.Errorf("解析修订任务图片失败: %w", err)
		}
		revision.Tasks = append(revision.Tasks, workflowdomain.FormRevisionTask{
			ID: taskRow.ID, RevisionRequestID: taskRow.RevisionRequestID, SourceTaskID: taskRow.SourceTaskID,
			NodeID: taskRow.NodeID, NodeName: taskRow.NodeName, Stage: taskRow.Stage,
			AssigneeID: taskRow.AssigneeID, AssigneeName: taskRow.AssigneeName,
			ApprovalMode: taskRow.ApprovalMode, CompletionRate: taskRow.CompletionRate,
			Sequence: taskRow.Sequence, Total: taskRow.Total,
			Status: workflowdomain.RevisionTaskStatus(taskRow.Status), Action: workflowdomain.RevisionAction(taskRow.Action),
			Comment: taskRow.Comment, Images: images, HandledBy: taskRow.HandledBy, HandledAt: taskRow.HandledAt,
		})
	}
	return revision, nil
}

func formRevisionSummaryFromModel(row workflowmodel.FormRevisionRequest) (application.FormRevisionSummary, error) {
	labels := make([]string, 0)
	if strings.TrimSpace(row.ChangedFieldLabelsJSON) != "" {
		if err := json.Unmarshal([]byte(row.ChangedFieldLabelsJSON), &labels); err != nil {
			return application.FormRevisionSummary{}, err
		}
	}
	return application.FormRevisionSummary{
		ID: row.ID, SourceInstanceID: row.SourceInstanceID, SourceNodeID: row.SourceNodeID,
		SourceNodeName: row.SourceNodeName, RequesterID: row.RequesterID,
		BaseFormRevision: row.BaseFormRevision, AppliedFormRevision: row.AppliedFormRevision,
		ChangedFieldLabels: labels, Reason: row.Reason, Mode: row.RevisionMode, Status: row.Status,
		CreatedAt: row.AddTime, CompletedAt: row.CompletedAt,
	}, nil
}

func formRevisionDetail(revision *workflowdomain.FormRevisionRequest) *application.FormRevisionDetail {
	if revision == nil {
		return nil
	}
	detail := &application.FormRevisionDetail{
		FormRevisionSummary: application.FormRevisionSummary{
			ID: revision.ID, SourceInstanceID: revision.SourceInstanceID,
			SourceNodeID: revision.SourceNodeID, SourceNodeName: revision.SourceNodeName,
			RequesterID: revision.RequesterID, BaseFormRevision: revision.BaseFormRevision,
			AppliedFormRevision: revision.AppliedFormRevision,
			ChangedFieldLabels:  append([]string(nil), revision.ChangedFieldLabels...),
			Reason:              revision.Reason, Mode: revision.Mode, Status: string(revision.Status),
			CreatedAt: revision.CreatedAt, CompletedAt: revision.CompletedAt,
		},
		BeforeFormData:   workflowcore.MergeFormData(nil, revision.BeforeFormData),
		Patch:            workflowcore.MergeFormData(nil, revision.Patch),
		ProposedFormData: workflowcore.MergeFormData(nil, revision.ProposedFormData),
		Tasks:            make([]application.FormRevisionTaskSummary, 0, len(revision.Tasks)),
	}
	for _, task := range revision.Tasks {
		detail.Tasks = append(detail.Tasks, application.FormRevisionTaskSummary{
			ID: task.ID, SourceTaskID: task.SourceTaskID, NodeID: task.NodeID, NodeName: task.NodeName,
			Stage: task.Stage, AssigneeID: task.AssigneeID, AssigneeName: task.AssigneeName,
			ApprovalMode: task.ApprovalMode, CompletionRate: task.CompletionRate,
			Sequence: task.Sequence, Total: task.Total, Status: string(task.Status), Action: string(task.Action),
			Comment: task.Comment, Images: append([]workflowcore.FormAttachment(nil), task.Images...),
			HandledBy: task.HandledBy, HandledAt: task.HandledAt,
		})
	}
	return detail
}

func encodeRevisionJSON(value, empty interface{}) (string, error) {
	if value == nil {
		value = empty
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func revisionEditTime(revision *workflowdomain.FormRevisionRequest) int64 {
	if revision.CompletedAt > 0 {
		return revision.CompletedAt
	}
	return revision.CreatedAt
}

func activeFormRevisionLock(revision *workflowdomain.FormRevisionRequest) *string {
	if revision.Status != workflowdomain.FormRevisionStatusPending {
		return nil
	}
	value := strings.TrimSpace(revision.SourceInstanceID)
	if value == "" {
		return nil
	}
	return &value
}
