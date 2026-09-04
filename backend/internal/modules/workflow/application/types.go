package application

import "wecheckin/backend/internal/workflowcore"

const (
	InstanceScopeStarted = "started"
	InstanceScopeHandled = "handled"
	InstanceScopeCopied  = "copied"
)

const (
	NodeProgressCompleted  = "completed"
	NodeProgressProcessing = "processing"
	NodeProgressNotStarted = "not_started"
	NodeProgressSkipped    = "skipped"
	NodeProgressReturned   = "returned"
	NodeProgressTerminated = "terminated"
)

type PublishedDefinition struct {
	ID                 uint                                      `json:"id"`
	Key                string                                    `json:"key"`
	Name               string                                    `json:"name"`
	Description        string                                    `json:"description"`
	Category           string                                    `json:"category"`
	LogoURL            string                                    `json:"logoUrl"`
	Version            int                                       `json:"version"`
	Form               []workflowcore.FormField                  `json:"form"`
	FieldPermissions   map[string][]workflowcore.FieldPermission `json:"fieldPermissions"`
	StartNodeID        string                                    `json:"startNodeId"`
	Initiator          workflowcore.InitiatorConfig              `json:"initiator"`
	Availability       workflowcore.StartAvailabilityConfig      `json:"availability"`
	AvailabilityStatus string                                    `json:"availabilityStatus"`
	StartLimit         workflowcore.StartLimitConfig             `json:"startLimit"`
	StartLimitStatus   StartLimitStatus                          `json:"startLimitStatus"`
	Nodes              []PublishedNode                           `json:"nodes,omitempty"`
	Edges              []PublishedEdge                           `json:"edges,omitempty"`
}

type StartLimitStatus struct {
	Allowed        bool  `json:"allowed"`
	UsedCount      int   `json:"usedCount"`
	RemainingCount int   `json:"remainingCount"`
	ResetsAt       int64 `json:"resetsAt,omitempty"`
}

type PublishedNode struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Name            string                 `json:"name"`
	Position        *workflowcore.Position `json:"position,omitempty"`
	ApprovalMode    string                 `json:"approvalMode,omitempty"`
	GatewayMode     string                 `json:"gatewayMode,omitempty"`
	AssigneeDisplay string                 `json:"assigneeDisplay,omitempty"`
	Assignee        *workflowcore.Assignee `json:"-"`
}

type PublishedEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
	Name         string `json:"name,omitempty"`
	Default      bool   `json:"default,omitempty"`
}

type StartDraft struct {
	DefinitionID      uint                   `json:"definitionId"`
	DefinitionVersion int                    `json:"definitionVersion"`
	StarterID         string                 `json:"-"`
	FormData          map[string]interface{} `json:"formData"`
	UpdatedAt         int64                  `json:"updatedAt"`
}

type InstanceQuery struct {
	DefinitionID       uint
	DefinitionVersion  int
	DefinitionName     string
	DefinitionCategory string
	StarterName        string
	Status             string
	BusinessType       string
	BusinessKey        string
	StarterID          string
	Scope              string
	ScopeUserID        string
	StartTimeFrom      int64
	StartTimeTo        int64
	EndTimeFrom        int64
	EndTimeTo          int64
	InstanceIDs        []string
	Visibility         *InstanceVisibility
	Page               int
	PageSize           int
}

// InstanceVisibility limits management queries by the current department
// membership of each workflow starter. A nil visibility keeps legacy callers
// unchanged; a non-nil but not-ready visibility denies all rows.
type InstanceVisibility struct {
	Ready         bool
	All           bool
	UserIDs       []uint
	DepartmentIDs []uint
}

const NotificationMessageTypeActionCard = "action_card"

type NotificationPayload struct {
	Title           string `json:"title"`
	Content         string `json:"content"`
	MessageType     string `json:"messageType,omitempty"`
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
	InstanceID         string
	AssigneeID         string
	Status             string
	HideAdminDeleted   bool
	DefinitionName     string
	DefinitionCategory string
	StarterName        string
	StartTimeFrom      int64
	StartTimeTo        int64
	Page               int
	PageSize           int
}

type InstanceSummary struct {
	ID                   string   `json:"id"`
	DefinitionID         uint     `json:"definitionId"`
	DefinitionVersion    int      `json:"definitionVersion"`
	DefinitionKey        string   `json:"definitionKey"`
	DefinitionName       string   `json:"definitionName"`
	BusinessType         string   `json:"businessType"`
	BusinessKey          string   `json:"businessKey"`
	StarterID            string   `json:"starterId"`
	StarterName          string   `json:"starterName"`
	OperatorID           string   `json:"operatorId"`
	OperatorName         string   `json:"operatorName"`
	CurrentNodeNames     []string `json:"currentNodeNames"`
	CurrentAssigneeNames []string `json:"currentAssigneeNames"`
	Status               string   `json:"status"`
	StartTime            int64    `json:"startTime"`
	EndTime              int64    `json:"endTime"`
}

