package application

import (
	"context"
	"errors"
	"strings"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowcore "wecheckin/backend/internal/workflow"
)

var (
	ErrDefinitionRequired        = errors.New("流程定义不能为空")
	ErrStarterRequired           = errors.New("流程发起人不能为空")
	ErrBusinessReferenceRequired = errors.New("业务类型和业务标识不能为空")
	ErrTaskIDRequired            = errors.New("流程任务不能为空")
	ErrActorRequired             = errors.New("任务处理人不能为空")
	ErrInstanceIDRequired        = errors.New("流程实例不能为空")
	ErrInstanceAccessDenied      = errors.New("无权访问该流程实例")
)

type Store interface {
	InTransaction(ctx context.Context, fn func(TransactionStore) error) error
	ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error)
	GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error)
	ListInstances(ctx context.Context, query InstanceQuery) (*InstanceList, error)
	GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error)
	ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error)
}

type TransactionStore interface {
	LoadPublishedDefinition(ctx context.Context, definitionID uint, version int) (workflowcore.Definition, int, error)
	CreateState(ctx context.Context, state *workflowdomain.State) error
	LoadStateByTaskForUpdate(ctx context.Context, taskID string) (workflowcore.Definition, *workflowdomain.State, error)
	LoadStateByInstanceForUpdate(ctx context.Context, instanceID string) (*workflowdomain.State, error)
	SaveState(ctx context.Context, state *workflowdomain.State) error
}

type Service struct {
	store     Store
	engine    *workflowdomain.Engine
	publisher EventPublisher
}

func NewService(store Store, resolver workflowdomain.AssigneeResolver, ids workflowdomain.IDGenerator) *Service {
	return NewServiceWithPublisher(store, resolver, ids, noopEventPublisher{})
}

func NewServiceWithPublisher(store Store, resolver workflowdomain.AssigneeResolver, ids workflowdomain.IDGenerator, publisher EventPublisher) *Service {
	if publisher == nil {
		publisher = noopEventPublisher{}
	}
	return &Service{store: store, engine: workflowdomain.NewEngine(resolver, ids), publisher: publisher}
}

type StartInstanceRequest struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	BusinessType      string                 `json:"businessType"`
	BusinessKey       string                 `json:"businessKey"`
	StarterID         string                 `json:"starterId"`
	Variables         map[string]interface{} `json:"variables"`
	FormData          map[string]interface{} `json:"formData"`
}

type CompleteTaskRequest struct {
	TaskID    string                    `json:"taskId"`
	ActorID   string                    `json:"actorId"`
	Action    workflowdomain.TaskAction `json:"action"`
	Comment   string                    `json:"comment"`
	Variables map[string]interface{}    `json:"variables"`
	FormData  map[string]interface{}    `json:"formData"`
}

type WithdrawInstanceRequest struct {
	InstanceID string `json:"instanceId"`
	ActorID    string `json:"actorId"`
	Reason     string `json:"reason"`
}

type CancelInstanceRequest struct {
	InstanceID string `json:"instanceId"`
	ActorID    string `json:"actorId"`
	Reason     string `json:"reason"`
}

func (service *Service) StartInstance(ctx context.Context, request StartInstanceRequest) (*workflowdomain.State, error) {
	request.BusinessType = strings.TrimSpace(request.BusinessType)
	request.BusinessKey = strings.TrimSpace(request.BusinessKey)
	request.StarterID = strings.TrimSpace(request.StarterID)
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if request.StarterID == "" {
		return nil, ErrStarterRequired
	}
	if request.BusinessType == "" || request.BusinessKey == "" {
		return nil, ErrBusinessReferenceRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var state *workflowdomain.State
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, publishedVersion, err := store.LoadPublishedDefinition(ctx, request.DefinitionID, request.DefinitionVersion)
		if err != nil {
			return err
		}
		if err := workflowcore.ValidateFormData(definition.Form, request.FormData, false); err != nil {
			return err
		}
		state, err = service.engine.Start(definition, workflowdomain.StartRequest{
			DefinitionID:      request.DefinitionID,
			DefinitionVersion: publishedVersion,
			BusinessType:      request.BusinessType,
			BusinessKey:       request.BusinessKey,
			StarterID:         request.StarterID,
			Variables:         request.Variables,
			FormData:          request.FormData,
		})
		if err != nil {
			return err
		}
		return store.CreateState(ctx, state)
	})
	if err == nil {
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceStarted, InstanceID: state.Instance.ID, ActorID: request.StarterID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, err
}

func (service *Service) CompleteTask(ctx context.Context, request CompleteTaskRequest) (*workflowdomain.State, error) {
	request.TaskID = strings.TrimSpace(request.TaskID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	if request.TaskID == "" {
		return nil, ErrTaskIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var state *workflowdomain.State
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
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
		if err := service.engine.Complete(definition, loaded, workflowdomain.CompleteRequest{
			TaskID: request.TaskID, ActorID: request.ActorID, Action: request.Action,
			Comment: strings.TrimSpace(request.Comment), Variables: request.Variables, FormData: request.FormData,
		}); err != nil {
			return err
		}
		state = loaded
		return store.SaveState(ctx, state)
	})
	if err == nil {
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

func (service *Service) ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error) {
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	return service.store.ListPublishedDefinitions(ctx)
}

func (service *Service) GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error) {
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	return service.store.GetPublishedDefinition(ctx, definitionID)
}

func (service *Service) changeInstanceStatus(ctx context.Context, instanceID, actorID, reason string, adminCancel bool) (*workflowdomain.State, error) {
	instanceID = strings.TrimSpace(instanceID)
	actorID = strings.TrimSpace(actorID)
	if instanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	var state *workflowdomain.State
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		loaded, err := store.LoadStateByInstanceForUpdate(ctx, instanceID)
		if err != nil {
			return err
		}
		if adminCancel {
			err = service.engine.Cancel(loaded, actorID, strings.TrimSpace(reason))
		} else {
			err = service.engine.Withdraw(loaded, actorID, strings.TrimSpace(reason))
		}
		if err != nil {
			return err
		}
		state = loaded
		return store.SaveState(ctx, state)
	})
	return state, err
}

func (service *Service) ListInstances(ctx context.Context, query InstanceQuery) (*InstanceList, error) {
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.store.ListInstances(ctx, query)
}

func (service *Service) ListMyInstances(ctx context.Context, actorID string, query InstanceQuery) (*InstanceList, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	query.StarterID = actorID
	return service.ListInstances(ctx, query)
}

func (service *Service) GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return nil, errors.New("流程实例不能为空")
	}
	return service.store.GetInstance(ctx, instanceID)
}

func (service *Service) GetMyInstance(ctx context.Context, actorID, instanceID string) (*InstanceDetail, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	detail, err := service.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if detail.Instance.StarterID == actorID {
		return detail, nil
	}
	for _, task := range detail.Tasks {
		if task.AssigneeID == actorID {
			return detail, nil
		}
	}
	return nil, ErrInstanceAccessDenied
}

func (service *Service) ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error) {
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.store.ListTasks(ctx, query)
}

func (service *Service) ListMyTasks(ctx context.Context, actorID string, query TaskQuery) (*TaskList, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	query.AssigneeID = actorID
	return service.ListTasks(ctx, query)
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (service *Service) publish(ctx context.Context, event LifecycleEvent) {
	if service != nil && service.publisher != nil {
		service.publisher.Publish(ctx, event)
	}
}
