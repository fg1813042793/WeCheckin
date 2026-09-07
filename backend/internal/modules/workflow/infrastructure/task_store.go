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
	if query.IncludeFormRevisions {
		return listUnifiedTasks(db, query)
	}
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

type unifiedTaskRow struct {
	ID                   string `gorm:"column:id"`
	InstanceID           string `gorm:"column:instance_id"`
	NodeID               string `gorm:"column:node_id"`
	NodeName             string `gorm:"column:node_name"`
	AssigneeID           string `gorm:"column:assignee_id"`
	ApprovalMode         string `gorm:"column:approval_mode"`
	CompletionRate       int    `gorm:"column:completion_rate"`
	Sequence             int    `gorm:"column:task_sequence"`
	Total                int    `gorm:"column:task_total"`
	ApprovalChainKey     string `gorm:"column:approval_chain_key"`
	ApprovalLayer        int    `gorm:"column:approval_layer"`
	ApprovalLayerTotal   int    `gorm:"column:approval_layer_total"`
	SourceDepartmentID   uint   `gorm:"column:source_department_id"`
	SourceDepartmentName string `gorm:"column:source_department_name"`
	Status               string `gorm:"column:task_status"`
	Action               string `gorm:"column:task_action"`
	Comment              string `gorm:"column:task_comment"`
	ImagesJSON           string `gorm:"column:task_images_json"`
	HandledBy            string `gorm:"column:handled_by"`
	HandledAt            int64  `gorm:"column:handled_at"`
	TaskType             string `gorm:"column:task_type"`
	RevisionRequestID    string `gorm:"column:revision_request_id"`
	SortTime             int64  `gorm:"column:sort_time"`
}

const unifiedWorkflowTaskProjection = `
	workflow_task.id AS id,
	workflow_task.instance_id AS instance_id,
	workflow_task.node_id AS node_id,
	workflow_task.node_name AS node_name,
	workflow_task.task_assignee_id AS assignee_id,
	workflow_task.approval_mode AS approval_mode,
	workflow_task.completion_rate AS completion_rate,
	workflow_task.task_sequence AS task_sequence,
	workflow_task.task_total AS task_total,
	workflow_task.approval_chain_key AS approval_chain_key,
	workflow_task.approval_layer AS approval_layer,
	workflow_task.approval_layer_total AS approval_layer_total,
	workflow_task.source_department_id AS source_department_id,
	workflow_task.source_department_name AS source_department_name,
	workflow_task.task_status AS task_status,
	workflow_task.task_action AS task_action,
	workflow_task.task_comment AS task_comment,
	workflow_task.task_images_json AS task_images_json,
	workflow_task.handled_by AS handled_by,
	workflow_task.handled_at AS handled_at,
	'workflow' AS task_type,
	'' AS revision_request_id,
	CAST(COALESCE(UNIX_TIMESTAMP(workflow_task.created_at) * 1000, 0) AS SIGNED) AS sort_time`

const unifiedRevisionTaskProjection = `
	revision_task.id AS id,
	revision_request.source_instance_id AS instance_id,
	revision_task.node_id AS node_id,
	revision_task.node_name AS node_name,
	revision_task.assignee_id AS assignee_id,
	revision_task.approval_mode AS approval_mode,
	revision_task.completion_rate AS completion_rate,
	revision_task.task_sequence AS task_sequence,
	revision_task.task_total AS task_total,
	'' AS approval_chain_key,
	0 AS approval_layer,
	0 AS approval_layer_total,
	0 AS source_department_id,
	'' AS source_department_name,
	revision_task.task_status AS task_status,
	revision_task.task_action AS task_action,
	revision_task.task_comment AS task_comment,
	revision_task.task_images_json AS task_images_json,
	revision_task.handled_by AS handled_by,
	revision_task.handled_at AS handled_at,
	'form_revision' AS task_type,
	revision_task.revision_request_id AS revision_request_id,
	revision_task.add_time AS sort_time`