type TokenSummary struct {
	ID          string `json:"id"`
	NodeID      string `json:"nodeId"`
	Status      string `json:"status"`
	BranchGroup string `json:"branchGroup"`
	BranchTotal int    `json:"branchTotal"`
}

type TaskSummary struct {
	ID                   string                        `json:"id"`
	InstanceID           string                        `json:"instanceId"`
	NodeID               string                        `json:"nodeId"`
	NodeName             string                        `json:"nodeName"`
	DefinitionName       string                        `json:"definitionName"`
	StarterID            string                        `json:"starterId"`
	StarterName          string                        `json:"starterName"`
	AssigneeID           string                        `json:"assigneeId"`
	AssigneeName         string                        `json:"assigneeName"`
	ApprovalMode         string                        `json:"approvalMode"`
	CompletionRate       int                           `json:"completionRate"`
	Sequence             int                           `json:"sequence"`
	Total                int                           `json:"total"`
	ApprovalChainKey     string                        `json:"approvalChainKey,omitempty"`
	ApprovalLayer        int                           `json:"approvalLayer,omitempty"`
	ApprovalLayerTotal   int                           `json:"approvalLayerTotal,omitempty"`
	SourceDepartmentID   uint                          `json:"sourceDepartmentId,omitempty"`
	SourceDepartmentName string                        `json:"sourceDepartmentName,omitempty"`
	Status               string                        `json:"status"`
	Action               string                        `json:"action"`
	Comment              string                        `json:"comment"`
	Images               []workflowcore.FormAttachment `json:"images,omitempty"`
	HandledBy            string                        `json:"handledBy"`
	HandledByName        string                        `json:"handledByName"`
	HandledAt            int64                         `json:"handledAt"`
}

type HistorySummary struct {
	ID        string                        `json:"id"`
	EventType string                        `json:"eventType"`
	NodeID    string                        `json:"nodeId"`
	TaskID    string                        `json:"taskId"`
	ActorID   string                        `json:"actorId"`
	ActorName string                        `json:"actorName"`
	Message   string                        `json:"message"`
	Images    []workflowcore.FormAttachment `json:"images,omitempty"`
	EventTime int64                         `json:"eventTime"`
}

type NodeProgressSummary struct {
	NodeID      string `json:"nodeId"`
	NodeName    string `json:"nodeName"`
	NodeType    string `json:"nodeType"`
	GatewayMode string `json:"gatewayMode,omitempty"`
	Status      string `json:"status"`
}

type ReminderPolicy struct {
	CooldownSeconds int `json:"cooldownSeconds"`
	DailyLimit      int `json:"dailyLimit"`
}

type ReminderNodeSummary struct {
	NodeID         string   `json:"nodeId"`
	NodeName       string   `json:"nodeName"`
	AssigneeNames  []string `json:"assigneeNames"`
	AssigneeCount  int      `json:"assigneeCount"`
	CanRemind      bool     `json:"canRemind"`
	BlockedReason  string   `json:"blockedReason"`
	LastRemindedAt int64    `json:"lastRemindedAt"`
	NextAllowedAt  int64    `json:"nextAllowedAt"`
	TodayCount     int      `json:"todayCount"`
	RemainingCount int      `json:"remainingCount"`
}

type InstanceDetail struct {
	Instance         InstanceSummary                           `json:"instance"`
	Variables        map[string]interface{}                    `json:"variables"`
	Form             []workflowcore.FormField                  `json:"form"`
	FormData         map[string]interface{}                    `json:"formData"`
	FieldPermissions map[string][]workflowcore.FieldPermission `json:"fieldPermissions"`
	StartNodeID      string                                    `json:"startNodeId"`
	NodeTypes        map[string]string                         `json:"nodeTypes"`
	Nodes            []PublishedNode                           `json:"nodes"`
	Edges            []PublishedEdge                           `json:"edges"`
	Tokens           []TokenSummary                            `json:"tokens"`
	Tasks            []TaskSummary                             `json:"tasks"`
	History          []HistorySummary                          `json:"history"`
	NodeProgress     []NodeProgressSummary                     `json:"nodeProgress"`
	UserNames        map[string]string                         `json:"userNames"`
	ReminderPolicy   ReminderPolicy                            `json:"reminderPolicy"`
	ReminderNodes    []ReminderNodeSummary                     `json:"reminderNodes"`
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
