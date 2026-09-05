package infrastructure

import (
	"gorm.io/gorm"
	"strconv"
	"strings"
	"wecheckin/backend/internal/model"
	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

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
		FormRevision: row.FormRevision,
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
