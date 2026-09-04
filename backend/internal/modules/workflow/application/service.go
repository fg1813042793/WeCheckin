package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

var (
	ErrDefinitionRequired                    = errors.New("流程定义不能为空")
	ErrStarterRequired                       = errors.New("流程发起人不能为空")
	ErrOperatorRequired                      = errors.New("流程发起操作人不能为空")
	ErrStarterInvalid                        = errors.New("流程发起人不存在或已停用")
	ErrStarterNotAllowed                     = errors.New("该用户不在流程允许的发起人范围内")
	ErrStartNotYetAvailable                  = errors.New("流程尚未到允许发起时间")
	ErrStartAvailabilityExpired              = errors.New("流程已超过允许发起时间")
	ErrStartOutsideAvailability              = errors.New("当前不在流程允许发起时间内")
	ErrStartLimitExceeded                    = errors.New("当前周期的流程发起次数已用完")
	ErrStarterAccessDenied                   = errors.New("无权为该用户发起流程")
	ErrBusinessReferenceRequired             = errors.New("业务类型和业务标识不能为空")
	ErrTaskIDRequired                        = errors.New("流程任务不能为空")
	ErrTaskDeleteNotAllowed                  = errors.New("待激活或待处理的流程任务不能删除")
	ErrTaskDeleteTargetNotFound              = errors.New("流程任务不存在或已删除")
	ErrActorRequired                         = errors.New("任务处理人不能为空")
	ErrInstanceIDRequired                    = errors.New("流程实例不能为空")
	ErrInstanceAccessDenied                  = errors.New("无权访问该流程实例")
	ErrInstanceScopeInvalid                  = errors.New("流程实例查询范围无效")
	ErrDefinitionNameSearchTooLong           = errors.New("流程名称关键字不能超过50个字符")
	ErrStarterNameSearchTooLong              = errors.New("发起人用户名关键字不能超过50个字符")
	ErrRunningInstanceCannotDelete           = errors.New("审批中的申请不能删除，请先撤回")
	ErrInstanceDeleteNotAllowed              = errors.New("当前流程状态不能删除")
	ErrInstanceDeleteTargetNotFound          = errors.New("部分流程实例不存在或已删除")
	ErrInstanceDeleteTooMany                 = errors.New("单次最多删除100个流程实例")
	ErrInstanceCommentRequired               = errors.New("流程评论不能为空")
	ErrInstanceCommentTooLong                = errors.New("流程评论不能超过500个字符")
	ErrCommentNotificationRecipientsRequired = errors.New("请选择评论通知对象")
	ErrCommentNotificationRecipientsTooMany  = errors.New("评论通知对象不能超过100人")
	ErrCommentNotificationChannelsRequired   = errors.New("请选择评论通知方式")
	ErrCommentNotificationChannelInvalid     = errors.New("评论通知方式无效")
	ErrCommentNotificationRecipientDenied    = errors.New("评论通知对象不是该流程参与人")
	ErrTaskRejectCommentRequired             = errors.New("驳回原因不能为空")
	ErrTaskReturnCommentRequired             = errors.New("退回原因不能为空")
	ErrTaskCommentTooLong                    = errors.New("处理意见不能超过500个字符")
	ErrReminderNodeRequired                  = errors.New("催办节点不能为空")
	ErrReminderStarterOnly                   = errors.New("只有流程发起人可以提醒处理")
	ErrReminderNodeUnavailable               = errors.New("当前节点没有可提醒的待处理任务")
	ErrReminderCooldown                      = errors.New("该节点刚刚提醒过，请稍后再试")
	ErrReminderDailyLimit                    = errors.New("该节点今天的提醒次数已用完")
	ErrDefinitionVersionChanged              = errors.New("流程定义已更新，请重新打开后保存")
	ErrDraftStoreUnavailable                 = errors.New("流程草稿存储未初始化")
	ErrNotificationUnavailable               = errors.New("工作流通知服务未初始化")
)

const (
	reminderCooldown   = 30 * time.Minute
	reminderDailyLimit = 3
)

type Store interface {
	UserDepartmentReader
	InTransaction(ctx context.Context, fn func(TransactionStore) error) error
	ListPublishedDefinitions(ctx context.Context) ([]PublishedDefinition, error)
	GetPublishedDefinition(ctx context.Context, definitionID uint) (*PublishedDefinition, error)
	CountStartQuotaUsage(ctx context.Context, definitionID uint, starterID string, window workflowcore.StartLimitWindow) (int, error)
	ListInstances(ctx context.Context, query InstanceQuery) (*InstanceList, error)
	GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error)
	HideStartedInstance(ctx context.Context, instanceID, starterID string, deletedAt int64) error
	ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error)
	HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error)
}

