package application

import "wecheckin/backend/internal/workflowcore"

const (
	InstanceScopeStarted = "started"
	InstanceScopeHandled = "handled"
	InstanceScopeCopied  = "copied"
)

type PublishedDefinition struct {
	ID                 uint                                      `json:"id"`
	Key                string                                    `json:"key"`
	Name               string                                    `json:"name"`
	Description        string                                    `json:"description"`
	Category           string                                    `json:"category"`
	Version            int                                       `json:"version"`
	Form               []workflowcore.FormField                  `json:"form"`
	FieldPermissions   map[string][]workflowcore.FieldPermission `json:"fieldPermissions"`
	StartNodeID        string                                    `json:"startNodeId"`
	Initiator          workflowcore.InitiatorConfig              `json:"initiator"`
	Availability       workflowcore.StartAvailabilityConfig      `json:"availability"`
	AvailabilityStatus string                                    `json:"availabilityStatus"`
}

type StartDraft struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	StarterID         string                 `json:"-"`
	FormData          map[string]interface{} `json:"formData"`
	UpdatedAt         int64                  `json:"updatedAt"`
}

type InstanceQuery struct {
	DefinitionID uint
	Status       string
	BusinessType string
	BusinessKey  string
	StarterID    string
	Scope        string
	ScopeUserID  string
	Page         int
	PageSize     int
}

type NotificationPayload struct {
	Title           string `json:"title"`
	Content         string `json:"content"`
	WorkflowName    string `json:"workflowName"`
	NodeName        string `json:"nodeName"`
	StarterID       string `json:"starterId"`
	StarterName     string `json:"starterName"`
	InstanceID      string `json:"instanceId"`
	TaskID          string `json:"taskId"`
	RecipientUserID string `json:"recipientUserId"`
	Kind            string `json:"kind"`
}

type TaskQuery struct {
	InstanceID string
	AssigneeID string
	Status     string
	Page       int
	PageSize   int
}

type InstanceSummary struct {
	ID                string `json:"id"`
	DefinitionID      uint   `json:"definitionId"`
	DefinitionVersion int    `json:"definitionVersion"`
	DefinitionKey     string `json:"definitionKey"`
	BusinessType      string `json:"businessType"`
	BusinessKey       string `json:"businessKey"`
	StarterID         string `json:"starterId"`
	OperatorID        string `json:"operatorId"`
	Status            string `json:"status"`
	StartTime         int64  `json:"startTime"`
	EndTime           int64  `json:"endTime"`
}

type TokenSummary struct {
	ID          string `json:"id"`
	NodeID      string `json:"nodeId"`
	Status      string `json:"status"`
	BranchGroup string `json:"branchGroup"`
	BranchTotal int    `json:"branchTotal"`
}

type TaskSummary struct {
	ID             string `json:"id"`
	InstanceID     string `json:"instanceId"`
	NodeID         string `json:"nodeId"`
	NodeName       string `json:"nodeName"`
	AssigneeID     string `json:"assigneeId"`
	ApprovalMode   string `json:"approvalMode"`
	CompletionRate int    `json:"completionRate"`
	Sequence       int    `json:"sequence"`
	Total          int    `json:"total"`
	Status         string `json:"status"`
	Action         string `json:"action"`
	Comment        string `json:"comment"`
	HandledBy      string `json:"handledBy"`
	HandledAt      int64  `json:"handledAt"`
}

type HistorySummary struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	NodeID    string `json:"nodeId"`
	TaskID    string `json:"taskId"`
	ActorID   string `json:"actorId"`
	Message   string `json:"message"`
	EventTime int64  `json:"eventTime"`
}

type InstanceDetail struct {
	Instance         InstanceSummary                           `json:"instance"`
	Variables        map[string]interface{}                    `json:"variables"`
	Form             []workflowcore.FormField                  `json:"form"`
	FormData         map[string]interface{}                    `json:"formData"`
	FieldPermissions map[string][]workflowcore.FieldPermission `json:"fieldPermissions"`
	StartNodeID      string                                    `json:"startNodeId"`
	NodeTypes        map[string]string                         `json:"nodeTypes"`
	Tokens           []TokenSummary                            `json:"tokens"`
	Tasks            []TaskSummary                             `json:"tasks"`
	History          []HistorySummary                          `json:"history"`
}

type InstanceList struct {
	List     []InstanceSummary `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

type TaskList struct {
	List     []TaskSummary `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
