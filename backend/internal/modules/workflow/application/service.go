package application

import (
	"context"
	"errors"
	"time"
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
	ErrInstanceFormRevisionNotAllowed        = errors.New("当前用户没有办理后修改表单的权限")
	ErrInstanceFormRevisionChanged           = errors.New("流程表单已被更新，请重新打开后修改")
	ErrInstanceFormRevisionRequired          = errors.New("流程表单版本不能为空")
	ErrInstanceFormRevisionDataRequired      = errors.New("请至少修改一个表单字段")
	ErrInstanceFormRevisionReasonRequired    = errors.New("请填写修改原因")
	ErrInstanceFormRevisionReasonTooLong     = errors.New("修改原因不能超过500个字符")
	ErrFormRevisionNotificationTooMany       = errors.New("表单修改通知对象不能超过100人")
	ErrFormRevisionNotificationChannels      = errors.New("请选择表单修改通知方式")
	ErrFormRevisionNotificationChannel       = errors.New("表单修改通知方式无效")
	ErrFormRevisionNotificationRecipient     = errors.New("表单修改通知对象不是该流程参与人")
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
	GetWorkflowOverview(ctx context.Context, actorID string) (*WorkflowOverview, error)
	GetInstance(ctx context.Context, instanceID string) (*InstanceDetail, error)
	HideStartedInstance(ctx context.Context, instanceID, starterID string, deletedAt int64) error
	ListTasks(ctx context.Context, query TaskQuery) (*TaskList, error)
	HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error)
}

type TransactionStore interface {
	UserDepartmentReader
	HasParticipant(ctx context.Context, instanceID, userID, role string) (bool, error)
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