type TransactionStore interface {
	UserDepartmentReader
	LoadPublishedDefinition(ctx context.Context, definitionID uint, version int) (workflowcore.Definition, int, error)
	IsActiveUser(ctx context.Context, userID string) (bool, error)
	CanOperatorStartFor(ctx context.Context, operatorID, starterID string) (bool, error)
	ConsumeStartQuota(ctx context.Context, definitionID uint, starterID string, window workflowcore.StartLimitWindow, maxCount int) (int, bool, error)
	CreateState(ctx context.Context, state *workflowdomain.State) error
	DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error
	LoadStateByTaskForUpdate(ctx context.Context, taskID string) (workflowcore.Definition, *workflowdomain.State, error)
	LoadStateByInstanceForUpdate(ctx context.Context, instanceID string) (*workflowdomain.State, error)
	LoadDefinitionAndStateByInstanceForUpdate(ctx context.Context, instanceID string) (workflowcore.Definition, *workflowdomain.State, error)
	SaveState(ctx context.Context, state *workflowdomain.State) error
	AppendInstanceHistory(ctx context.Context, instanceID string, event workflowdomain.HistoryEvent, eventTime int64) error
	PersistEffects(ctx context.Context, state *workflowdomain.State) ([]string, error)
	LoadInstancesForDelete(ctx context.Context, instanceIDs []string) ([]InstanceSummary, error)
	SoftDeleteInstances(ctx context.Context, instanceIDs []string, actorID string, deletedAt int64) (int64, error)
	LoadTaskForDelete(ctx context.Context, taskID string) (*TaskSummary, error)
	SoftDeleteTask(ctx context.Context, taskID, actorID string, deletedAt int64) (int64, error)
}

type StartDraftStore interface {
	GetStartDraft(ctx context.Context, definitionID uint, starterID string) (*StartDraft, error)
	SaveStartDraft(ctx context.Context, draft StartDraft) (*StartDraft, error)
	DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error
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
	resolver      workflowdomain.AssigneeResolver
	publisher     EventPublisher
	notifications NotificationDispatcher
	ids           workflowdomain.IDGenerator
	now           func() time.Time
}

