package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

var (
	ErrDefinitionRequired        = errors.New("流程定义不能为空")
	ErrStarterRequired           = errors.New("流程发起人不能为空")
	ErrOperatorRequired          = errors.New("流程发起操作人不能为空")
	ErrStarterInvalid            = errors.New("流程发起人不存在或已停用")
	ErrStarterNotAllowed         = errors.New("该用户不在流程允许的发起人范围内")
	ErrStartNotYetAvailable      = errors.New("流程尚未到允许发起时间")
	ErrStartAvailabilityExpired  = errors.New("流程已超过允许发起时间")
	ErrStartOutsideAvailability  = errors.New("当前不在流程允许发起时间内")
	ErrStarterAccessDenied       = errors.New("无权为该用户发起流程")
	ErrBusinessReferenceRequired = errors.New("业务类型和业务标识不能为空")
	ErrTaskIDRequired            = errors.New("流程任务不能为空")
	ErrActorRequired             = errors.New("任务处理人不能为空")
	ErrInstanceIDRequired        = errors.New("流程实例不能为空")
	ErrInstanceAccessDenied      = errors.New("无权访问该流程实例")
	ErrInstanceScopeInvalid      = errors.New("流程实例查询范围无效")
	ErrDefinitionVersionChanged  = errors.New("流程定义已更新，请重新打开后保存")
	ErrDraftStoreUnavailable     = errors.New("流程草稿存储未初始化")
	ErrNotificationUnavailable   = errors.New("工作流通知服务未初始化")
)

type Store interface {
	UserDepartmentReader
	InTransaction(ctx context.Context, fn func(TransactionStore) error) error
	ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error)
	GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error)
	ListInstances(ctx context.Context, query InstanceQuery) (*InstanceList, error)
	GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error)
	ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error)
	HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error)
}

type TransactionStore interface {
	UserDepartmentReader
	LoadPublishedDefinition(ctx context.Context, definitionID uint, version int) (workflowcore.Definition, int, error)
	IsActiveUser(ctx context.Context, userID string) (bool, error)
	CanOperatorStartFor(ctx context.Context, operatorID, starterID string) (bool, error)
	CreateState(ctx context.Context, state *workflowdomain.State) error
	DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error
	LoadStateByTaskForUpdate(ctx context.Context, taskID string) (workflowcore.Definition, *workflowdomain.State, error)
	LoadStateByInstanceForUpdate(ctx context.Context, instanceID string) (*workflowdomain.State, error)
	LoadDefinitionAndStateByInstanceForUpdate(ctx context.Context, instanceID string) (workflowcore.Definition, *workflowdomain.State, error)
	SaveState(ctx context.Context, state *workflowdomain.State) error
	PersistEffects(ctx context.Context, state *workflowdomain.State) ([]string, error)
}

type StartDraftStore interface {
	GetStartDraft(ctx context.Context, definitionID uint, starterID string) (*StartDraft, error)
	SaveStartDraft(ctx context.Context, draft StartDraft) (*StartDraft, error)
}

type UserDepartmentReader interface {
	UserDepartmentIDs(ctx context.Context, userID string) ([]uint, error)
}

type BusinessStateReader interface {
	FindStateByBusiness(ctx context.Context, businessType, businessKey string) (*workflowdomain.State, bool, error)
}

type Service struct {
	store         Store
	engine        *workflowdomain.Engine
	publisher     EventPublisher
	notifications NotificationDispatcher
	now           func() time.Time
}

func NewService(store Store, resolver workflowdomain.AssigneeResolver, ids workflowdomain.IDGenerator) *Service {
	return NewServiceWithPublisher(store, resolver, ids, DefaultLifecycleEventPublisher())
}

func NewServiceWithPublisher(store Store, resolver workflowdomain.AssigneeResolver, ids workflowdomain.IDGenerator, publisher EventPublisher) *Service {
	return NewServiceWithNotifications(store, resolver, ids, publisher, nil)
}

