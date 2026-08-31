package domain

import workflowcore "wecheckin/backend/internal/workflow"

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
	TaskStatusApproved  TaskStatus = "approved"
	TaskStatusRejected  TaskStatus = "rejected"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type TaskAction string

const (
	TaskActionApprove TaskAction = "approve"
	TaskActionReject  TaskAction = "reject"
)

type HistoryEventType string

const (
	HistoryInstanceStarted   HistoryEventType = "instance_started"
	HistoryTaskCreated       HistoryEventType = "task_created"
	HistoryTaskActivated     HistoryEventType = "task_activated"
	HistoryTaskApproved      HistoryEventType = "task_approved"
	HistoryTaskRejected      HistoryEventType = "task_rejected"
	HistoryTaskCancelled     HistoryEventType = "task_cancelled"
	HistoryInstanceCompleted HistoryEventType = "instance_completed"
	HistoryInstanceRejected  HistoryEventType = "instance_rejected"
	HistoryInstanceWithdrawn HistoryEventType = "instance_withdrawn"
	HistoryInstanceCancelled HistoryEventType = "instance_cancelled"
)

type ProcessInstance struct {
	ID                string
	DefinitionID      uint
	DefinitionVersion int
	DefinitionKey     string
	BusinessType      string
	BusinessKey       string
	StarterID         string
	Status            InstanceStatus
}

type Token struct {
	ID          string
	NodeID      string
	Status      TokenStatus
	BranchGroup string
	BranchTotal int
}

type Task struct {
	ID             string
	TokenID        string
	NodeID         string
	NodeName       string
	GroupKey       string
	AssigneeID     string
	ApprovalMode   string
	CompletionRate int
	Sequence       int
	Total          int
	Status         TaskStatus
	Action         TaskAction
	Comment        string
}

type HistoryEvent struct {
	ID      string
	Type    HistoryEventType
	NodeID  string
	TaskID  string
	ActorID string
	Message string
}

type State struct {
	Instance  ProcessInstance
	Tokens    []Token
	Tasks     []Task
	Variables map[string]interface{}
	FormData  map[string]interface{}
	History   []HistoryEvent
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
	Variables         map[string]interface{}
	FormData          map[string]interface{}
}

type CompleteRequest struct {
	TaskID    string
	ActorID   string
	Action    TaskAction
	Comment   string
	Variables map[string]interface{}
	FormData  map[string]interface{}
}

type AssigneeRequest struct {
	Instance  ProcessInstance
	Node      workflowcore.Node
	Variables map[string]interface{}
}

type AssigneeResolver interface {
	Resolve(AssigneeRequest) ([]string, error)
}

type IDGenerator interface {
	NewID(prefix string) string
}
