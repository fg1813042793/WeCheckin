package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func completedRevisionTaskByID(revision *workflowdomain.FormRevisionRequest, taskID string) workflowdomain.FormRevisionTask {
	if revision == nil {
		return workflowdomain.FormRevisionTask{}
	}
	for _, task := range revision.Tasks {
		if task.ID == taskID {
			return task
		}
	}
	return workflowdomain.FormRevisionTask{}
}

func completedRevisionHasHandledTask(revision *workflowdomain.FormRevisionRequest) bool {
	if revision == nil {
		return false
	}
	for _, task := range revision.Tasks {
		if task.HandledAt > 0 || task.Status == workflowdomain.RevisionTaskStatusApproved || task.Status == workflowdomain.RevisionTaskStatusRejected {
			return true
		}
	}
	return false
}

func normalizeCompletedRevisionPreviewRequest(request PreviewCompletedFormRevisionRequest) (PreviewCompletedFormRevisionRequest, error) {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.SourceNodeID = strings.TrimSpace(request.SourceNodeID)
	request.RequesterID = strings.TrimSpace(request.RequesterID)
	if request.InstanceID == "" {
		return request, ErrInstanceIDRequired
	}
	if request.SourceNodeID == "" {
		return request, ErrCompletedRevisionNotAllowed
	}
	if request.RequesterID == "" {
		return request, ErrActorRequired
	}
	if request.ExpectedRevision <= 0 {
		return request, ErrInstanceFormRevisionRequired
	}
	if len(request.FormData) == 0 {
		return request, ErrInstanceFormRevisionDataRequired
	}
	return request, nil
}