func NewServiceWithNotifications(
	store Store,
	resolver workflowdomain.AssigneeResolver,
	ids workflowdomain.IDGenerator,
	publisher EventPublisher,
	notifications NotificationDispatcher,
) *Service {
	if publisher == nil {
		publisher = DefaultLifecycleEventPublisher()
	}
	return &Service{
		store: store, engine: workflowdomain.NewEngine(resolver, ids),
		publisher: publisher, notifications: notifications, now: time.Now,
	}
}

type StartInstanceRequest struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	BusinessType      string                 `json:"businessType"`
	BusinessKey       string                 `json:"businessKey"`
	StarterID         string                 `json:"starterId"`
	OperatorID        string                 `json:"operatorId"`
	AdminInitiated    bool                   `json:"-"`
	Idempotent        bool                   `json:"-"`
	ClearStartDraft   bool                   `json:"-"`
	Variables         map[string]interface{} `json:"variables"`
	FormData          map[string]interface{} `json:"formData"`
}

type SaveStartDraftRequest struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	StarterID         string                 `json:"starterId"`
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
	request.OperatorID = strings.TrimSpace(request.OperatorID)
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if request.StarterID == "" {
		return nil, ErrStarterRequired
	}
	if request.OperatorID == "" {
		return nil, ErrOperatorRequired
	}
	if request.BusinessType == "" || request.BusinessKey == "" {
		return nil, ErrBusinessReferenceRequired
	}
	if service == nil || service.store == nil || service.engine == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var state *workflowdomain.State
	var outboxIDs []string
	created := false
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		if request.Idempotent {
			reader, ok := store.(BusinessStateReader)
			if !ok {
				return errors.New("工作流存储不支持业务幂等查询")
			}
			existing, found, err := reader.FindStateByBusiness(ctx, request.BusinessType, request.BusinessKey)
			if err != nil {
				return err
			}
			if found {
				state = existing
				return nil
			}
		}
		definition, publishedVersion, err := store.LoadPublishedDefinition(ctx, request.DefinitionID, request.DefinitionVersion)
		if err != nil {
			return err
		}
		initiator := definitionInitiatorConfig(definition)
		departmentIDs, err := loadInitiatorDepartmentIDs(ctx, store, initiator, request.StarterID)
		if err != nil {
			return err
		}
		if !publishedInitiatorAllows(initiator, request.StarterID, departmentIDs) {
			return ErrStarterNotAllowed
		}
		if err := startAvailabilityError(workflowcore.EvaluateStartAvailability(definitionStartAvailabilityConfig(definition), service.currentTime())); err != nil {
			return err
		}
		active, err := store.IsActiveUser(ctx, request.StarterID)
		if err != nil {
			return err
		}
		if !active {
			return ErrStarterInvalid
		}
		if !request.AdminInitiated && request.OperatorID != request.StarterID {
			return ErrStarterAccessDenied
		}
		if request.AdminInitiated {
			allowed, err := store.CanOperatorStartFor(ctx, request.OperatorID, request.StarterID)
			if err != nil {
				return err
			}
			if !allowed {
				return ErrStarterAccessDenied
			}
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
			OperatorID:        request.OperatorID,
			Variables:         request.Variables,
			FormData:          request.FormData,
		})
		if err != nil {
			return err
		}
		if err := store.CreateState(ctx, state); err != nil {
			return err
		}
		if request.ClearStartDraft {
			if err := store.DeleteStartDraft(ctx, request.DefinitionID, request.StarterID); err != nil {
				return err
			}
		}
		created = true
		outboxIDs, err = store.PersistEffects(ctx, state)
		return err
	})
	if err != nil && request.Idempotent {
		if reader, ok := service.store.(BusinessStateReader); ok {
			if existing, found, lookupErr := reader.FindStateByBusiness(ctx, request.BusinessType, request.BusinessKey); lookupErr == nil && found {
				return existing, nil
			}
		}
	}
	if err == nil && created {
		service.dispatchNotifications(ctx, outboxIDs)
		service.publish(ctx, LifecycleEvent{Type: LifecycleInstanceStarted, InstanceID: state.Instance.ID, ActorID: request.OperatorID,
			BusinessType: state.Instance.BusinessType, BusinessKey: state.Instance.BusinessKey, Status: string(state.Instance.Status)})
	}
	return state, err
}

