package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

type completedRevisionPlan struct {
	definition         workflowcore.Definition
	state              *workflowdomain.State
	capability         workflowcore.CompletedRevisionCapability
	patch              map[string]interface{}
	proposedFormData   map[string]interface{}
	changedFieldLabels []string
	mode               string
	confirmationStages []FormRevisionConfirmationStage
}

func (service *Service) PreviewCompletedFormRevision(ctx context.Context, request PreviewCompletedFormRevisionRequest) (*FormRevisionPreview, error) {
	normalized, err := normalizeCompletedRevisionPreviewRequest(request)
	if err != nil {
		return nil, err
	}
	if service == nil || service.store == nil || service.resolver == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	var plan *completedRevisionPlan
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		var planErr error
		plan, planErr = service.planCompletedRevision(ctx, store, normalized)
		return planErr
	})
	if err != nil {
		return nil, err
	}
	return &FormRevisionPreview{
		Mode: plan.mode, ChangedFields: append([]string(nil), plan.changedFieldLabels...),
		ConfirmationStages: cloneRevisionConfirmationStages(plan.confirmationStages),
	}, nil
}

func (service *Service) CreateCompletedFormRevision(ctx context.Context, request CreateCompletedFormRevisionRequest) (*FormRevisionDetail, error) {
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return nil, ErrInstanceFormRevisionReasonRequired
	}
	if utf8.RuneCountInString(reason) > 500 {
		return nil, ErrInstanceFormRevisionReasonTooLong
	}
	normalized, err := normalizeCompletedRevisionPreviewRequest(PreviewCompletedFormRevisionRequest{
		InstanceID: request.InstanceID, SourceNodeID: request.SourceNodeID,
		RequesterID: request.RequesterID, ExpectedRevision: request.ExpectedRevision, FormData: request.FormData,
	})
	if err != nil {
		return nil, err
	}
	if service == nil || service.store == nil || service.resolver == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var result *FormRevisionDetail
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		plan, err := service.planCompletedRevision(ctx, store, normalized)
		if err != nil {
			return err
		}
		now := service.currentTime().UnixMilli()
		revision := &workflowdomain.FormRevisionRequest{
			ID: service.ids.NewID("form-revision"), SourceInstanceID: normalized.InstanceID,
			SourceNodeID: plan.capability.NodeID, SourceNodeName: plan.capability.NodeName,
			RequesterID: normalized.RequesterID, BaseFormRevision: plan.state.Instance.FormRevision,
			BeforeFormData:     workflowcore.MergeFormData(nil, plan.state.FormData),
			Patch:              workflowcore.MergeFormData(nil, plan.patch),
			ProposedFormData:   workflowcore.MergeFormData(nil, plan.proposedFormData),
			ChangedFieldLabels: append([]string(nil), plan.changedFieldLabels...),
			Reason:             reason, Mode: plan.mode, Status: workflowdomain.FormRevisionStatusPending,
			CreatedAt: now,
		}
		if plan.mode == workflowcore.CompletedRevisionModeDirect {
			revision.Status = workflowdomain.FormRevisionStatusApplied
			revision.CompletedAt = now
			plan.state.FormData = workflowcore.MergeFormData(nil, plan.proposedFormData)
			plan.state.Instance.FormRevision++
			revision.AppliedFormRevision = plan.state.Instance.FormRevision
		} else {
			revision.Tasks = service.newCompletedRevisionTasks(plan.state, revision.ID, plan.confirmationStages)
		}

		requestedHistory := workflowdomain.HistoryEvent{
			ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionRequested,
			NodeID: revision.SourceNodeID, ActorID: revision.RequesterID,
			Message:   fmt.Sprintf("申请修订字段：%s；修订原因：%s", strings.Join(revision.ChangedFieldLabels, "、"), revision.Reason),
			EventTime: now,
		}
		plan.state.History = append(plan.state.History, requestedHistory)
		if revision.Status == workflowdomain.FormRevisionStatusApplied {
			plan.state.History = append(plan.state.History, workflowdomain.HistoryEvent{
				ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionApplied,
				NodeID: revision.SourceNodeID, ActorID: revision.RequesterID,
				Message: fmt.Sprintf("表单修订已生效，当前版本：%d", revision.AppliedFormRevision), EventTime: now,
			})
		}
		if err := store.CreateFormRevision(ctx, revision); err != nil {
			return err
		}
		if err := store.SaveState(ctx, plan.state); err != nil {
			return err
		}
		result = completedRevisionDetail(revision)
		return nil
	})
	return result, err
}