func listUnifiedTasks(db *gorm.DB, query application.TaskQuery) (*application.TaskList, error) {
	unionSQL, args := unifiedTaskUnionSQL(query)
	var total int64
	if err := db.Raw("SELECT COUNT(*) FROM ("+unionSQL+") unified_task", args...).Scan(&total).Error; err != nil {
		return nil, err
	}
	listSQL, listArgs := unifiedTaskListSQL(query)
	var rows []unifiedTaskRow
	if err := db.Raw(listSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list, err := loadUnifiedTaskSummaries(db, rows)
	if err != nil {
		return nil, err
	}
	return &application.TaskList{List: list, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
}

func unifiedTaskListSQL(query application.TaskQuery) (string, []interface{}) {
	unionSQL, args := unifiedTaskUnionSQL(query)
	offset := (query.Page - 1) * query.PageSize
	args = append(args, query.PageSize, offset)
	return "SELECT unified_task.* FROM (" + unionSQL + ") unified_task " +
		"ORDER BY unified_task.sort_time DESC, unified_task.id DESC LIMIT ? OFFSET ?", args
}

func unifiedTaskUnionSQL(query application.TaskQuery) (string, []interface{}) {
	workflowWhere, workflowArgs := unifiedTaskBranchWhere(query, false)
	revisionWhere, revisionArgs := unifiedTaskBranchWhere(query, true)
	querySQL := "SELECT " + unifiedWorkflowTaskProjection + `
		FROM workflow_process_tasks workflow_task
		INNER JOIN workflow_process_instances task_instance ON task_instance.id = workflow_task.instance_id
		LEFT JOIN workflow_definitions task_definition ON task_definition.id = task_instance.definition_id
		WHERE ` + workflowWhere + `
		UNION ALL
		SELECT ` + unifiedRevisionTaskProjection + `
		FROM workflow_form_revision_tasks revision_task
		INNER JOIN workflow_form_revision_requests revision_request ON revision_request.id = revision_task.revision_request_id
		INNER JOIN workflow_process_instances task_instance ON task_instance.id = revision_request.source_instance_id
		LEFT JOIN workflow_definitions task_definition ON task_definition.id = task_instance.definition_id
		WHERE ` + revisionWhere
	return querySQL, append(workflowArgs, revisionArgs...)
}

func unifiedTaskBranchWhere(query application.TaskQuery, revision bool) (string, []interface{}) {
	taskAlias := "workflow_task"
	instanceIDColumn := taskAlias + ".instance_id"
	assigneeColumn := taskAlias + ".task_assignee_id"
	statusColumn := taskAlias + ".task_status"
	if revision {
		taskAlias = "revision_task"
		instanceIDColumn = "revision_request.source_instance_id"
		assigneeColumn = taskAlias + ".assignee_id"
		statusColumn = taskAlias + ".task_status"
	}
	clauses := []string{"task_instance.admin_deleted_at = 0"}
	args := make([]interface{}, 0)
	if query.HideAdminDeleted && !revision {
		clauses = append(clauses, taskAlias+".admin_deleted_at = 0")
	}
	if value := strings.TrimSpace(query.InstanceID); value != "" {
		clauses = append(clauses, instanceIDColumn+" = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.AssigneeID); value != "" {
		clauses = append(clauses, assigneeColumn+" = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		clauses = append(clauses, statusColumn+" = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.InstanceTitle); value != "" {
		clauses = append(clauses, "task_instance.instance_title LIKE ? ESCAPE '!'")
		args = append(args, containsLikePattern(value))
	}
	if value := strings.TrimSpace(query.DefinitionName); value != "" {
		clauses = append(clauses, "COALESCE(NULLIF(task_instance.definition_name_snapshot, ''), NULLIF(task_definition.definition_display_name, ''), task_definition.definition_name) LIKE ? ESCAPE '!'")
		args = append(args, containsLikePattern(value))
	}
	if value := strings.TrimSpace(query.DefinitionCategory); value != "" {
		clauses = append(clauses, "task_definition.definition_category = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(query.StarterName); value != "" {
		clauses = append(clauses, `EXISTS (
			SELECT 1 FROM users task_starter
			WHERE task_starter.id = CAST(task_instance.starter_id AS UNSIGNED)
			AND task_starter.user_name LIKE ? ESCAPE '!'
		)`)
		args = append(args, containsLikePattern(value))
	}
	if query.StartTimeFrom > 0 {
		clauses = append(clauses, "task_instance.start_time >= ?")
		args = append(args, query.StartTimeFrom)
	}
	if query.StartTimeTo > 0 {
		clauses = append(clauses, "task_instance.start_time <= ?")
		args = append(args, query.StartTimeTo)
	}
	return strings.Join(clauses, " AND "), args
}

func loadUnifiedTaskSummaries(db *gorm.DB, rows []unifiedTaskRow) ([]application.TaskSummary, error) {
	processRows := make([]workflowmodel.ProcessTask, 0, len(rows))
	for _, row := range rows {
		processRows = append(processRows, unifiedTaskProcessModel(row))
	}
	list, err := loadTaskSummaries(db, processRows)
	if err != nil {
		return nil, err
	}
	for index := range list {
		list[index].TaskType = rows[index].TaskType
		list[index].RevisionRequestID = rows[index].RevisionRequestID
	}
	return list, nil
}

func unifiedTaskSummary(row unifiedTaskRow) (application.TaskSummary, error) {
	summary, err := taskSummary(unifiedTaskProcessModel(row))
	if err != nil {
		return application.TaskSummary{}, err
	}
	summary.TaskType = row.TaskType
	summary.RevisionRequestID = row.RevisionRequestID
	return summary, nil
}

func unifiedTaskProcessModel(row unifiedTaskRow) workflowmodel.ProcessTask {
	return workflowmodel.ProcessTask{
		ID: row.ID, InstanceID: row.InstanceID, NodeID: row.NodeID, NodeName: row.NodeName,
		AssigneeID: row.AssigneeID, ApprovalMode: row.ApprovalMode, CompletionRate: row.CompletionRate,
		Sequence: row.Sequence, Total: row.Total, ApprovalChainKey: row.ApprovalChainKey,
		ApprovalLayer: row.ApprovalLayer, ApprovalLayerTotal: row.ApprovalLayerTotal,
		SourceDepartmentID: row.SourceDepartmentID, SourceDepartmentName: row.SourceDepartmentName,
		Status: row.Status, Action: row.Action, Comment: row.Comment, ImagesJSON: row.ImagesJSON,
		HandledBy: row.HandledBy, HandledAt: row.HandledAt,
	}
}

func taskSummary(row workflowmodel.ProcessTask) (application.TaskSummary, error) {
	images, err := decodeWorkflowImages(row.ImagesJSON)
	if err != nil {
		return application.TaskSummary{}, err
	}
	return application.TaskSummary{
		ID: row.ID, TaskType: application.TaskTypeWorkflow,
		InstanceID: row.InstanceID, NodeID: row.NodeID, NodeName: row.NodeName,
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
		if err := db.Select("id", "definition_id", "definition_name_snapshot", "starter_id").Where("id IN ?", instanceIDs).Find(&instances).Error; err != nil {
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
			summary.DefinitionName = strings.TrimSpace(instance.DefinitionNameSnapshot)
			if summary.DefinitionName == "" {
				summary.DefinitionName = definitionNames[instance.DefinitionID]
			}
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