func (service *Service) planCompletedRevision(
	ctx context.Context,
	store TransactionStore,
	request PreviewCompletedFormRevisionRequest,
) (*completedRevisionPlan, error) {
	definition, state, err := store.LoadCompletedRevisionSourceForUpdate(ctx, request.InstanceID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.Instance.Status != workflowdomain.InstanceStatusCompleted {
		return nil, ErrCompletedRevisionInstanceNotCompleted
	}
	if state.Instance.FormRevision != request.ExpectedRevision {
		return nil, ErrInstanceFormRevisionChanged
	}
	active, err := store.LoadActiveFormRevisionForUpdate(ctx, request.InstanceID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrCompletedRevisionActive
	}
	if !completedRevisionSourceHandledBy(state, request.SourceNodeID, request.RequesterID) {
		return nil, ErrCompletedRevisionNotAllowed
	}
	capability, err := workflowcore.CompletedRevisionCapabilityForNode(definition, request.SourceNodeID)
	if err != nil {
		return nil, ErrCompletedRevisionNotAllowed
	}
	patch := changedFormDataPatch(state.FormData, request.FormData)
	if len(patch) == 0 {
		return nil, ErrInstanceFormRevisionDataRequired
	}
	if err := workflowcore.ValidateCompletedRevisionPatch(definition, request.SourceNodeID, state.FormData, patch); err != nil {
		return nil, err
	}
	proposed, err := workflowcore.ApplyFormCalculations(definition.Form, workflowcore.MergeFormData(state.FormData, patch))
	if err != nil {
		return nil, err
	}
	mode := workflowcore.CompletedRevisionMode(capability, patch)
	plan := &completedRevisionPlan{
		definition: definition, state: state, capability: capability,
		patch: patch, proposedFormData: proposed,
		changedFieldLabels: completedRevisionFieldLabels(definition.Form, patch), mode: mode,
	}
	if mode == workflowcore.CompletedRevisionModeDirect {
		return plan, nil
	}
	stages, err := workflowcore.BuildCompletedRevisionPlan(definition, completedRevisionActualHumanNodes(state), request.SourceNodeID)
	if err != nil {
		return nil, err
	}
	if len(stages) == 0 {
		return nil, ErrCompletedRevisionNoReviewPath
	}
	plan.confirmationStages, err = service.resolveCompletedRevisionStages(ctx, store, definition, state, stages)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func completedRevisionSourceHandledBy(state *workflowdomain.State, nodeID, actorID string) bool {
	nodeID = strings.TrimSpace(nodeID)
	actorID = strings.TrimSpace(actorID)
	if state == nil || nodeID == "" || actorID == "" {
		return false
	}
	for _, event := range state.History {
		if strings.TrimSpace(event.NodeID) != nodeID || strings.TrimSpace(event.ActorID) != actorID {
			continue
		}
		switch event.Type {
		case workflowdomain.HistoryTaskApproved, workflowdomain.HistoryTaskSubmitted, workflowdomain.HistoryTaskReturned:
			return true
		}
	}
	return false
}

func completedRevisionActualHumanNodes(state *workflowdomain.State) []string {
	if state == nil {
		return nil
	}
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, event := range state.History {
		switch event.Type {
		case workflowdomain.HistoryTaskApproved, workflowdomain.HistoryTaskSubmitted, workflowdomain.HistoryTaskReturned:
		default:
			continue
		}
		nodeID := strings.TrimSpace(event.NodeID)
		if nodeID == "" {
			continue
		}
		if _, exists := seen[nodeID]; exists {
			continue
		}
		seen[nodeID] = struct{}{}
		result = append(result, nodeID)
	}
	return result
}

func (service *Service) resolveCompletedRevisionStages(
	ctx context.Context,
	store TransactionStore,
	definition workflowcore.Definition,
	state *workflowdomain.State,
	stages []workflowcore.RevisionPlanStage,
) ([]FormRevisionConfirmationStage, error) {
	nodes := make(map[string]workflowcore.Node, len(definition.Nodes))
	for _, node := range definition.Nodes {
		nodes[node.ID] = node
	}
	result := make([]FormRevisionConfirmationStage, 0, len(stages))
	for _, stage := range stages {
		stageResult := FormRevisionConfirmationStage{Stage: stage.Stage, Nodes: make([]FormRevisionConfirmationNode, 0, len(stage.Nodes))}
		for _, nodeID := range stage.Nodes {
			node, ok := nodes[nodeID]
			if !ok {
				return nil, fmt.Errorf("确认节点 %s 不存在", nodeID)
			}
			assignees, err := service.resolver.Resolve(ctx, workflowdomain.AssigneeRequest{
				Instance: state.Instance, Node: node, Variables: workflowcore.MergeFormData(nil, state.Variables),
			})
			if err != nil {
				return nil, err
			}
			assignees = uniqueTrimmedStrings(assignees)
			mode := node.ApprovalMode
			completionRate := node.CompletionRate
			if node.Type == workflowcore.NodeTypeHandle || mode == workflowcore.ApprovalModeSingle {
				mode = workflowcore.ApprovalModeSingle
				completionRate = 100
				if len(assignees) > 1 {
					assignees = assignees[:1]
				}
			}
			if len(assignees) == 0 {
				return nil, workflowdomain.ErrAssigneeUnavailable
			}
			names := make([]string, 0, len(assignees))
			for _, assigneeID := range assignees {
				name, err := store.UserDisplayName(ctx, assigneeID)
				if err != nil {
					return nil, err
				}
				names = append(names, name)
			}
			stageResult.Nodes = append(stageResult.Nodes, FormRevisionConfirmationNode{
				NodeID: node.ID, NodeName: node.Name, ApprovalMode: mode, CompletionRate: completionRate,
				AssigneeIDs: append([]string(nil), assignees...), AssigneeNames: names,
			})
		}
		result = append(result, stageResult)
	}
	return result, nil
}

func (service *Service) newCompletedRevisionTasks(
	state *workflowdomain.State,
	revisionID string,
	stages []FormRevisionConfirmationStage,
) []workflowdomain.FormRevisionTask {
	result := make([]workflowdomain.FormRevisionTask, 0)
	for _, stage := range stages {
		for _, node := range stage.Nodes {
			for index, assigneeID := range node.AssigneeIDs {
				status := workflowdomain.RevisionTaskStatusWaiting
				if stage.Stage == 1 && (node.ApprovalMode != workflowcore.ApprovalModeSequential || index == 0) {
					status = workflowdomain.RevisionTaskStatusPending
				}
				result = append(result, workflowdomain.FormRevisionTask{
					ID: service.ids.NewID("form-revision-task"), RevisionRequestID: revisionID,
					SourceTaskID: completedRevisionSourceTaskID(state, node.NodeID, assigneeID),
					NodeID:       node.NodeID, NodeName: node.NodeName, Stage: stage.Stage,
					AssigneeID: assigneeID, AssigneeName: node.AssigneeNames[index],
					ApprovalMode: node.ApprovalMode, CompletionRate: node.CompletionRate,
					Sequence: index + 1, Total: len(node.AssigneeIDs), Status: status,
				})
			}
		}
	}
	return result
}

func completedRevisionSourceTaskID(state *workflowdomain.State, nodeID, assigneeID string) string {
	if state == nil {
		return ""
	}
	fallback := ""
	for index := len(state.Tasks) - 1; index >= 0; index-- {
		task := state.Tasks[index]
		if task.NodeID != nodeID {
			continue
		}
		if fallback == "" {
			fallback = task.ID
		}
		if task.AssigneeID == assigneeID {
			return task.ID
		}
	}
	return fallback
}

func completedRevisionFieldLabels(fields []workflowcore.FormField, patch map[string]interface{}) []string {
	labelsByField := make(map[string]string)
	var collect func([]workflowcore.FormField)
	collect = func(items []workflowcore.FormField) {
		for _, field := range items {
			if field.Type == workflowcore.FormFieldTypeGroup {
				collect(field.Fields)
				continue
			}
			labelsByField[field.Key] = strings.TrimSpace(field.Label)
		}
	}
	collect(fields)
	keys := make([]string, 0, len(patch))
	for key := range patch {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	labels := make([]string, 0, len(keys))
	for _, key := range keys {
		label := labelsByField[key]
		if label == "" {
			label = key
		}
		labels = append(labels, label)
	}
	return labels
}

func completedRevisionDetail(revision *workflowdomain.FormRevisionRequest) *FormRevisionDetail {
	if revision == nil {
		return nil
	}
	result := &FormRevisionDetail{
		FormRevisionSummary: FormRevisionSummary{
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
		Tasks:            make([]FormRevisionTaskSummary, 0, len(revision.Tasks)),
	}
	for _, task := range revision.Tasks {
		result.Tasks = append(result.Tasks, FormRevisionTaskSummary{
			ID: task.ID, SourceTaskID: task.SourceTaskID, NodeID: task.NodeID, NodeName: task.NodeName,
			Stage: task.Stage, AssigneeID: task.AssigneeID, AssigneeName: task.AssigneeName,
			ApprovalMode: task.ApprovalMode, CompletionRate: task.CompletionRate,
			Sequence: task.Sequence, Total: task.Total, Status: string(task.Status), Action: string(task.Action),
			Comment: task.Comment, Images: append([]workflowcore.FormAttachment(nil), task.Images...),
			HandledBy: task.HandledBy, HandledAt: task.HandledAt,
		})
	}
	return result
}

func cloneRevisionConfirmationStages(source []FormRevisionConfirmationStage) []FormRevisionConfirmationStage {
	result := make([]FormRevisionConfirmationStage, 0, len(source))
	for _, stage := range source {
		cloned := FormRevisionConfirmationStage{Stage: stage.Stage, Nodes: make([]FormRevisionConfirmationNode, 0, len(stage.Nodes))}
		for _, node := range stage.Nodes {
			node.AssigneeIDs = append([]string(nil), node.AssigneeIDs...)
			node.AssigneeNames = append([]string(nil), node.AssigneeNames...)
			cloned.Nodes = append(cloned.Nodes, node)
		}
		result = append(result, cloned)
	}
	return result
}
