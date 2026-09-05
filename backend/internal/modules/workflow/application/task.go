package application

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func (service *Service) CompleteTask(ctx context.Context, request CompleteTaskRequest) (*workflowdomain.State, error) {
	request.TaskID = strings.TrimSpace(request.TaskID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Comment = strings.TrimSpace(request.Comment)
	request.ReturnTargetNodeID = strings.TrimSpace(request.ReturnTargetNodeID)
	if request.TaskID == "" {
		return nil, ErrTaskIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.Action == workflowdomain.TaskActionReject && request.Comment == "" {
		return nil, ErrTaskRejectCommentRequired
	}
	if request.Action == workflowdomain.TaskActionReturn && request.Comment == "" {
		return nil, ErrTaskReturnCommentRequired
	}
	if utf8.RuneCountInString(request.Comment) > 500 {
		return nil, ErrTaskCommentTooLong
	}
	images, err := normalizeWorkflowImages(request.Images)
	if err != nil {
		return nil, err
	}
	request.Images = images
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var state *workflowdomain.State
	var outboxIDs []string
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, loaded, err := store.LoadStateByTaskForUpdate(ctx, request.TaskID)
		if err != nil {
			return err
		}
		if loaded.Variables == nil {
			loaded.Variables = make(map[string]interface{})
		}
		if loaded.FormData == nil {
			loaded.FormData = make(map[string]interface{})
		}
		if request.FormData == nil {
			request.FormData = make(map[string]interface{})
		}
		taskNodeID := ""
		for _, task := range loaded.Tasks {
			if task.ID == request.TaskID {
				taskNodeID = task.NodeID
				break
			}
		}
		if err := workflowcore.ValidateNodeFormPatch(definition, taskNodeID, loaded.FormData, request.FormData); err != nil {
			return err
		}
		calculatedData, err := workflowcore.ApplyFormCalculations(definition.Form, workflowcore.MergeFormData(loaded.FormData, request.FormData))
		if err != nil {
			return err
		}
		for key, value := range changedFormDataPatch(loaded.FormData, calculatedData) {
			request.FormData[key] = value
		}
		formDataChanged := len(changedFormDataPatch(loaded.FormData, request.FormData)) > 0
		if err := service.engine.Complete(ctx, definition, loaded, workflowdomain.CompleteRequest{
			TaskID: request.TaskID, ActorID: request.ActorID, Action: request.Action,
			Comment: request.Comment, Images: request.Images, ReturnTargetNodeID: request.ReturnTargetNodeID,
			Variables: request.Variables, FormData: request.FormData,
		}); err != nil {
			return err
		}
		if formDataChanged {
			loaded.Instance.FormRevision++
		}
		state = loaded
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		outboxIDs, err = store.PersistEffects(ctx, state)
		return err
	})
	if err == nil {
		service.dispatchNotifications(ctx, outboxIDs)
		service.publish(ctx, LifecycleEvent{Type: LifecycleTaskCompleted, InstanceID: state.Instance.ID, TaskID: request.TaskID,
			ActorID: request.ActorID, BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
		switch state.Instance.Status {
		case workflowdomain.InstanceStatusCompleted:
			service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceCompleted, InstanceID: state.Instance.ID, ActorID: request.ActorID,
				BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
		case workflowdomain.InstanceStatusRejected:
			service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceRejected, InstanceID: state.Instance.ID, ActorID: request.ActorID,
				BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
		}
	}
	return state, err
}

func (service *Service) WithdrawInstance(ctx context.Context, request WithdrawInstanceRequest) (*workflowdomain.State, error) {
	state, err := service.changeInstanceStatus(ctx, request.InstanceID, request.ActorID, request.Reason, false)
	if err == nil {
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceWithdrawn, InstanceID: state.Instance.ID, ActorID: request.ActorID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, err
}

func (service *Service) CancelInstance(ctx context.Context, request CancelInstanceRequest) (*workflowdomain.State, error) {
	state, err := service.changeInstanceStatus(ctx, request.InstanceID, request.ActorID, request.Reason, true)
	if err == nil {
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceCancelled, InstanceID: state.Instance.ID, ActorID: request.ActorID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, err
}

func (service *Service) ResumeTimers(ctx context.Context, instanceID, actorID string) (*workflowdomain.State, int, error) {
	instanceID = strings.TrimSpace(instanceID)
	actorID = strings.TrimSpace(actorID)
	if instanceID == "" {
		return nil, 0, ErrInstanceIDRequired
	}
	if actorID == "" {
		return nil, 0, ErrActorRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, 0, errors.New("工作流应用服务未初始化")
	}
	var state *workflowdomain.State
	var outboxIDs []string
	advanced := 0
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, loaded, err := store.LoadDefinitionAndStateByInstanceForUpdate(ctx, instanceID)
		if err != nil {
			return err
		}
		advanced, err = service.engine.ResumeTimers(ctx, definition, loaded, time.Now().Unix())
		if err != nil {
			return err
		}
		state = loaded
		if advanced == 0 {
			return nil
		}
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		outboxIDs, err = store.PersistEffects(ctx, state)
		return err
	})
	if err == nil && advanced > 0 {
		service.dispatchNotifications(ctx, outboxIDs)
	}
	if err == nil && advanced > 0 && state.Instance.Status == workflowdomain.InstanceStatusCompleted {
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceCompleted, InstanceID: state.Instance.ID, ActorID: actorID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, advanced, err
}

func (service *Service) DeleteTask(ctx context.Context, actorID, taskID string) error {
	actorID = strings.TrimSpace(actorID)
	taskID = strings.TrimSpace(taskID)
	if actorID == "" {
		return ErrActorRequired
	}
	if taskID == "" {
		return ErrTaskIDRequired
	}
	if service == nil || service.store == nil {
		return errors.New("工作流应用服务未初始化")
	}
	return service.store.InTransaction(ctx, func(store TransactionStore) error {
		task, err := store.LoadTaskForDelete(ctx, taskID)
		if err != nil {
			return err
		}
		if task == nil {
			return ErrTaskDeleteTargetNotFound
		}
		switch task.Status {
		case string(workflowdomain.TaskStatusCompleted),
			string(workflowdomain.TaskStatusApproved),
			string(workflowdomain.TaskStatusRejected),
			string(workflowdomain.TaskStatusReturned),
			string(workflowdomain.TaskStatusCancelled):
		default:
			return ErrTaskDeleteNotAllowed
		}
		count, err := store.SoftDeleteTask(ctx, taskID, actorID, service.currentTime().UnixMilli())
		if err != nil {
			return err
		}
		if count != 1 {
			return ErrTaskDeleteTargetNotFound
		}
		return nil
	})
}

func (service *Service) ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error) {
	definitionName, err := normalizeDefinitionNameSearch(query.DefinitionName)
	if err != nil {
		return nil, err
	}
	starterName, err := normalizeStarterNameSearch(query.StarterName)
	if err != nil {
		return nil, err
	}
	query.DefinitionName = definitionName
	query.StarterName = starterName
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.store.ListTasks(ctx, query)
}