func (service *Service) CompleteFormRevisionTask(ctx context.Context, request CompleteFormRevisionTaskRequest) (*FormRevisionDetail, error) {
	request.TaskID = strings.TrimSpace(request.TaskID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Comment = strings.TrimSpace(request.Comment)
	if request.TaskID == "" {
		return nil, ErrTaskIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.Action == workflowdomain.RevisionActionReject && request.Comment == "" {
		return nil, ErrTaskRejectCommentRequired
	}
	if utf8.RuneCountInString(request.Comment) > 500 {
		return nil, ErrTaskCommentTooLong
	}
	images, err := normalizeWorkflowImages(request.Images)
	if err != nil {
		return nil, err
	}
	request.Images = images
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var detail *FormRevisionDetail
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		revision, err := store.LoadFormRevisionByTaskForUpdate(ctx, request.TaskID)
		if err != nil {
			return err
		}
		now := service.currentTime().UnixMilli()
		transition, err := revision.CompleteTask(
			request.TaskID, request.ActorID, request.Action, request.Comment, request.Images, now,
		)
		if err != nil {
			return err
		}
		if transition.Idempotent {
			detail = completedRevisionDetail(revision)
			detail.Idempotent = true
			return nil
		}

		_, state, err := store.LoadCompletedRevisionSourceForUpdate(ctx, revision.SourceInstanceID)
		if err != nil {
			return err
		}
		stateChanged := false
		task := completedRevisionTaskByID(revision, request.TaskID)
		switch revision.Status {
		case workflowdomain.FormRevisionStatusRejected:
			state.History = append(state.History, workflowdomain.HistoryEvent{
				ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionRejected,
				NodeID: task.NodeID, TaskID: task.ID, ActorID: request.ActorID,
				Message: "表单修订已驳回：" + request.Comment,
				Images:  append([]workflowcore.FormAttachment(nil), request.Images...), EventTime: now,
			})
			stateChanged = true
		case workflowdomain.FormRevisionStatusApplied:
			if state.Instance.FormRevision != revision.BaseFormRevision {
				revision.Status = workflowdomain.FormRevisionStatusConflict
				revision.AppliedFormRevision = 0
				state.History = append(state.History, workflowdomain.HistoryEvent{
					ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionConflict,
					NodeID: task.NodeID, TaskID: task.ID, ActorID: request.ActorID,
					Message:   fmt.Sprintf("表单版本已变化，修订基准版本：%d，当前版本：%d", revision.BaseFormRevision, state.Instance.FormRevision),
					EventTime: now,
				})
				stateChanged = true
				break
			}
			state.FormData = workflowcore.MergeFormData(nil, revision.ProposedFormData)
			state.Instance.FormRevision++
			revision.AppliedFormRevision = state.Instance.FormRevision
			state.History = append(state.History, workflowdomain.HistoryEvent{
				ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionApplied,
				NodeID: task.NodeID, TaskID: task.ID, ActorID: request.ActorID,
				Message: fmt.Sprintf("表单修订已完成确认并生效，当前版本：%d", revision.AppliedFormRevision), EventTime: now,
			})
			stateChanged = true
		}
		if err := store.SaveFormRevision(ctx, revision); err != nil {
			return err
		}
		if stateChanged {
			if err := store.SaveState(ctx, state); err != nil {
				return err
			}
		}
		detail = completedRevisionDetail(revision)
		return nil
	})
	return detail, err
}

func (service *Service) CancelCompletedFormRevision(ctx context.Context, request CancelCompletedFormRevisionRequest) (*FormRevisionDetail, error) {
	request.RevisionID = strings.TrimSpace(request.RevisionID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	if request.RevisionID == "" {
		return nil, ErrCompletedRevisionIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var detail *FormRevisionDetail
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		revision, err := store.LoadFormRevisionForUpdate(ctx, request.RevisionID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(revision.RequesterID) != request.ActorID {
			return ErrCompletedRevisionNotAllowed
		}
		if revision.Status != workflowdomain.FormRevisionStatusPending || completedRevisionHasHandledTask(revision) {
			return ErrCompletedRevisionCannotCancel
		}
		now := service.currentTime().UnixMilli()
		for index := range revision.Tasks {
			if revision.Tasks[index].Status == workflowdomain.RevisionTaskStatusPending ||
				revision.Tasks[index].Status == workflowdomain.RevisionTaskStatusWaiting {
				revision.Tasks[index].Status = workflowdomain.RevisionTaskStatusCancelled
			}
		}
		revision.Status = workflowdomain.FormRevisionStatusCancelled
		revision.CompletedAt = now
		_, state, err := store.LoadCompletedRevisionSourceForUpdate(ctx, revision.SourceInstanceID)
		if err != nil {
			return err
		}
		state.History = append(state.History, workflowdomain.HistoryEvent{
			ID: service.ids.NewID("history"), Type: workflowdomain.HistoryInstanceFormRevisionCancelled,
			NodeID: revision.SourceNodeID, ActorID: request.ActorID, Message: "表单修订已取消", EventTime: now,
		})
		if err := store.SaveFormRevision(ctx, revision); err != nil {
			return err
		}
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		detail = completedRevisionDetail(revision)
		return nil
	})
	return detail, err
}

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
