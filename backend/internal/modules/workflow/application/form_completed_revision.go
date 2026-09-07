package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) ListFormRevisions(ctx context.Context, instanceID, actorID string) ([]FormRevisionSummary, error) {
	instanceID = strings.TrimSpace(instanceID)
	actorID = strings.TrimSpace(actorID)
	if instanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	if _, err := service.GetMyInstance(ctx, actorID, instanceID); err != nil {
		return nil, err
	}
	return service.store.ListFormRevisions(ctx, instanceID, actorID)
}

func (service *Service) GetFormRevision(ctx context.Context, revisionID, actorID string) (*FormRevisionDetail, error) {
	revisionID = strings.TrimSpace(revisionID)
	actorID = strings.TrimSpace(actorID)
	if revisionID == "" {
		return nil, ErrCompletedRevisionIDRequired
	}
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	return service.store.GetFormRevision(ctx, revisionID, actorID)
}

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
	var outboxIDs []string
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
		kind := workflowdomain.NotificationKindInstanceFormRevisionRequested
		recipients := formRevisionTaskRecipients(revision.Tasks, nil)
		eventKey := "requested:stage:1"
		if revision.Status == workflowdomain.FormRevisionStatusApplied {
			kind = workflowdomain.NotificationKindInstanceFormRevisionApproved
			recipients = formRevisionResultRecipients(plan.state, revision, false)
			eventKey = "approved"
			if err := store.CreateBusinessEvent(ctx, revisionBusinessEvent(service, plan.state, revision)); err != nil {
				return err
			}
		}
		intents, err := service.buildFormRevisionNotifications(ctx, store, plan.definition, plan.state, revision, kind, recipients, eventKey)
		if err != nil {
			return err
		}
		outboxIDs, err = store.PersistRevisionNotifications(ctx, plan.state, revision, intents)
		if err != nil {
			return err
		}
		result = completedRevisionDetail(revision)
		return nil
	})
	if err == nil {
		service.dispatchNotifications(ctx, outboxIDs)
	}
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
	var outboxIDs []string
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

		definition, state, err := store.LoadCompletedRevisionSourceForUpdate(ctx, revision.SourceInstanceID)
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
		var kind workflowdomain.NotificationKind
		var recipients []formRevisionNotificationRecipient
		eventKey := ""
		switch revision.Status {
		case workflowdomain.FormRevisionStatusRejected:
			kind = workflowdomain.NotificationKindInstanceFormRevisionRejected
			recipients = formRevisionResultRecipients(state, revision, false)
			eventKey = "rejected"
		case workflowdomain.FormRevisionStatusApplied:
			kind = workflowdomain.NotificationKindInstanceFormRevisionApproved
			recipients = formRevisionResultRecipients(state, revision, true)
			eventKey = "approved"
			if err := store.CreateBusinessEvent(ctx, revisionBusinessEvent(service, state, revision)); err != nil {
				return err
			}
		case workflowdomain.FormRevisionStatusConflict:
			kind = workflowdomain.NotificationKindInstanceFormRevisionConflict
			recipients = formRevisionResultRecipients(state, revision, false)
			eventKey = "conflict"
		default:
			if len(transition.ActivatedTaskIDs) > 0 {
				kind = workflowdomain.NotificationKindInstanceFormRevisionRequested
				activated := make(map[string]struct{}, len(transition.ActivatedTaskIDs))
				for _, taskID := range transition.ActivatedTaskIDs {
					activated[taskID] = struct{}{}
				}
				recipients = formRevisionTaskRecipients(revision.Tasks, activated)
				eventKey = fmt.Sprintf("requested:stage:%d", completedRevisionTaskByID(revision, transition.ActivatedTaskIDs[0]).Stage)
			}
		}
		if kind != "" {
			intents, err := service.buildFormRevisionNotifications(ctx, store, definition, state, revision, kind, recipients, eventKey)
			if err != nil {
				return err
			}
			outboxIDs, err = store.PersistRevisionNotifications(ctx, state, revision, intents)
			if err != nil {
				return err
			}
		}
		detail = completedRevisionDetail(revision)
		return nil
	})
	if err == nil {
		service.dispatchNotifications(ctx, outboxIDs)
	}
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
	var outboxIDs []string
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
		recipients := formRevisionTaskRecipients(revision.Tasks, nil)
		now := service.currentTime().UnixMilli()
		for index := range revision.Tasks {
			if revision.Tasks[index].Status == workflowdomain.RevisionTaskStatusPending ||
				revision.Tasks[index].Status == workflowdomain.RevisionTaskStatusWaiting {
				revision.Tasks[index].Status = workflowdomain.RevisionTaskStatusCancelled
			}
		}
		revision.Status = workflowdomain.FormRevisionStatusCancelled
		revision.CompletedAt = now
		definition, state, err := store.LoadCompletedRevisionSourceForUpdate(ctx, revision.SourceInstanceID)
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
		intents, err := service.buildFormRevisionNotifications(
			ctx, store, definition, state, revision,
			workflowdomain.NotificationKindInstanceFormRevisionCancelled,
			recipients, "cancelled",
		)
		if err != nil {
			return err
		}
		outboxIDs, err = store.PersistRevisionNotifications(ctx, state, revision, intents)
		if err != nil {
			return err
		}
		detail = completedRevisionDetail(revision)
		return nil
	})
	if err == nil {
		service.dispatchNotifications(ctx, outboxIDs)
	}
	return detail, err
}