type assigneeDisplayResolver interface {
	ResolveDisplayNames(context.Context, workflowdomain.AssigneeRequest) ([]string, error)
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
		store: store, engine: workflowdomain.NewEngine(resolver, ids), resolver: resolver,
		publisher: publisher, notifications: notifications, ids: ids, now: time.Now,
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
	TaskID             string                        `json:"taskId"`
	ActorID            string                        `json:"actorId"`
	Action             workflowdomain.TaskAction     `json:"action"`
	Comment            string                        `json:"comment"`
	Images             []workflowcore.FormAttachment `json:"images"`
	ReturnTargetNodeID string                        `json:"returnTargetNodeId"`
	Variables          map[string]interface{}        `json:"variables"`
	FormData           map[string]interface{}        `json:"formData"`
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

type CommentInstanceRequest struct {
	InstanceID   string                        `json:"instanceId"`
	ActorID      string                        `json:"actorId"`
	Comment      string                        `json:"comment"`
	Images       []workflowcore.FormAttachment `json:"images"`
	Notification *CommentNotificationRequest   `json:"notification,omitempty"`
}

type CommentNotificationRequest struct {
	UserIDs  []string `json:"userIds"`
	Channels []string `json:"channels"`
}

type RemindInstanceRequest struct {
	InstanceID string `json:"instanceId"`
	ActorID    string `json:"actorId"`
	NodeID     string `json:"nodeId"`
}

type RemindInstanceResult struct {
	NodeID         string `json:"nodeId"`
	NodeName       string `json:"nodeName"`
	RemindedCount  int    `json:"remindedCount"`
	RemindedAt     int64  `json:"remindedAt"`
	NextAllowedAt  int64  `json:"nextAllowedAt"`
	RemainingCount int    `json:"remainingCount"`
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
		startedAt := service.currentTime()
		availability := definitionStartAvailabilityConfig(definition)
		if err := startAvailabilityError(workflowcore.EvaluateStartAvailability(availability, startedAt)); err != nil {
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
		if err := workflowcore.ValidateStartFormData(definition, request.FormData); err != nil {
			return err
		}
		startLimit := definitionStartLimitConfig(definition)
		if startLimit.Mode == workflowcore.StartLimitModeLimited {
			window, ok := workflowcore.ResolveStartLimitWindow(&startLimit, availability, startedAt)
			if !ok {
				return errors.New("流程发起次数限制配置无效")
			}
			_, allowed, err := store.ConsumeStartQuota(ctx, request.DefinitionID, request.StarterID, window, startLimit.MaxCount)
			if err != nil {
				return err
			}
			if !allowed {
				return ErrStartLimitExceeded
			}
		}
		state, err = service.engine.Start(ctx, definition, workflowdomain.StartRequest{
			DefinitionID:      request.DefinitionID,
			DefinitionVersion: publishedVersion,
			BusinessType:      request.BusinessType,
			BusinessKey:       request.BusinessKey,
			StarterID:         request.StarterID,
			OperatorID:        request.OperatorID,
			StartTime:         startedAt.UnixMilli(),
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

func (service *Service) DeleteStartDraft(ctx context.Context, definitionID uint, starterID string) error {
	starterID = strings.TrimSpace(starterID)
	if definitionID == 0 {
		return ErrDefinitionRequired
	}
	if starterID == "" {
		return ErrStarterRequired
	}
	store, err := service.startDraftStore()
	if err != nil {
		return err
	}
	return store.DeleteStartDraft(ctx, definitionID, starterID)
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
		if err := service.engine.Complete(ctx, definition, loaded, workflowdomain.CompleteRequest{
			TaskID: request.TaskID, ActorID: request.ActorID, Action: request.Action,
			Comment: request.Comment, Images: request.Images, ReturnTargetNodeID: request.ReturnTargetNodeID,
			Variables: request.Variables, FormData: request.FormData,
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

func (service *Service) ListPublishedDefinitionCategories(ctx context.Context) ([]string, error) {
	definitions, err := service.ListPublishedDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(definitions))
	categories := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		category := strings.TrimSpace(definition.Category)
		if category == "" {
			continue
		}
		if _, exists := seen[category]; exists {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories, nil
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
	if err := service.decorateStartLimitStatuses(ctx, result, starterID); err != nil {
		return nil, err
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
	value := *definition
	value.Nodes = append([]PublishedNode(nil), definition.Nodes...)
	definition = &value
	if err := service.resolvePublishedAssigneeDisplays(ctx, definition, starterID); err != nil {
		return nil, err
	}
	definitions := []PublishedDefinition{*definition}
	if err := service.decorateStartLimitStatuses(ctx, definitions, starterID); err != nil {
		return nil, err
	}
	*definition = definitions[0]
	return definition, nil
}

func (service *Service) resolvePublishedAssigneeDisplays(ctx context.Context, definition *PublishedDefinition, starterID string) error {
	if definition == nil {
		return nil
	}
	instance := workflowdomain.ProcessInstance{StarterID: starterID, OperatorID: starterID}
	return service.resolveNodeAssigneeDisplays(ctx, definition.Nodes, instance)
}

func (service *Service) resolveNodeAssigneeDisplays(ctx context.Context, nodes []PublishedNode, instance workflowdomain.ProcessInstance) error {
	resolver, ok := service.resolver.(assigneeDisplayResolver)
	if !ok {
		return nil
	}
	for index := range nodes {
		node := &nodes[index]
		if node.Assignee == nil {
			continue
		}
		names, err := resolver.ResolveDisplayNames(ctx, workflowdomain.AssigneeRequest{
			Instance: instance,
			Node: workflowcore.Node{
				ID: node.ID, Type: node.Type, Name: node.Name, ApprovalMode: node.ApprovalMode,
				GatewayMode: node.GatewayMode, Assignee: node.Assignee,
			},
		})
		if err != nil {
			return err
		}
		resolved := make([]string, 0, len(names))
		for _, name := range names {
			if value := strings.TrimSpace(name); value != "" {
				resolved = append(resolved, value)
			}
		}
		if len(resolved) > 0 {
			node.AssigneeDisplay = strings.Join(resolved, "、")
		}
	}
	return nil
}

func publishedInitiatorAllows(initiator workflowcore.InitiatorConfig, starterID string, departmentIDs []uint) bool {
	if initiatorContainsUser(initiator.ExcludedUserIDs, starterID) {
		return false
	}
	if initiator.Scope != workflowcore.InitiatorScopeSpecified {
		return true
	}
	if initiatorContainsUser(initiator.UserIDs, starterID) {
		return true
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
	if initiatorContainsUser(initiator.ExcludedUserIDs, starterID) {
		return false
	}
	return !initiatorContainsUser(initiator.UserIDs, starterID)
}

func initiatorContainsUser(userIDs []uint, starterID string) bool {
	starterID = strings.TrimSpace(starterID)
	for _, userID := range userIDs {
		if starterID == strconv.FormatUint(uint64(userID), 10) {
			return true
		}
	}
	return false
}

func definitionStartAvailabilityConfig(definition workflowcore.Definition) *workflowcore.StartAvailabilityConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return definition.Nodes[index].Availability
		}
	}
	return nil
}

func definitionStartLimitConfig(definition workflowcore.Definition) workflowcore.StartLimitConfig {
	for index := range definition.Nodes {
		if definition.Nodes[index].Type == workflowcore.NodeTypeStart {
			return workflowcore.CloneStartLimit(definition.Nodes[index].StartLimit)
		}
	}
	return workflowcore.DefaultStartLimit()
}

func (service *Service) decorateStartLimitStatuses(ctx context.Context, definitions []PublishedDefinition, starterID string) error {
	now := service.currentTime()
	for index := range definitions {
		definition := &definitions[index]
		definition.StartLimitStatus = StartLimitStatus{Allowed: true}
		if definition.StartLimit.Mode != workflowcore.StartLimitModeLimited {
			continue
		}
		window, ok := workflowcore.ResolveStartLimitWindow(&definition.StartLimit, &definition.Availability, now)
		if !ok {
			return errors.New("流程发起次数限制配置无效")
		}
		usedCount, err := service.store.CountStartQuotaUsage(ctx, definition.ID, starterID, window)
		if err != nil {
			return err
		}
		remainingCount := definition.StartLimit.MaxCount - usedCount
		if remainingCount < 0 {
			remainingCount = 0
		}
		definition.StartLimitStatus = StartLimitStatus{
			Allowed: usedCount < definition.StartLimit.MaxCount, UsedCount: usedCount,
			RemainingCount: remainingCount, ResetsAt: window.EndsAt,
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
	detail, err := service.store.GetInstance(ctx, instanceID)
	if err != nil || detail == nil {
		return detail, err
	}
	value := *detail
	value.Nodes = append([]PublishedNode(nil), detail.Nodes...)
	value.History = append([]HistorySummary(nil), detail.History...)
	detail = &value
	applyTaskCreatedActors(detail)
	applyHistoricalCommentNodeIDs(detail)
	if err := service.resolveNodeAssigneeDisplays(ctx, detail.Nodes, workflowdomain.ProcessInstance{
		StarterID:  detail.Instance.StarterID,
		OperatorID: detail.Instance.OperatorID,
	}); err != nil {
		return nil, err
	}
	applyTaskAssigneeDisplays(detail.Nodes, detail.Tasks)
	return detail, nil
}

type historyActor struct {
	id   string
	name string
}

func applyTaskCreatedActors(detail *InstanceDetail) {
	if detail == nil || len(detail.History) == 0 {
		return
	}
	current := historyActor{
		id:   strings.TrimSpace(detail.Instance.OperatorID),
		name: strings.TrimSpace(detail.Instance.OperatorName),
	}
	if current.id == "" {
		current.id = strings.TrimSpace(detail.Instance.StarterID)
		current.name = strings.TrimSpace(detail.Instance.StarterName)
	}
	indexes := make([]int, len(detail.History))
	for index := range detail.History {
		indexes[index] = index
	}
	sort.SliceStable(indexes, func(left, right int) bool {
		return detail.History[indexes[left]].EventTime < detail.History[indexes[right]].EventTime
	})
	for start := 0; start < len(indexes); {
		end := start + 1
		eventTime := detail.History[indexes[start]].EventTime
		for end < len(indexes) && detail.History[indexes[end]].EventTime == eventTime {
			end++
		}
		trigger := historyActor{}
		for _, index := range indexes[start:end] {
			event := detail.History[index]
			if isTaskCreationTrigger(event.EventType) && strings.TrimSpace(event.ActorID) != "" {
				trigger = historyActor{id: strings.TrimSpace(event.ActorID), name: historyActorName(detail, event.ActorID, event.ActorName)}
			}
		}
		actor := current
		if trigger.id != "" {
			actor = trigger
		}
		if actor.id != "" {
			for _, index := range indexes[start:end] {
				if detail.History[index].EventType == string(workflowdomain.HistoryTaskCreated) {
					detail.History[index].ActorID = actor.id
					detail.History[index].ActorName = actor.name
				}
			}
		}
		if trigger.id != "" {
			current = trigger
		}
		start = end
	}
}

func isTaskCreationTrigger(eventType string) bool {
	switch workflowdomain.HistoryEventType(eventType) {
	case workflowdomain.HistoryInstanceStarted,
		workflowdomain.HistoryTaskApproved,
		workflowdomain.HistoryTaskReturned,
		workflowdomain.HistoryTaskSubmitted:
		return true
	default:
		return false
	}
}

func historyActorName(detail *InstanceDetail, actorID, actorName string) string {
	if name := strings.TrimSpace(actorName); name != "" {
		return name
	}
	actorID = strings.TrimSpace(actorID)
	if name := strings.TrimSpace(detail.UserNames[actorID]); name != "" {
		return name
	}
	if actorID == strings.TrimSpace(detail.Instance.OperatorID) {
		return strings.TrimSpace(detail.Instance.OperatorName)
	}
	if actorID == strings.TrimSpace(detail.Instance.StarterID) {
		return strings.TrimSpace(detail.Instance.StarterName)
	}
	return ""
}

func applyHistoricalCommentNodeIDs(detail *InstanceDetail) {
	if detail == nil || len(detail.History) == 0 {
		return
	}
	assigneesByTask := make(map[string]string, len(detail.Tasks))
	for _, task := range detail.Tasks {
		if taskID := strings.TrimSpace(task.ID); taskID != "" {
			assigneesByTask[taskID] = strings.TrimSpace(task.AssigneeID)
		}
	}
	indexes := make([]int, len(detail.History))
	for index := range detail.History {
		indexes[index] = index
	}
	sort.SliceStable(indexes, func(left, right int) bool {
		return detail.History[indexes[left]].EventTime < detail.History[indexes[right]].EventTime
	})
	activeTasks := make(map[string]string)
	for _, index := range indexes {
		event := &detail.History[index]
		taskID := strings.TrimSpace(event.TaskID)
		nodeID := strings.TrimSpace(event.NodeID)
		switch workflowdomain.HistoryEventType(event.EventType) {
		case workflowdomain.HistoryTaskCreated, workflowdomain.HistoryTaskActivated:
			if taskID != "" && nodeID != "" {
				activeTasks[taskID] = nodeID
			}
		case workflowdomain.HistoryTaskApproved,
			workflowdomain.HistoryTaskRejected,
			workflowdomain.HistoryTaskReturned,
			workflowdomain.HistoryTaskSubmitted,
			workflowdomain.HistoryTaskCancelled:
			delete(activeTasks, taskID)
		case workflowdomain.HistoryInstanceCompleted,
			workflowdomain.HistoryInstanceRejected,
			workflowdomain.HistoryInstanceWithdrawn,
			workflowdomain.HistoryInstanceCancelled:
			clear(activeTasks)
		case workflowdomain.HistoryInstanceCommented:
			if nodeID == "" {
				event.NodeID = commentNodeFromActiveTasks(activeTasks, assigneesByTask, event.ActorID)
			}
		}
	}
}

func currentCommentNodeID(detail *InstanceDetail, actorID string) string {
	if detail == nil {
		return ""
	}
	activeTasks := make(map[string]string)
	assigneesByTask := make(map[string]string)
	for _, task := range detail.Tasks {
		if task.Status != string(workflowdomain.TaskStatusPending) && task.Status != string(workflowdomain.TaskStatusWaiting) {
			continue
		}
		taskID := strings.TrimSpace(task.ID)
		nodeID := strings.TrimSpace(task.NodeID)
		if taskID == "" || nodeID == "" {
			continue
		}
		activeTasks[taskID] = nodeID
		assigneesByTask[taskID] = strings.TrimSpace(task.AssigneeID)
	}
	if nodeID := commentNodeFromActiveTasks(activeTasks, assigneesByTask, actorID); nodeID != "" {
		return nodeID
	}
	processingNodes := make(map[string]struct{})
	for _, node := range detail.NodeProgress {
		if node.Status == NodeProgressProcessing {
			if nodeID := strings.TrimSpace(node.NodeID); nodeID != "" {
				processingNodes[nodeID] = struct{}{}
			}
		}
	}
	return uniqueNodeID(processingNodes)
}

func commentNodeFromActiveTasks(activeTasks, assigneesByTask map[string]string, actorID string) string {
	actorID = strings.TrimSpace(actorID)
	actorNodes := make(map[string]struct{})
	activeNodes := make(map[string]struct{})
	for taskID, nodeID := range activeTasks {
		nodeID = strings.TrimSpace(nodeID)
		if nodeID == "" {
			continue
		}
		activeNodes[nodeID] = struct{}{}
		if actorID != "" && strings.TrimSpace(assigneesByTask[taskID]) == actorID {
			actorNodes[nodeID] = struct{}{}
		}
	}
	if nodeID := uniqueNodeID(actorNodes); nodeID != "" {
		return nodeID
	}
	return uniqueNodeID(activeNodes)
}

func uniqueNodeID(nodes map[string]struct{}) string {
	if len(nodes) != 1 {
		return ""
	}
	for nodeID := range nodes {
		return nodeID
	}
	return ""
}

func applyTaskAssigneeDisplays(nodes []PublishedNode, tasks []TaskSummary) {
	namesByNode := make(map[string][]string)
	seenByNode := make(map[string]map[string]struct{})
	for _, task := range tasks {
		nodeID := strings.TrimSpace(task.NodeID)
		name := strings.TrimSpace(task.HandledByName)
		if name == "" {
			name = strings.TrimSpace(task.AssigneeName)
		}
		if nodeID == "" || name == "" {
			continue
		}
		if seenByNode[nodeID] == nil {
			seenByNode[nodeID] = make(map[string]struct{})
		}
		if _, exists := seenByNode[nodeID][name]; exists {
			continue
		}
		seenByNode[nodeID][name] = struct{}{}
		namesByNode[nodeID] = append(namesByNode[nodeID], name)
	}
	for index := range nodes {
		node := &nodes[index]
		if node.Assignee != nil && node.Assignee.Type == workflowcore.AssigneeTypeInitiator {
			continue
		}
		if names := namesByNode[node.ID]; len(names) > 0 {
			node.AssigneeDisplay = strings.Join(names, "、")
		}
	}
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
	allowed := detail.Instance.StarterID == actorID
	if !allowed {
		for _, task := range detail.Tasks {
			if task.AssigneeID == actorID {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		allowed, err = service.store.HasParticipant(ctx, instanceID, actorID, string(workflowdomain.ParticipantRoleCC))
		if err != nil {
			return nil, err
		}
	}
	if !allowed {
		return nil, ErrInstanceAccessDenied
	}
	decorateInstanceReminders(detail, actorID, service.currentTime())
	return detail, nil
}

func (service *Service) RemindInstance(ctx context.Context, request RemindInstanceRequest) (*RemindInstanceResult, error) {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.NodeID = strings.TrimSpace(request.NodeID)
	if request.InstanceID == "" {
		return nil, ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return nil, ErrActorRequired
	}
	if request.NodeID == "" {
		return nil, ErrReminderNodeRequired
	}
	if service == nil || service.store == nil || service.ids == nil {
		return nil, errors.New("工作流应用服务未初始化")
	}

	var result *RemindInstanceResult
	var outboxIDs []string
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		definition, state, err := store.LoadDefinitionAndStateByInstanceForUpdate(ctx, request.InstanceID)
		if err != nil {
			return err
		}
		if state == nil || state.Instance.Status != workflowdomain.InstanceStatusRunning {
			return workflowdomain.ErrInstanceNotRunning
		}
		if strings.TrimSpace(state.Instance.StarterID) != request.ActorID {
			return ErrReminderStarterOnly
		}
		node, ok := reminderDefinitionNode(definition, request.NodeID)
		if !ok || (node.Type != workflowcore.NodeTypeApproval && node.Type != workflowcore.NodeTypeHandle) {
			return ErrReminderNodeUnavailable
		}
		recipients := reminderRecipients(state.Tasks, request.NodeID, request.ActorID)
		if len(recipients) == 0 {
			return ErrReminderNodeUnavailable
		}

		now := service.currentTime()
		nowMillis := now.UnixMilli()
		lastRemindedAt, todayCount := reminderUsage(state.History, request.NodeID, now)
		if lastRemindedAt > 0 && nowMillis < lastRemindedAt+reminderCooldown.Milliseconds() {
			return ErrReminderCooldown
		}
		if todayCount >= reminderDailyLimit {
			return ErrReminderDailyLimit
		}

		reminderID := service.ids.NewID("reminder")
		state.History = append(state.History, workflowdomain.HistoryEvent{
			ID: reminderID, Type: workflowdomain.HistoryInstanceReminded, NodeID: node.ID,
			ActorID: request.ActorID, Message: fmt.Sprintf("已提醒 %d 位处理人处理“%s”", len(recipients), node.Name), EventTime: nowMillis,
		})
		config := reminderNotificationConfig(node)
		for _, recipient := range recipients {
			state.NotificationIntents = append(state.NotificationIntents, workflowdomain.NotificationIntent{
				ID: service.ids.NewID("notification"), Kind: workflowdomain.NotificationKindTaskReminder,
				NodeID: node.ID, NodeName: node.Name, TaskID: recipient.TaskID,
				RecipientUserID: recipient.UserID, WorkflowName: definition.Name,
				Config: config, DedupeKeySuffix: reminderID,
			})
		}
		if err := store.SaveState(ctx, state); err != nil {
			return err
		}
		outboxIDs, err = store.PersistEffects(ctx, state)
		if err != nil {
			return err
		}
		result = &RemindInstanceResult{
			NodeID: node.ID, NodeName: node.Name, RemindedCount: len(recipients),
			RemindedAt: nowMillis, NextAllowedAt: now.Add(reminderCooldown).UnixMilli(),
			RemainingCount: reminderDailyLimit - todayCount - 1,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	service.dispatchNotifications(ctx, outboxIDs)
	return result, nil
}

type reminderRecipient struct {
	UserID string
	TaskID string
}

func reminderRecipients(tasks []workflowdomain.Task, nodeID, actorID string) []reminderRecipient {
	seen := make(map[string]struct{})
	result := make([]reminderRecipient, 0)
	for _, task := range tasks {
		userID := strings.TrimSpace(task.AssigneeID)
		if task.Status != workflowdomain.TaskStatusPending || task.NodeID != nodeID || userID == "" || userID == actorID {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		result = append(result, reminderRecipient{UserID: userID, TaskID: task.ID})
	}
	return result
}

func reminderDefinitionNode(definition workflowcore.Definition, nodeID string) (workflowcore.Node, bool) {
	for _, node := range definition.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return workflowcore.Node{}, false
}

func reminderUsage(history []workflowdomain.HistoryEvent, nodeID string, now time.Time) (int64, int) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).UnixMilli()
	lastRemindedAt := int64(0)
	todayCount := 0
	for _, event := range history {
		if event.Type != workflowdomain.HistoryInstanceReminded || event.NodeID != nodeID {
			continue
		}
		if event.EventTime > lastRemindedAt {
			lastRemindedAt = event.EventTime
		}
		if event.EventTime >= dayStart && event.EventTime < dayEnd {
			todayCount++
		}
	}
	return lastRemindedAt, todayCount
}

func reminderNotificationConfig(node workflowcore.Node) workflowcore.NotificationConfig {
	channels := make([]string, 0, 2)
	seen := make(map[string]struct{})
	if node.Notification != nil {
		for _, channel := range node.Notification.Channels {
			channel = strings.TrimSpace(channel)
			if channel == "" {
				continue
			}
			if _, exists := seen[channel]; exists {
				continue
			}
			seen[channel] = struct{}{}
			channels = append(channels, channel)
		}
	}
	if len(channels) == 0 {
		channels = append(channels, workflowcore.NotificationChannelDingTalkOA)
	}
	return workflowcore.NotificationConfig{
		Enabled: true, Channels: channels, Title: "流程处理提醒",
		Content: "{{starterName}} 提醒你处理《{{workflowName}}》中的“{{nodeName}}”",
	}
}

func decorateInstanceReminders(detail *InstanceDetail, actorID string, now time.Time) {
	detail.ReminderPolicy = ReminderPolicy{CooldownSeconds: int(reminderCooldown.Seconds()), DailyLimit: reminderDailyLimit}
	detail.ReminderNodes = []ReminderNodeSummary{}
	if detail.Instance.StarterID != actorID || detail.Instance.Status != string(workflowdomain.InstanceStatusRunning) {
		return
	}

	type reminderGroup struct {
		nodeName string
		userIDs  map[string]struct{}
		names    []string
	}
	groups := make(map[string]*reminderGroup)
	order := make([]string, 0)
	for _, task := range detail.Tasks {
		userID := strings.TrimSpace(task.AssigneeID)
		if task.Status != string(workflowdomain.TaskStatusPending) || userID == "" || userID == actorID {
			continue
		}
		group := groups[task.NodeID]
		if group == nil {
			group = &reminderGroup{nodeName: strings.TrimSpace(task.NodeName), userIDs: make(map[string]struct{})}
			groups[task.NodeID] = group
			order = append(order, task.NodeID)
		}
		if _, exists := group.userIDs[userID]; exists {
			continue
		}
		group.userIDs[userID] = struct{}{}
		name := strings.TrimSpace(task.AssigneeName)
		if name != "" {
			group.names = append(group.names, name)
		}
	}

	nowMillis := now.UnixMilli()
	for _, nodeID := range order {
		group := groups[nodeID]
		lastRemindedAt, todayCount := reminderSummaryUsage(detail.History, nodeID, now)
		nextAllowedAt := int64(0)
		blockedReason := ""
		canRemind := true
		if lastRemindedAt > 0 {
			nextAllowedAt = lastRemindedAt + reminderCooldown.Milliseconds()
			if nowMillis < nextAllowedAt {
				canRemind = false
				blockedReason = "cooldown"
			}
		}
		if todayCount >= reminderDailyLimit {
			canRemind = false
			blockedReason = "daily_limit"
		}
		remainingCount := reminderDailyLimit - todayCount
		if remainingCount < 0 {
			remainingCount = 0
		}
		nodeName := group.nodeName
		if nodeName == "" {
			nodeName = "当前节点"
		}
		detail.ReminderNodes = append(detail.ReminderNodes, ReminderNodeSummary{
			NodeID: nodeID, NodeName: nodeName, AssigneeNames: append([]string(nil), group.names...),
			AssigneeCount: len(group.userIDs), CanRemind: canRemind, BlockedReason: blockedReason,
			LastRemindedAt: lastRemindedAt, NextAllowedAt: nextAllowedAt,
			TodayCount: todayCount, RemainingCount: remainingCount,
		})
	}
}

func reminderSummaryUsage(history []HistorySummary, nodeID string, now time.Time) (int64, int) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).UnixMilli()
	lastRemindedAt := int64(0)
	todayCount := 0
	for _, event := range history {
		if event.EventType != string(workflowdomain.HistoryInstanceReminded) || event.NodeID != nodeID {
			continue
		}
		if event.EventTime > lastRemindedAt {
			lastRemindedAt = event.EventTime
		}
		if event.EventTime >= dayStart && event.EventTime < dayEnd {
			todayCount++
		}
	}
	return lastRemindedAt, todayCount
}

func (service *Service) CommentInstance(ctx context.Context, request CommentInstanceRequest) error {
	request.InstanceID = strings.TrimSpace(request.InstanceID)
	request.ActorID = strings.TrimSpace(request.ActorID)
	request.Comment = strings.TrimSpace(request.Comment)
	if request.InstanceID == "" {
		return ErrInstanceIDRequired
	}
	if request.ActorID == "" {
		return ErrActorRequired
	}
	images, err := normalizeWorkflowImages(request.Images)
	if err != nil {
		return err
	}
	request.Images = images
	if request.Comment == "" && len(request.Images) == 0 {
		return ErrInstanceCommentRequired
	}
	if utf8.RuneCountInString(request.Comment) > 500 {
		return ErrInstanceCommentTooLong
	}
	if service == nil || service.store == nil || service.ids == nil {
		return errors.New("工作流应用服务未初始化")
	}
	detail, err := service.GetMyInstance(ctx, request.ActorID, request.InstanceID)
	if err != nil {
		return err
	}
	notification, err := service.normalizeCommentNotification(ctx, detail, request.ActorID, request.Notification)
	if err != nil {
		return err
	}
	historyID := service.ids.NewID("history")
	nodeID := currentCommentNodeID(detail, request.ActorID)
	eventTime := service.currentTime().UnixMilli()
	event := workflowdomain.HistoryEvent{
		ID:      historyID,
		Type:    workflowdomain.HistoryInstanceCommented,
		NodeID:  nodeID,
		ActorID: request.ActorID,
		Message: request.Comment,
		Images:  request.Images,
	}

	var outboxIDs []string
	err = service.store.InTransaction(ctx, func(store TransactionStore) error {
		if err := store.AppendInstanceHistory(ctx, request.InstanceID, event, eventTime); err != nil {
			return err
		}
		if notification == nil {
			return nil
		}
		effects := commentNotificationEffects(detail, event, *notification, service.ids)
		outboxIDs, err = store.PersistEffects(ctx, effects)
		return err
	})
	if err != nil {
		return err
	}
	service.dispatchNotifications(ctx, outboxIDs)
	return nil
}

func (service *Service) DeleteMyInstance(ctx context.Context, actorID, instanceID string) error {
	actorID = strings.TrimSpace(actorID)
	instanceID = strings.TrimSpace(instanceID)
	if actorID == "" {
		return ErrActorRequired
	}
	if instanceID == "" {
		return ErrInstanceIDRequired
	}
	if service == nil || service.store == nil {
		return errors.New("工作流应用服务未初始化")
	}
	detail, err := service.store.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(detail.Instance.StarterID) != actorID {
		return ErrInstanceAccessDenied
	}
	switch detail.Instance.Status {
	case string(workflowdomain.InstanceStatusRunning):
		return ErrRunningInstanceCannotDelete
	case string(workflowdomain.InstanceStatusCompleted),
		string(workflowdomain.InstanceStatusRejected),
		string(workflowdomain.InstanceStatusCancelled),
		"withdrawn":
		return service.store.HideStartedInstance(ctx, instanceID, actorID, service.currentTime().UnixMilli())
	default:
		return ErrInstanceDeleteNotAllowed
	}
}

func (service *Service) DeleteInstances(ctx context.Context, actorID string, instanceIDs []string) (int, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return 0, ErrActorRequired
	}
	instanceIDs = uniqueTrimmedStrings(instanceIDs)
	if len(instanceIDs) == 0 {
		return 0, ErrInstanceIDRequired
	}
	if len(instanceIDs) > 100 {
		return 0, ErrInstanceDeleteTooMany
	}
	if service == nil || service.store == nil {
		return 0, errors.New("工作流应用服务未初始化")
	}

	deleted := 0
	err := service.store.InTransaction(ctx, func(store TransactionStore) error {
		instances, err := store.LoadInstancesForDelete(ctx, instanceIDs)
		if err != nil {
			return err
		}
		if len(instances) != len(instanceIDs) {
			return ErrInstanceDeleteTargetNotFound
		}
		for _, instance := range instances {
			switch instance.Status {
			case string(workflowdomain.InstanceStatusRunning):
				return ErrRunningInstanceCannotDelete
			case string(workflowdomain.InstanceStatusCompleted),
				string(workflowdomain.InstanceStatusRejected),
				string(workflowdomain.InstanceStatusCancelled),
				"withdrawn":
			default:
				return ErrInstanceDeleteNotAllowed
			}
		}
		count, err := store.SoftDeleteInstances(ctx, instanceIDs, actorID, service.currentTime().UnixMilli())
		if err != nil {
			return err
		}
		if count != int64(len(instanceIDs)) {
			return ErrInstanceDeleteTargetNotFound
		}
		deleted = int(count)
		return nil
	})
	return deleted, err
}

func uniqueTrimmedStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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

const (
	maxDefinitionNameSearchLength = 50
	maxStarterNameSearchLength    = 50
)

func normalizeDefinitionNameSearch(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxDefinitionNameSearchLength {
		return "", ErrDefinitionNameSearchTooLong
	}
	return value, nil
}

func normalizeStarterNameSearch(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxStarterNameSearchLength {
		return "", ErrStarterNameSearchTooLong
	}
	return value, nil
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
