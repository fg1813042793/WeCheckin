package workflowmodel

import "time"

const (
	FormRevisionStatusPending   = "pending"
	FormRevisionStatusApplied   = "applied"
	FormRevisionStatusRejected  = "rejected"
	FormRevisionStatusCancelled = "cancelled"
	FormRevisionStatusConflict  = "conflict"

	FormRevisionTaskStatusWaiting   = "waiting"
	FormRevisionTaskStatusPending   = "pending"
	FormRevisionTaskStatusApproved  = "approved"
	FormRevisionTaskStatusRejected  = "rejected"
	FormRevisionTaskStatusCancelled = "cancelled"

	BusinessEventStatusPending = "pending"
	BusinessEventStatusSending = "sending"
	BusinessEventStatusSent    = "sent"
	BusinessEventStatusFailed  = "failed"
	BusinessEventStatusDead    = "dead"
)

type FormRevisionRequest struct {
	ID                     string    `json:"id" gorm:"size:64;primaryKey;comment:表单修订请求ID"`
	SourceInstanceID       string    `json:"sourceInstanceId" gorm:"size:64;column:source_instance_id;index:idx_workflow_form_revision_instance_time,priority:1;comment:原流程实例ID"`
	SourceNodeID           string    `json:"sourceNodeId" gorm:"size:100;column:source_node_id;comment:修订来源节点ID快照"`
	SourceNodeName         string    `json:"sourceNodeName" gorm:"size:200;column:source_node_name;comment:修订来源节点名称快照"`
	RequesterID            string    `json:"requesterId" gorm:"size:64;column:requester_id;index:idx_workflow_form_revision_requester_time,priority:1;comment:修订发起人ID"`
	BaseFormRevision       int64     `json:"baseFormRevision" gorm:"column:base_form_revision;comment:修订基准表单版本"`
	BeforeFormDataJSON     string    `json:"-" gorm:"type:mediumtext;column:before_form_data_json;comment:修改前表单JSON快照"`
	PatchJSON              string    `json:"-" gorm:"type:mediumtext;column:patch_json;comment:修订补丁JSON"`
	ProposedFormDataJSON   string    `json:"-" gorm:"type:mediumtext;column:proposed_form_data_json;comment:候选表单JSON快照"`
	ChangedFieldLabelsJSON string    `json:"-" gorm:"type:mediumtext;column:changed_field_labels_json;comment:修改字段名称JSON快照"`
	Reason                 string    `json:"reason" gorm:"size:500;column:reason;comment:修订原因"`
	RevisionMode           string    `json:"revisionMode" gorm:"size:32;column:revision_mode;comment:修订模式"`
	Status                 string    `json:"status" gorm:"size:24;column:revision_status;comment:修订状态"`
	ActiveLockKey          *string   `json:"-" gorm:"size:64;column:active_lock_key;uniqueIndex:uk_workflow_form_revision_active;comment:进行中修订唯一锁"`
	AppliedFormRevision    int64     `json:"appliedFormRevision" gorm:"column:applied_form_revision;default:0;comment:生效后的表单版本"`
	AddTime                int64     `json:"addTime" gorm:"column:add_time;index:idx_workflow_form_revision_instance_time,priority:2;index:idx_workflow_form_revision_requester_time,priority:2;comment:创建时间"`
	EditTime               int64     `json:"editTime" gorm:"column:edit_time;comment:更新时间"`
	CompletedAt            int64     `json:"completedAt" gorm:"column:completed_at;default:0;comment:终态时间"`
	CreatedAt              time.Time `json:"-"`
	UpdatedAt              time.Time `json:"-"`
}

func (FormRevisionRequest) TableName() string { return "workflow_form_revision_requests" }

