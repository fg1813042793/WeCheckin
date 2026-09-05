package workflowmodel

import "time"

const (
	InstanceStatusRunning   = "running"
	InstanceStatusCompleted = "completed"
	InstanceStatusRejected  = "rejected"
	InstanceStatusCancelled = "cancelled"

	TokenStatusActive    = "active"
	TokenStatusWaiting   = "waiting"
	TokenStatusCompleted = "completed"
	TokenStatusCancelled = "cancelled"

	TaskStatusWaiting   = "waiting"
	TaskStatusPending   = "pending"
	TaskStatusCompleted = "completed"
	TaskStatusApproved  = "approved"
	TaskStatusRejected  = "rejected"
	TaskStatusReturned  = "returned"
	TaskStatusCancelled = "cancelled"

	ParticipantRoleCC = "cc"

	NotificationKindNodeCC                 = "node_cc"
	NotificationKindNodeNotify             = "node_notify"
	NotificationKindTaskArrived            = "task_arrived"
	NotificationKindTaskReminder           = "task_reminder"
	NotificationKindInstanceFormRevised    = "instance_form_revised"
	NotificationKindApprovalResultApproved = "approval_result_approved"
	NotificationKindApprovalResultRejected = "approval_result_rejected"
	NotificationKindApprovalResultReturned = "approval_result_returned"

	NotificationChannelInApp      = "in_app"
	NotificationChannelDingTalkOA = "dingtalk_oa"

	NotificationStatusPending = "pending"
	NotificationStatusSending = "sending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
	NotificationStatusDead    = "dead"
)

