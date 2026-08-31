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
	TaskStatusApproved  = "approved"
	TaskStatusRejected  = "rejected"
	TaskStatusCancelled = "cancelled"
)

// ProcessInstance is the persistence root of the generic workflow runtime.
// Business modules only associate through BusinessType and BusinessKey.
type ProcessInstance struct {
	ID                string    `json:"id" gorm:"size:64;primaryKey;comment:流程实例ID"`
	DefinitionID      uint      `json:"definitionId" gorm:"column:definition_id;uniqueIndex:idx_workflow_instance_business;index:idx_workflow_instances_definition_status,priority:1;comment:流程定义ID"`
	DefinitionVersion int       `json:"definitionVersion" gorm:"column:definition_version;comment:流程定义版本"`
	DefinitionKey     string    `json:"definitionKey" gorm:"size:100;column:definition_key;comment:流程定义编码"`
	BusinessType      string    `json:"businessType" gorm:"size:100;column:business_type;uniqueIndex:idx_workflow_instance_business;comment:业务类型"`
	BusinessKey       string    `json:"businessKey" gorm:"size:160;column:business_key;uniqueIndex:idx_workflow_instance_business;comment:业务唯一标识"`
	StarterID         string    `json:"starterId" gorm:"size:64;column:starter_id;index:idx_workflow_instances_starter_status,priority:1;comment:发起人ID"`
	Status            string    `json:"status" gorm:"size:24;column:instance_status;default:running;index:idx_workflow_instances_definition_status,priority:2;index:idx_workflow_instances_starter_status,priority:2;comment:实例状态"`
	FormDataJSON      string    `json:"formDataJson" gorm:"type:mediumtext;column:form_data_json;comment:流程表单数据JSON"`
	StartTime         int64     `json:"startTime" gorm:"column:start_time;comment:开始时间"`
	EndTime           int64     `json:"endTime" gorm:"column:end_time;comment:结束时间"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (ProcessInstance) TableName() string { return "workflow_process_instances" }

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
	ID             string    `json:"id" gorm:"size:64;primaryKey;comment:流程任务ID"`
	InstanceID     string    `json:"instanceId" gorm:"size:64;column:instance_id;index:idx_workflow_tasks_instance_status,priority:1;comment:流程实例ID"`
	TokenID        string    `json:"tokenId" gorm:"size:64;column:token_id;comment:流程令牌ID"`
	NodeID         string    `json:"nodeId" gorm:"size:100;column:node_id;comment:节点ID"`
	NodeName       string    `json:"nodeName" gorm:"size:200;column:node_name;comment:节点名称"`
	GroupKey       string    `json:"groupKey" gorm:"size:64;column:task_group_key;index:idx_workflow_tasks_group_status,priority:1;comment:多人审批任务组"`
	AssigneeID     string    `json:"assigneeId" gorm:"size:64;column:task_assignee_id;index:idx_workflow_tasks_assignee_status,priority:1;comment:处理人ID"`
	ApprovalMode   string    `json:"approvalMode" gorm:"size:24;column:approval_mode;default:single;comment:审批模式"`
	CompletionRate int       `json:"completionRate" gorm:"column:completion_rate;default:100;comment:会签通过比例百分数"`
	Sequence       int       `json:"sequence" gorm:"column:task_sequence;default:1;comment:顺序审批序号"`
	Total          int       `json:"total" gorm:"column:task_total;default:1;comment:任务组任务数"`
	Status         string    `json:"status" gorm:"size:24;column:task_status;default:pending;index:idx_workflow_tasks_assignee_status,priority:2;index:idx_workflow_tasks_instance_status,priority:2;index:idx_workflow_tasks_group_status,priority:2;comment:任务状态"`
	Action         string    `json:"action" gorm:"size:24;column:task_action;comment:处理动作"`
	Comment        string    `json:"comment" gorm:"size:1000;column:task_comment;comment:处理意见"`
	HandledBy      string    `json:"handledBy" gorm:"size:64;column:handled_by;comment:实际处理人ID"`
	HandledAt      int64     `json:"handledAt" gorm:"column:handled_at;comment:处理时间"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
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
	EventTime  int64     `json:"eventTime" gorm:"column:event_time;index:idx_workflow_history_instance_time,priority:2;comment:事件时间"`
	CreatedAt  time.Time `json:"-"`
}

func (ProcessHistory) TableName() string { return "workflow_process_history" }