type FormRevisionTask struct {
	ID                string    `json:"id" gorm:"size:64;primaryKey;comment:表单修订任务ID"`
	RevisionRequestID string    `json:"revisionRequestId" gorm:"size:64;column:revision_request_id;index:idx_workflow_form_revision_task_stage,priority:1;comment:修订请求ID"`
	SourceTaskID      string    `json:"sourceTaskId" gorm:"size:64;column:source_task_id;comment:原流程任务ID快照"`
	NodeID            string    `json:"nodeId" gorm:"size:100;column:node_id;comment:确认节点ID快照"`
	NodeName          string    `json:"nodeName" gorm:"size:200;column:node_name;comment:确认节点名称快照"`
	Stage             int       `json:"stage" gorm:"column:stage;index:idx_workflow_form_revision_task_stage,priority:2;comment:确认阶段"`
	AssigneeID        string    `json:"assigneeId" gorm:"size:64;column:assignee_id;index:idx_workflow_form_revision_task_assignee,priority:1;comment:任务处理人ID"`
	AssigneeName      string    `json:"assigneeName" gorm:"size:200;column:assignee_name;comment:任务处理人名称快照"`
	ApprovalMode      string    `json:"approvalMode" gorm:"size:24;column:approval_mode;default:single;comment:审批模式"`
	CompletionRate    int       `json:"completionRate" gorm:"column:completion_rate;default:100;comment:会签通过比例"`
	Sequence          int       `json:"sequence" gorm:"column:task_sequence;default:1;comment:组内顺序"`
	Total             int       `json:"total" gorm:"column:task_total;default:1;comment:组内任务总数"`
	Status            string    `json:"status" gorm:"size:24;column:task_status;index:idx_workflow_form_revision_task_assignee,priority:2;index:idx_workflow_form_revision_task_stage,priority:3;comment:任务状态"`
	Action            string    `json:"action" gorm:"size:24;column:task_action;comment:处理动作"`
	Comment           string    `json:"comment" gorm:"size:1000;column:task_comment;comment:处理意见"`
	ImagesJSON        string    `json:"-" gorm:"type:mediumtext;column:task_images_json;comment:处理图片JSON"`
	HandledBy         string    `json:"handledBy" gorm:"size:64;column:handled_by;comment:实际处理人ID"`
	HandledAt         int64     `json:"handledAt" gorm:"column:handled_at;comment:处理时间"`
	AddTime           int64     `json:"addTime" gorm:"column:add_time;index:idx_workflow_form_revision_task_assignee,priority:3;comment:创建时间"`
	EditTime          int64     `json:"editTime" gorm:"column:edit_time;comment:更新时间"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (FormRevisionTask) TableName() string { return "workflow_form_revision_tasks" }

type BusinessEventOutbox struct {
	ID           string    `json:"id" gorm:"size:64;primaryKey;comment:业务事件Outbox ID"`
	EventType    string    `json:"eventType" gorm:"size:100;column:event_type;comment:业务事件类型"`
	AggregateID  string    `json:"aggregateId" gorm:"size:64;column:aggregate_id;index:idx_workflow_business_event_aggregate;comment:聚合ID"`
	BusinessType string    `json:"businessType" gorm:"size:100;column:business_type;index:idx_workflow_business_event_reference,priority:1;comment:业务类型"`
	BusinessKey  string    `json:"businessKey" gorm:"size:160;column:business_key;index:idx_workflow_business_event_reference,priority:2;comment:业务标识"`
	PayloadJSON  string    `json:"-" gorm:"type:mediumtext;column:payload_json;comment:事件负载JSON"`
	Status       string    `json:"status" gorm:"size:24;column:event_status;index:idx_workflow_business_event_due,priority:1;comment:派发状态"`
	DedupeKey    string    `json:"dedupeKey" gorm:"size:191;column:dedupe_key;uniqueIndex:uk_workflow_business_event_dedupe;comment:事件幂等键"`
	Attempts     int       `json:"attempts" gorm:"column:attempts;default:0;comment:派发尝试次数"`
	NextRetryAt  int64     `json:"nextRetryAt" gorm:"column:next_retry_at;index:idx_workflow_business_event_due,priority:2;comment:下次重试时间"`
	LastError    string    `json:"lastError" gorm:"size:1000;column:last_error;comment:最近失败摘要"`
	SentAt       int64     `json:"sentAt" gorm:"column:sent_at;comment:派发成功时间"`
	AddTime      int64     `json:"addTime" gorm:"column:add_time;comment:创建时间"`
	EditTime     int64     `json:"editTime" gorm:"column:edit_time;comment:更新时间"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (BusinessEventOutbox) TableName() string { return "workflow_business_event_outbox" }
