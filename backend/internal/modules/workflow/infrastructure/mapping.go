package infrastructure

import (
	"bytes"
	"encoding/json"
	"strings"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func taskToModel(instanceID string, task workflowdomain.Task, handledAt int64, actorByTask map[string]string) workflowmodel.ProcessTask {
	handledBy := ""
	if actorByTask != nil {
		handledBy = actorByTask[task.ID]
	}
	if handledBy == "" && (task.Status == workflowdomain.TaskStatusApproved || task.Status == workflowdomain.TaskStatusRejected) {
		handledBy = task.AssigneeID
	}
	if task.Status != workflowdomain.TaskStatusApproved && task.Status != workflowdomain.TaskStatusRejected {
		handledAt = 0
	}
	return workflowmodel.ProcessTask{
		ID: task.ID, InstanceID: instanceID, TokenID: task.TokenID,
		NodeID: task.NodeID, NodeName: task.NodeName, GroupKey: task.GroupKey,
		AssigneeID: task.AssigneeID, ApprovalMode: task.ApprovalMode,
		CompletionRate: task.CompletionRate, Sequence: task.Sequence, Total: task.Total,
		Status: string(task.Status), Action: string(task.Action), Comment: task.Comment,
		HandledBy: handledBy, HandledAt: handledAt,
	}
}

func taskFromModel(model workflowmodel.ProcessTask) workflowdomain.Task {
	return workflowdomain.Task{
		ID: model.ID, TokenID: model.TokenID, NodeID: model.NodeID, NodeName: model.NodeName,
		GroupKey: model.GroupKey, AssigneeID: model.AssigneeID, ApprovalMode: model.ApprovalMode,
		CompletionRate: model.CompletionRate, Sequence: model.Sequence, Total: model.Total,
		Status: workflowdomain.TaskStatus(model.Status), Action: workflowdomain.TaskAction(model.Action),
		Comment: model.Comment,
	}
}

func decodeVariable(raw string) (interface{}, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func encodeFormData(formData map[string]interface{}) (string, error) {
	if formData == nil {
		formData = map[string]interface{}{}
	}
	encoded, err := json.Marshal(formData)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func decodeFormData(raw string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func stateFromModels(
	instance workflowmodel.ProcessInstance,
	tokens []workflowmodel.ProcessToken,
	tasks []workflowmodel.ProcessTask,
	variables []workflowmodel.ProcessVariable,
	history []workflowmodel.ProcessHistory,
) (*workflowdomain.State, error) {
	state := &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{
			ID: instance.ID, DefinitionID: instance.DefinitionID,
			DefinitionVersion: instance.DefinitionVersion, DefinitionKey: instance.DefinitionKey,
			BusinessType: instance.BusinessType, BusinessKey: instance.BusinessKey,
			StarterID: instance.StarterID, Status: workflowdomain.InstanceStatus(instance.Status),
		},
		Variables: make(map[string]interface{}, len(variables)),
		FormData:  make(map[string]interface{}),
		Tokens:    make([]workflowdomain.Token, 0, len(tokens)),
		Tasks:     make([]workflowdomain.Task, 0, len(tasks)),
		History:   make([]workflowdomain.HistoryEvent, 0, len(history)),
	}
	formData, err := decodeFormData(instance.FormDataJSON)
	if err != nil {
		return nil, err
	}
	state.FormData = formData
	for _, item := range tokens {
		state.Tokens = append(state.Tokens, workflowdomain.Token{
			ID: item.ID, NodeID: item.NodeID, Status: workflowdomain.TokenStatus(item.Status),
			BranchGroup: item.BranchGroup, BranchTotal: item.BranchTotal,
		})
	}
	for _, item := range tasks {
		state.Tasks = append(state.Tasks, taskFromModel(item))
	}
	for _, item := range variables {
		value, err := decodeVariable(item.ValueJSON)
		if err != nil {
			return nil, err
		}
		state.Variables[item.Key] = value
	}
	for _, item := range history {
		state.History = append(state.History, workflowdomain.HistoryEvent{
			ID: item.ID, Type: workflowdomain.HistoryEventType(item.EventType), NodeID: item.NodeID,
			TaskID: item.TaskID, ActorID: item.ActorID, Message: item.Message,
		})
	}
	return state, nil
}

func handledActors(history []workflowdomain.HistoryEvent) map[string]string {
	result := make(map[string]string)
	for _, event := range history {
		if strings.TrimSpace(event.TaskID) == "" {
			continue
		}
		if event.Type == workflowdomain.HistoryTaskApproved || event.Type == workflowdomain.HistoryTaskRejected {
			result[event.TaskID] = event.ActorID
		}
	}
	return result
}