// ProcessInstance is the persistence root of the generic workflow runtime.
// Business modules only associate through BusinessType and BusinessKey.
type ProcessInstance struct {
	ID                string    `json:"id" gorm:"size:64;primaryKey;comment:流程实例ID"`
	DefinitionID      uint      `json:"definitionId" gorm:"column:definition_id;uniqueIndex:idx_workflow_instance_business;index:idx_workflow_instances_definition_status,priority:1;index:idx_workflow_instances_definition_starter_time,priority:1;comment:流程定义ID"`
	DefinitionVersion int       `json:"definitionVersion" gorm:"column:definition_version;comment:流程定义版本"`
	DefinitionKey     string    `json:"definitionKey" gorm:"size:100;column:definition_key;comment:流程定义编码"`
	BusinessType      string    `json:"businessType" gorm:"size:100;column:business_type;uniqueIndex:idx_workflow_instance_business;comment:业务类型"`
	BusinessKey       string    `json:"businessKey" gorm:"size:160;column:business_key;uniqueIndex:idx_workflow_instance_business;comment:业务唯一标识"`
	StarterID         string    `json:"starterId" gorm:"size:64;column:starter_id;index:idx_workflow_instances_starter_status,priority:1;index:idx_workflow_instances_starter_deleted_time,priority:1;index:idx_workflow_instances_definition_starter_time,priority:2;comment:发起人ID"`
	OperatorID        string    `json:"operatorId" gorm:"size:64;column:operator_id;index:idx_workflow_instances_operator_status,priority:1;comment:实际发起操作人ID"`
	Status            string    `json:"status" gorm:"size:24;column:instance_status;default:running;index:idx_workflow_instances_definition_status,priority:2;index:idx_workflow_instances_starter_status,priority:2;index:idx_workflow_instances_operator_status,priority:2;comment:实例状态"`
	FormDataJSON      string    `json:"formDataJson" gorm:"type:mediumtext;column:form_data_json;comment:流程表单数据JSON"`
	FormRevision      int64     `json:"formRevision" gorm:"column:form_revision;default:1;comment:流程表单修订版本"`
	StartTime         int64     `json:"startTime" gorm:"column:start_time;index:idx_workflow_instances_starter_deleted_time,priority:3;index:idx_workflow_instances_definition_starter_time,priority:3;comment:开始时间"`
	EndTime           int64     `json:"endTime" gorm:"column:end_time;comment:结束时间"`
	StarterDeletedAt  int64     `json:"-" gorm:"column:starter_deleted_at;index:idx_workflow_instances_starter_deleted_time,priority:2;comment:发起人从我的申请删除时间"`
	AdminDeletedAt    int64     `json:"-" gorm:"column:admin_deleted_at;index:idx_workflow_instances_admin_deleted_time,priority:1;comment:管理员删除时间"`
	AdminDeletedBy    string    `json:"-" gorm:"size:64;column:admin_deleted_by;comment:删除操作管理员ID"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (ProcessInstance) TableName() string { return "workflow_process_instances" }

type StartDraft struct {
	ID                uint      `json:"id" gorm:"primaryKey;comment:流程发起草稿ID"`
	DefinitionID      uint      `json:"definitionId" gorm:"column:definition_id;uniqueIndex:uk_workflow_start_draft_owner,priority:1;comment:流程定义ID"`
	DefinitionVersion int       `json:"definitionVersion" gorm:"column:definition_version;comment:流程定义版本"`
	StarterID         string    `json:"starterId" gorm:"size:64;column:starter_id;uniqueIndex:uk_workflow_start_draft_owner,priority:2;index:idx_workflow_start_drafts_starter_time,priority:1;comment:草稿所属发起人ID"`
	FormDataJSON      string    `json:"formDataJson" gorm:"type:mediumtext;column:form_data_json;comment:草稿表单数据JSON"`
	EditTime          int64     `json:"editTime" gorm:"column:edit_time;index:idx_workflow_start_drafts_starter_time,priority:2;comment:草稿更新时间"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (StartDraft) TableName() string { return "workflow_start_drafts" }

type StartQuotaUsage struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:流程发起额度记录ID"`
	DefinitionID uint      `json:"definitionId" gorm:"column:definition_id;uniqueIndex:uk_workflow_start_quota_period,priority:1;comment:流程定义ID"`
	StarterID    string    `json:"starterId" gorm:"size:64;column:starter_id;uniqueIndex:uk_workflow_start_quota_period,priority:2;index:idx_workflow_start_quota_starter,priority:1;comment:业务发起人ID"`
	PeriodKey    string    `json:"periodKey" gorm:"size:100;column:period_key;uniqueIndex:uk_workflow_start_quota_period,priority:3;comment:额度周期唯一标识"`
	UsedCount    int       `json:"usedCount" gorm:"column:used_count;default:0;comment:已使用次数快照"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (StartQuotaUsage) TableName() string { return "workflow_start_quota_usage" }

type ProcessToken struct {
	ID          string    `json:"id" gorm:"size:64;primaryKey;comment:流程令牌ID"`
	InstanceID  string    `json:"instanceId" gorm:"size:64;column:instance_id;index:idx_workflow_tokens_instance_status,priority:1;comment:流程实例ID"`
	NodeID      string    `json:"nodeId" gorm:"size:100;column:node_id;index:idx_workflow_tokens_branch_node_status,priority:2;comment:当前节点ID"`
	Status      string    `json:"status" gorm:"size:24;column:token_status;default:active;index:idx_workflow_tokens_instance_status,priority:2;index:idx_workflow_tokens_branch_node_status,priority:3;comment:令牌状态"`
	BranchGroup string    `json:"branchGroup" gorm:"size:64;column:branch_group;index:idx_workflow_tokens_branch_node_status,priority:1;comment:并行分支组"`
	BranchTotal int       `json:"branchTotal" gorm:"column:branch_total;comment:并行分支总数"`
	ArrivedAt   int64     `json:"arrivedAt" gorm:"column:arrived_at;comment:到达时间"`
	CompletedAt int64     `json:"completedAt" gorm:"column:completed_at;comment:完成时间"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (ProcessToken) TableName() string { return "workflow_process_tokens" }

type ProcessTask struct {
	ID                   string    `json:"id" gorm:"size:64;primaryKey;comment:流程任务ID"`
	InstanceID           string    `json:"instanceId" gorm:"size:64;column:instance_id;index:idx_workflow_tasks_instance_status,priority:1;comment:流程实例ID"`
	TokenID              string    `json:"tokenId" gorm:"size:64;column:token_id;comment:流程令牌ID"`
	NodeID               string    `json:"nodeId" gorm:"size:100;column:node_id;comment:节点ID"`
	NodeName             string    `json:"nodeName" gorm:"size:200;column:node_name;comment:节点名称"`
	GroupKey             string    `json:"groupKey" gorm:"size:64;column:task_group_key;index:idx_workflow_tasks_group_status,priority:1;comment:多人审批任务组"`
	AssigneeID           string    `json:"assigneeId" gorm:"size:64;column:task_assignee_id;index:idx_workflow_tasks_assignee_status,priority:1;comment:处理人ID"`
	ApprovalMode         string    `json:"approvalMode" gorm:"size:24;column:approval_mode;default:single;comment:审批模式"`
	CompletionRate       int       `json:"completionRate" gorm:"column:completion_rate;default:100;comment:会签通过比例百分数"`
	Sequence             int       `json:"sequence" gorm:"column:task_sequence;default:1;comment:顺序审批序号"`
	Total                int       `json:"total" gorm:"column:task_total;default:1;comment:任务组任务数"`
	ApprovalChainKey     string    `json:"approvalChainKey,omitempty" gorm:"size:64;column:approval_chain_key;index:idx_workflow_tasks_chain_layer,priority:1;comment:分层审批链快照标识"`
	ApprovalLayer        int       `json:"approvalLayer,omitempty" gorm:"column:approval_layer;default:0;index:idx_workflow_tasks_chain_layer,priority:2;comment:分层审批层级序号"`
	ApprovalLayerTotal   int       `json:"approvalLayerTotal,omitempty" gorm:"column:approval_layer_total;default:0;comment:分层审批总层数"`
	SourceDepartmentID   uint      `json:"sourceDepartmentId,omitempty" gorm:"column:source_department_id;default:0;comment:审批层来源部门ID快照"`
	SourceDepartmentName string    `json:"sourceDepartmentName,omitempty" gorm:"size:100;column:source_department_name;comment:审批层来源部门名称快照"`
	Status               string    `json:"status" gorm:"size:24;column:task_status;default:pending;index:idx_workflow_tasks_assignee_status,priority:2;index:idx_workflow_tasks_instance_status,priority:2;index:idx_workflow_tasks_group_status,priority:2;comment:任务状态"`
	Action               string    `json:"action" gorm:"size:24;column:task_action;comment:处理动作"`
	Comment              string    `json:"comment" gorm:"size:1000;column:task_comment;comment:处理意见"`
	ImagesJSON           string    `json:"-" gorm:"type:mediumtext;column:task_images_json;comment:处理图片JSON"`
	HandledBy            string    `json:"handledBy" gorm:"size:64;column:handled_by;comment:实际处理人ID"`
	HandledAt            int64     `json:"handledAt" gorm:"column:handled_at;comment:处理时间"`
	AdminDeletedAt       int64     `json:"-" gorm:"column:admin_deleted_at;index:idx_workflow_tasks_admin_deleted_time,priority:1;comment:管理员删除时间"`
	AdminDeletedBy       string    `json:"-" gorm:"size:64;column:admin_deleted_by;comment:删除操作管理员ID"`
	CreatedAt            time.Time `json:"-"`
	UpdatedAt            time.Time `json:"-"`
}

func (ProcessTask) TableName() string { return "workflow_process_tasks" }

type ProcessVariable struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:流程变量ID"`
	InstanceID string    `json:"instanceId" gorm:"size:64;column:instance_id;uniqueIndex:idx_workflow_variable_instance_key;comment:流程实例ID"`
	Key        string    `json:"key" gorm:"size:100;column:variable_key;uniqueIndex:idx_workflow_variable_instance_key;comment:变量名"`
	ValueJSON  string    `json:"valueJson" gorm:"type:mediumtext;column:variable_value_json;comment:变量JSON值"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (ProcessVariable) TableName() string { return "workflow_process_variables" }

type ProcessHistory struct {
	ID         string    `json:"id" gorm:"size:64;primaryKey;comment:流程历史ID"`
	InstanceID string    `json:"instanceId" gorm:"size:64;column:instance_id;index:idx_workflow_history_instance_time,priority:1;comment:流程实例ID"`
	EventType  string    `json:"eventType" gorm:"size:40;column:event_type;comment:事件类型"`
	NodeID     string    `json:"nodeId" gorm:"size:100;column:node_id;comment:节点ID"`
	TaskID     string    `json:"taskId" gorm:"size:64;column:task_id;index:idx_workflow_history_task;comment:任务ID"`
	ActorID    string    `json:"actorId" gorm:"size:64;column:actor_id;comment:操作人ID"`
	Message    string    `json:"message" gorm:"size:1000;column:event_message;comment:事件说明"`
	ImagesJSON string    `json:"-" gorm:"type:mediumtext;column:event_images_json;comment:事件图片JSON"`
	EventTime  int64     `json:"eventTime" gorm:"column:event_time;index:idx_workflow_history_instance_time,priority:2;comment:事件时间"`
	CreatedAt  time.Time `json:"-"`
}

func (ProcessHistory) TableName() string { return "workflow_process_history" }

type InstanceParticipant struct {
	ID         string    `json:"id" gorm:"size:64;primaryKey;comment:参与人记录ID"`
	InstanceID string    `json:"instanceId" gorm:"size:64;column:instance_id;uniqueIndex:uk_workflow_participant_source,priority:1;index:idx_workflow_participant_instance;index:idx_workflow_participant_user_role,priority:3;comment:流程实例ID"`
	UserID     string    `json:"userId" gorm:"size:64;column:user_id;uniqueIndex:uk_workflow_participant_source,priority:2;index:idx_workflow_participant_user_role,priority:1;comment:本地用户ID"`
	Role       string    `json:"role" gorm:"size:24;column:participant_role;uniqueIndex:uk_workflow_participant_source,priority:3;index:idx_workflow_participant_user_role,priority:2;comment:参与角色"`
	NodeID     string    `json:"nodeId" gorm:"size:100;column:node_id;uniqueIndex:uk_workflow_participant_source,priority:4;comment:来源节点ID"`
	AddTime    int64     `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (InstanceParticipant) TableName() string { return "workflow_instance_participants" }

type NotificationOutbox struct {
	ID                string    `json:"id" gorm:"size:64;primaryKey;comment:通知Outbox ID"`
	InstanceID        string    `json:"instanceId" gorm:"size:64;column:instance_id;index:idx_workflow_notification_instance,priority:1;comment:流程实例ID"`
	NodeID            string    `json:"nodeId" gorm:"size:100;column:node_id;comment:来源节点ID"`
	TaskID            string    `json:"taskId" gorm:"size:64;column:task_id;comment:来源任务ID"`
	RecipientUserID   string    `json:"recipientUserId" gorm:"size:64;column:recipient_user_id;index:idx_workflow_notification_recipient,priority:1;comment:本地接收人ID"`
	Kind              string    `json:"kind" gorm:"size:32;column:notification_kind;comment:通知类型"`
	Channel           string    `json:"channel" gorm:"size:32;column:notification_channel;comment:通知渠道"`
	Status            string    `json:"status" gorm:"size:24;column:notification_status;index:idx_workflow_notification_due,priority:1;index:idx_workflow_notification_recipient,priority:2;comment:投递状态"`
	DedupeKey         string    `json:"dedupeKey" gorm:"size:255;column:dedupe_key;uniqueIndex:uk_workflow_notification_dedupe;comment:通知幂等键"`
	PayloadJSON       string    `json:"payloadJson" gorm:"type:mediumtext;column:payload_json;comment:通知负载JSON"`
	CorpID            string    `json:"corpId" gorm:"size:120;column:corp_id;comment:钉钉企业ID"`
	ProviderMessageID string    `json:"providerMessageId" gorm:"size:160;column:provider_message_id;comment:渠道消息标识"`
	Attempts          int       `json:"attempts" gorm:"column:attempts;comment:投递尝试次数"`
	NextRetryAt       int64     `json:"nextRetryAt" gorm:"column:next_retry_at;index:idx_workflow_notification_due,priority:2;comment:下次重试时间"`
	LastError         string    `json:"lastError" gorm:"size:1000;column:last_error;comment:最近失败摘要"`
	SentAt            int64     `json:"sentAt" gorm:"column:sent_at;comment:发送成功时间"`
	AddTime           int64     `json:"addTime" gorm:"column:add_time;index:idx_workflow_notification_instance,priority:2;comment:创建时间"`
	EditTime          int64     `json:"editTime" gorm:"column:edit_time;comment:更新时间"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (NotificationOutbox) TableName() string { return "workflow_notification_outbox" }