func (service *Service) GetStartDraft(ctx context.Context, definitionID uint, starterID string) (*StartDraft, error) {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if starterID == "" {
		return nil, ErrStarterRequired
	}
	if _, err := service.GetPublishedDefinitionForStarter(ctx, definitionID, starterID); err != nil {
		return nil, err
	}
	store, err := service.startDraftStore()
	if err != nil {
		return nil, err
	}
	return store.GetStartDraft(ctx, definitionID, starterID)
}

func (service *Service) SaveStartDraft(ctx context.Context, request SaveStartDraftRequest) (*StartDraft, error) {
	request.StarterID = strings.TrimSpace(request.StarterID)
	if request.DefinitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if request.StarterID == "" {
		return nil, ErrStarterRequired
	}
	definition, err := service.GetPublishedDefinitionForStarter(ctx, request.DefinitionID, request.StarterID)
	if err != nil {
		return nil, err
	}
	if request.DefinitionVersion > 0 && request.DefinitionVersion != definition.Version {
		return nil, ErrDefinitionVersionChanged
	}
	if request.FormData == nil {
		request.FormData = make(map[string]interface{})
	}
	if err := workflowcore.ValidateFormData(definition.Form, request.FormData, true); err != nil {
		return nil, err
	}
	store, err := service.startDraftStore()
	if err != nil {
		return nil, err
	}
	return store.SaveStartDraft(ctx, StartDraft{
		DefinitionID: request.DefinitionID, DefinitionVersion: definition.Version,
		StarterID: request.StarterID, FormData: request.FormData,
		UpdatedAt: service.currentTime().UnixMilli(),
	})
}

func (service *Service) startDraftStore() (StartDraftStore, error) {
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	store, ok := service.store.(StartDraftStore)
	if !ok {
		return nil, ErrDraftStoreUnavailable
	}
	return store, nil
}

func definitionInitiatorConfig(definition workflowcore.Definition) workflowcore.InitiatorConfig {
	for _, node := range definition.Nodes {
		if node.Type == workflowcore.NodeTypeStart && node.Initiator != nil {
			return *node.Initiator
		}
	}
	return workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeAll}
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
	var outboxIDs []string
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
		advanced, err = service.engine.ResumeTimers(definition, loaded, time.Now().Unix())
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

func (service *Service) ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error) {
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	definitions, err := service.store.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	return decoratePublishedDefinitionsAvailability(definitions, service.currentTime()), nil
}

func (service *Service) ListPublishedDefinitionsForStarter(ctx context.Context, starterID string) ([]PublishedDefinition, error) {
	definitions, err := service.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	var departmentIDs []uint
	for _, definition := range definitions {
		if initiatorNeedsDepartments(definition.Initiator, starterID) {
			departmentIDs, err = service.store.UserDepartmentIDs(ctx, starterID)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	result := make([]PublishedDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if publishedInitiatorAllows(definition.Initiator, starterID, departmentIDs) {
			result = append(result, definition)
		}
	}
	return result, nil
}

func (service *Service) GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error) {
	if definitionID == 0 {
		return nil, ErrDefinitionRequired
	}
	if service == nil || service.store == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}
	definition, err := service.store.GetPublishedDefinition(ctx, definitionID)
	if err != nil || definition == nil {
		return definition, err
	}
	definition.AvailabilityStatus = workflowcore.EvaluateStartAvailability(&definition.Availability, service.currentTime())
	return definition, nil
}

func (service *Service) GetPublishedDefinitionForStarter(ctx context.Context, definitionID uint, starterID string) (*PublishedDefinition, error) {
	definition, err := service.GetPublishedDefinition(ctx, definitionID)
	if err != nil {
		return nil, err
	}
	departmentIDs, err := loadInitiatorDepartmentIDs(ctx, service.store, definition.Initiator, starterID)
	if err != nil {
		return nil, err
	}
	if !publishedInitiatorAllows(definition.Initiator, starterID, departmentIDs) {
		return nil, ErrStarterNotAllowed
	}
	return definition, nil
}

