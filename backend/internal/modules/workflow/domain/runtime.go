package domain

import (
	"context"

	"wecheckin/backend/internal/workflowcore"
)

type InstanceStatus string

const (
	InstanceStatusRunning   InstanceStatus = "running"
	InstanceStatusCompleted InstanceStatus = "completed"
	InstanceStatusRejected  InstanceStatus = "rejected"
	InstanceStatusCancelled InstanceStatus = "cancelled"
)

type TokenStatus string

const (
	TokenStatusActive    TokenStatus = "active"
	TokenStatusWaiting   TokenStatus = "waiting"
	TokenStatusCompleted TokenStatus = "completed"
	TokenStatusCancelled TokenStatus = "cancelled"
)

type TaskStatus string

const (
	TaskStatusWaiting   TaskStatus = "waiting"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusApproved  TaskStatus = "approved"
	TaskStatusRejected  TaskStatus = "rejected"
	TaskStatusReturned  TaskStatus = "returned"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type TaskAction string

const (
	TaskActionApprove TaskAction = "approve"
	TaskActionReject  TaskAction = "reject"
	TaskActionReturn  TaskAction = "return"
	TaskActionSubmit  TaskAction = "submit"
)

type HistoryEventType string

const (
	HistoryInstanceStarted     HistoryEventType = "instance_started"
	HistoryTaskCreated         HistoryEventType = "task_created"
	HistoryTaskActivated       HistoryEventType = "task_activated"
	HistoryTaskApproved        HistoryEventType = "task_approved"
	HistoryTaskRejected        HistoryEventType = "task_rejected"
	HistoryTaskReturned        HistoryEventType = "task_returned"
	HistoryTaskSubmitted       HistoryEventType = "task_submitted"
	HistoryTaskCancelled       HistoryEventType = "task_cancelled"
	HistoryNodeCC              HistoryEventType = "node_cc"
	HistoryNodeNotify          HistoryEventType = "node_notify"
	HistoryNodeAutomated       HistoryEventType = "node_automated"
	HistoryTimerWaiting        HistoryEventType = "timer_waiting"
	HistoryTimerResumed        HistoryEventType = "timer_resumed"
	HistoryInstanceCompleted   HistoryEventType = "instance_completed"
	HistoryInstanceRejected    HistoryEventType = "instance_rejected"
	HistoryInstanceWithdrawn   HistoryEventType = "instance_withdrawn"
	HistoryInstanceCancelled   HistoryEventType = "instance_cancelled"
	HistoryInstanceCommented   HistoryEventType = "instance_commented"
	HistoryInstanceReminded    HistoryEventType = "instance_reminded"
	HistoryInstanceFormRevised HistoryEventType = "instance_form_revised"
)

type ParticipantRole string

const ParticipantRoleCC ParticipantRole = "cc"

type NotificationKind string

const (
	NotificationKindNodeCC                 NotificationKind = "node_cc"
	NotificationKindNodeNotify             NotificationKind = "node_notify"
	NotificationKindTaskArrived            NotificationKind = "task_arrived"
	NotificationKindTaskReminder           NotificationKind = "task_reminder"
	NotificationKindInstanceCommented      NotificationKind = "instance_commented"
	NotificationKindInstanceFormRevised    NotificationKind = "instance_form_revised"
	NotificationKindApprovalResultApproved NotificationKind = "approval_result_approved"
	NotificationKindApprovalResultRejected NotificationKind = "approval_result_rejected"
	NotificationKindApprovalResultReturned NotificationKind = "approval_result_returned"
)

type ProcessInstance struct {
	ID                string
	DefinitionID      uint
	DefinitionVersion int
	DefinitionKey     string
	BusinessType      string
	BusinessKey       string
	StarterID         string
	OperatorID        string
	Status            InstanceStatus
	StartTime         int64
	FormRevision      int64
}

type Token struct {
	ID          string
	NodeID      string
	Status      TokenStatus
	BranchGroup string
	BranchTotal int
}

type Task struct {
	ID                 string
	TokenID            string
	NodeID             string
	NodeName           string
	GroupKey           string
	AssigneeID         string
	ApprovalMode       string
	CompletionRate     int
	Sequence           int
	Total              int
	ApprovalChainKey   string
	ApprovalLayer      int
	ApprovalLayerTotal int
	DepartmentID       uint
	DepartmentName     string
	Status             TaskStatus
	Action             TaskAction
	Comment            string
	Images             []workflowcore.FormAttachment
}

type HistoryEvent struct {
	ID        string
	Type      HistoryEventType
	NodeID    string
	TaskID    string
	ActorID   string
	Message   string
	Images    []workflowcore.FormAttachment
	EventTime int64
}

type Participant struct {
	ID     string
	UserID string
	Role   ParticipantRole
	NodeID string
}

type NotificationIntent struct {
	ID              string
	Kind            NotificationKind
	NodeID          string
	NodeName        string
	TargetNodeName  string
	TaskID          string
	RecipientUserID string
	WorkflowName    string
	Config          workflowcore.NotificationConfig
	DedupeKeySuffix string
}

type State struct {
	Instance            ProcessInstance
	Tokens              []Token
	Tasks               []Task
	Variables           map[string]interface{}
	FormData            map[string]interface{}
	History             []HistoryEvent
	Participants        []Participant
	NotificationIntents []NotificationIntent
}

func (state *State) PendingTasks() []Task {
	result := make([]Task, 0)
	for _, task := range state.Tasks {
		if task.Status == TaskStatusPending {
			result = append(result, task)
		}
	}
	return result
}

type StartRequest struct {
	DefinitionID      uint
	DefinitionVersion int
	BusinessType      string
	BusinessKey       string
	StarterID         string
	OperatorID        string
	StartTime         int64
	Variables         map[string]interface{}
	FormData          map[string]interface{}
}

type CompleteRequest struct {
	TaskID             string
	ActorID            string
	Action             TaskAction
	Comment            string
	Images             []workflowcore.FormAttachment
	ReturnTargetNodeID string
	Variables          map[string]interface{}
	FormData           map[string]interface{}
}

type AssigneeRequest struct {
	Instance  ProcessInstance
	Node      workflowcore.Node
	Variables map[string]interface{}
}

type ApprovalLayer struct {
	DepartmentID   uint
	DepartmentName string
	AssigneeIDs    []string
}

type AssigneeResolver interface {
	Resolve(context.Context, AssigneeRequest) ([]string, error)
}

type ApprovalLayerResolver interface {
	ResolveApprovalLayers(context.Context, AssigneeRequest) ([]ApprovalLayer, error)
}

type IDGenerator interface {
	NewID(prefix string) string
}
