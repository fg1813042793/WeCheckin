package application

import workflowcore "wecheckin/backend/internal/workflow"

type PublishedDefinition struct {
	ID          uint                     `json:"id"`
	Key         string                   `json:"key"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Category    string                   `json:"category"`
	Version     int                      `json:"version"`
	Form        []workflowcore.FormField `json:"form"`
}

type InstanceQuery struct {
	DefinitionID uint
	Status       string
	BusinessType string
	BusinessKey  string
	StarterID    string
	Page         int
	PageSize     int
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
	Instance  InstanceSummary        `json:"instance"`
	Variables map[string]interface{} `json:"variables"`
	FormData  map[string]interface{} `json:"formData"`
	Tokens    []TokenSummary         `json:"tokens"`
	Tasks     []TaskSummary          `json:"tasks"`
	History   []HistorySummary       `json:"history"`
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
