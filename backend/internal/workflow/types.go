package workflow

import (
	"fmt"
	"strings"
)

const CurrentSchemaVersion = 1

const (
	NodeTypeStart     = "start"
	NodeTypeApproval  = "approval"
	NodeTypeExclusive = "exclusive"
	NodeTypeParallel  = "parallel"
	NodeTypeEnd       = "end"
)

const (
	ApprovalModeSingle      = "single"
	ApprovalModeSequential  = "sequential"
	ApprovalModeParallel    = "parallel"
	ApprovalModeCountersign = "countersign"
)

const (
	AssigneeTypeUser             = "user"
	AssigneeTypeRole             = "role"
	AssigneeTypeDepartmentLeader = "department_leader"
	AssigneeTypeManager          = "manager"
	AssigneeTypeVariable         = "variable"
)

const (
	GatewayModeSplit = "split"
	GatewayModeJoin  = "join"
)

const (
	ConditionEQ  = "eq"
	ConditionNE  = "ne"
	ConditionGT  = "gt"
	ConditionGTE = "gte"
	ConditionLT  = "lt"
	ConditionLTE = "lte"
)

const (
	FormFieldTypeText        = "text"
	FormFieldTypeTextarea    = "textarea"
	FormFieldTypeNumber      = "number"
	FormFieldTypeSelect      = "select"
	FormFieldTypeMultiSelect = "multi_select"
	FormFieldTypeDate        = "date"
	FormFieldTypeDateTime    = "datetime"
	FormFieldTypeUser        = "user"
	FormFieldTypeDepartment  = "department"
	FormFieldTypeAttachment  = "attachment"
	FormFieldTypeBoolean     = "boolean"
)

const (
	FieldAccessHidden = "hidden"
	FieldAccessRead   = "read"
	FieldAccessWrite  = "write"
)

const (
	ValidationSchemaVersion            = "schema_version_invalid"
	ValidationDefinitionKey            = "definition_key_invalid"
	ValidationMissingStart             = "start_missing"
	ValidationMultipleStarts           = "start_multiple"
	ValidationMissingEnd               = "end_missing"
	ValidationDuplicateNode            = "node_duplicate"
	ValidationDuplicateEdge            = "edge_duplicate"
	ValidationEdgeEndpoint             = "edge_endpoint_invalid"
	ValidationIncomingRequired         = "incoming_required"
	ValidationOutgoingRequired         = "outgoing_required"
	ValidationUnreachableNode          = "node_unreachable"
	ValidationNoPathToEnd              = "end_unreachable"
	ValidationAssigneeRequired         = "assignee_required"
	ValidationApprovalMode             = "approval_mode_invalid"
	ValidationCompletionRate           = "completion_rate_invalid"
	ValidationGatewayMode              = "gateway_mode_invalid"
	ValidationBranchCount              = "branch_count_invalid"
	ValidationBranchConditionRequired  = "branch_condition_required"
	ValidationMultipleDefaultBranches  = "branch_default_multiple"
	ValidationConditionInvalid         = "condition_invalid"
	ValidationFormFieldKey             = "form_field_key_invalid"
	ValidationFormFieldDuplicate       = "form_field_duplicate"
	ValidationFormFieldType            = "form_field_type_invalid"
	ValidationFormFieldOptions         = "form_field_options_invalid"
	ValidationFormFieldRange           = "form_field_range_invalid"
	ValidationFieldPermissionField     = "field_permission_field_invalid"
	ValidationFieldPermissionAccess    = "field_permission_access_invalid"
	ValidationFieldPermissionDuplicate = "field_permission_duplicate"
)

type Definition struct {
	SchemaVersion int         `json:"schemaVersion"`
	Key           string      `json:"key"`
	Name          string      `json:"name"`
	Form          []FormField `json:"form,omitempty"`
	Nodes         []Node      `json:"nodes"`
	Edges         []Edge      `json:"edges"`
}

type FormField struct {
	Key         string       `json:"key"`
	Label       string       `json:"label"`
	Type        string       `json:"type"`
	Required    bool         `json:"required,omitempty"`
	Default     interface{}  `json:"default,omitempty"`
	Placeholder string       `json:"placeholder,omitempty"`
	MaxLength   int          `json:"maxLength,omitempty"`
	Min         *float64     `json:"min,omitempty"`
	Max         *float64     `json:"max,omitempty"`
	Options     []FormOption `json:"options,omitempty"`
}

type FormOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type FieldPermission struct {
	Field  string `json:"field"`
	Access string `json:"access"`
}

type Node struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	Name            string            `json:"name"`
	Position        *Position         `json:"position,omitempty"`
	ApprovalMode    string            `json:"approvalMode,omitempty"`
	Assignee        *Assignee         `json:"assignee,omitempty"`
	CompletionRate  int               `json:"completionRate,omitempty"`
	GatewayMode     string            `json:"gatewayMode,omitempty"`
	FormPermissions []FieldPermission `json:"formPermissions,omitempty"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Assignee struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Edge struct {
	ID        string     `json:"id"`
	Source    string     `json:"source"`
	Target    string     `json:"target"`
	Name      string     `json:"name,omitempty"`
	Default   bool       `json:"default,omitempty"`
	Condition *Condition `json:"condition,omitempty"`
}

type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"nodeId,omitempty"`
	EdgeID  string `json:"edgeId,omitempty"`
}

type ValidationErrors []ValidationError

func (errors ValidationErrors) Error() string {
	if len(errors) == 0 {
		return ""
	}
	messages := make([]string, 0, len(errors))
	for _, item := range errors {
		messages = append(messages, item.Message)
	}
	return fmt.Sprintf("流程定义校验失败：%s", strings.Join(messages, "；"))
}