func publishedInitiatorAllows(initiator workflowcore.InitiatorConfig, starterID string, departmentIDs []uint) bool {
	if initiator.Scope != workflowcore.InitiatorScopeSpecified {
		return true
	}
	for _, userID := range initiator.UserIDs {
		if strings.TrimSpace(starterID) == strconv.FormatUint(uint64(userID), 10) {
			return true
		}
	}
	allowedDepartments := make(map[uint]struct{}, len(initiator.DepartmentIDs))
	for _, departmentID := range initiator.DepartmentIDs {
		allowedDepartments[departmentID] = struct{}{}
	}
	for _, departmentID := range departmentIDs {
		if _, allowed := allowedDepartments[departmentID]; allowed {
			return true
		}
	}
	return false
}

func loadInitiatorDepartmentIDs(
	ctx context.Context,
	reader UserDepartmentReader,
	initiator workflowcore.InitiatorConfig,
	starterID string,
) ([]uint, error) {
	if !initiatorNeedsDepartments(initiator, starterID) {
		return nil, nil
	}
	return reader.UserDepartmentIDs(ctx, starterID)
}

func initiatorNeedsDepartments(initiator workflowcore.InitiatorConfig, starterID string) bool {
	if initiator.Scope != workflowcore.InitiatorScopeSpecified || len(initiator.DepartmentIDs) == 0 {
		return false
	}
	starterID = strings.TrimSpace(starterID)
	for _, userID := range initiator.UserIDs {
		if starterID == strconv.FormatUint(uint64(userID), 10) {
			return false
		}
	}
	return true
}

func definitionStartAvailabilityConfig(definition workflowcore.Definition) *workflowcore.StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return definition.Nodes[index].Availability
		}
	}
	return nil
}

func decoratePublishedDefinitionsAvailability(definitions []PublishedDefinition, now time.Time) []PublishedDefinition {
	result := append([]PublishedDefinition(nil), definitions...)
	for index := range result {
		result[index].AvailabilityStatus = workflowcore.EvaluateStartAvailability(&result[index].Availability, now)
	}
	return result
}

func startAvailabilityError(state string) error {
	switch state {
	case workflowcore.StartAvailabilityStateAvailable:
		return nil
	case workflowcore.StartAvailabilityStateNotStarted:
		return ErrStartNotYetAvailable
	case workflowcore.StartAvailabilityStateExpired:
		return ErrStartAvailabilityExpired
	default:
		return ErrStartOutsideAvailability
	}
}

func (service *Service) currentTime() time.Time {
	if service != nil && service.now != nil {
		return service.now()
	}
	return time.Now()
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
	query.Scope = strings.TrimSpace(query.Scope)
	if query.Scope == "" {
		query.Scope = InstanceScopeStarted
	}
	switch query.Scope {
	case InstanceScopeStarted, InstanceScopeHandled, InstanceScopeCopied:
	default:
		return nil, ErrInstanceScopeInvalid
	}
	query.StarterID = ""
	query.ScopeUserID = actorID
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
	allowed, err := service.store.HasParticipant(ctx, instanceID, actorID, string(workflowdomain.ParticipantRoleCC))
	if err != nil {
		return nil, err
	}
	if allowed {
		return detail, nil
	}
	return nil, ErrInstanceAccessDenied
}

func (service *Service) ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error) {
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.store.ListTasks(ctx, query)
}

func (service *Service) ListNotifications(ctx context.Context, query NotificationQuery) (*NotificationList, error) {
	if service == nil || service.notifications == nil {
		return nil, ErrNotificationUnavailable
	}
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)
	return service.notifications.List(ctx, query)
}

func (service *Service) RetryNotification(ctx context.Context, id string) error {
	if service == nil || service.notifications == nil {
		return ErrNotificationUnavailable
	}
	return service.notifications.Retry(ctx, strings.TrimSpace(id))
}

func (service *Service) DispatchDueNotifications(ctx context.Context, limit int) (int, error) {
	if service == nil || service.notifications == nil {
		return 0, ErrNotificationUnavailable
	}
	return service.notifications.DispatchDue(ctx, limit)
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

func (service *Service) dispatchNotifications(ctx context.Context, ids []string) {
	if service == nil || service.notifications == nil || len(ids) == 0 {
		return
	}
	_, _ = service.notifications.Dispatch(ctx, ids)
}
