package workflowcore

import (
	"fmt"
	"strings"
)

const CurrentSchemaVersion = 1

const (
	NodeTypeStart      = "start"
	NodeTypeApproval   = "approval"
	NodeTypeHandle     = "handle"
	NodeTypeCC         = "cc"
	NodeTypeNotify     = "notify"
	NodeTypeAutomation = "automation"
	NodeTypeTimer      = "timer"
	NodeTypeExclusive  = "exclusive"
	NodeTypeParallel   = "parallel"
	NodeTypeEnd        = "end"
)

const AutomationTypeSetVariables = "set_variables"

const (
	NotificationChannelInApp      = "in_app"
	NotificationChannelDingTalkOA = "dingtalk_oa"
)

const (
	ApprovalModeSingle      = "single"
	ApprovalModeSequential  = "sequential"
	ApprovalModeParallel    = "parallel"
	ApprovalModeCountersign = "countersign"
)

const (
	AssigneeTypeInitiator        = "initiator"
	AssigneeTypeUser             = "user"
	AssigneeTypeRole             = "role"
	AssigneeTypeDepartmentLeader = "department_leader"
	AssigneeTypeManager          = "manager"
	AssigneeTypeVariable         = "variable"
	AssigneeTypeOrgIdentity      = "org_identity"
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
	FormFieldTypeText            = "text"
	FormFieldTypeTextarea        = "textarea"
	FormFieldTypeNumber          = "number"
	FormFieldTypeSelect          = "select"
	FormFieldTypeMultiSelect     = "multi_select"
	FormFieldTypeDate            = "date"
	FormFieldTypeDateTime        = "datetime"
	FormFieldTypeUser            = "user"
	FormFieldTypeDepartment      = "department"
	FormFieldTypeAttachment      = "attachment"
	FormFieldTypeBoolean         = "boolean"
	FormFieldTypeAmount          = "amount"
	FormFieldTypePhone           = "phone"
	FormFieldTypeEmail           = "email"
	FormFieldTypeRadio           = "radio"
	FormFieldTypeCheckbox        = "checkbox"
	FormFieldTypeTime            = "time"
	FormFieldTypeDateRange       = "date_range"
	FormFieldTypeUserMulti       = "user_multi"
	FormFieldTypeDepartmentMulti = "department_multi"
	FormFieldTypeDetailList      = "detail_list"
	FormFieldTypeGroup           = "group"
	FormFieldTypeLabel           = "label"
	FormFieldTypeDescription     = "description"
	FormFieldTypeButton          = "button"
)

const (
	FieldAccessHidden = "hidden"
	FieldAccessRead   = "read"
	FieldAccessWrite  = "write"
)

const (
	FieldActionAdd    = "add"
	FieldActionDelete = "delete"
)

const (
	InitiatorScopeAll       = "all"
	InitiatorScopeSpecified = "specified"
)

const (
	StartAvailabilityAlways  = "always"
	StartAvailabilityFixed   = "fixed"
	StartAvailabilityWeekly  = "weekly"
	StartAvailabilityMonthly = "monthly"
)

const (
	StartLimitModeUnlimited = "unlimited"
	StartLimitModeLimited   = "limited"
)

const (
	StartLimitPeriodTotal        = "total"
	StartLimitPeriodDay          = "day"
	StartLimitPeriodWeek         = "week"
	StartLimitPeriodMonth        = "month"
	StartLimitPeriodAvailability = "availability"
)

const MaxStartLimitCount = 10000

const (
	StartAvailabilityStateAvailable     = "available"
	StartAvailabilityStateNotStarted    = "not_started"
	StartAvailabilityStateExpired       = "expired"
	StartAvailabilityStateOutsideWindow = "outside_window"
)

const DefaultStartAvailabilityTimezone = "Asia/Shanghai"

const (
	OptionSourceStatic = "static"
	OptionSourceAPI    = "api"
)

const (
	FormRuleMinLength           = "min_length"
	FormRuleMaxLength           = "max_length"
	FormRulePattern             = "pattern"
	FormRuleNumberRange         = "number_range"
	FormRuleDecimalPlaces       = "decimal_places"
	FormRuleSelectionCount      = "selection_count"
	FormRuleCompareField        = "compare_field"
	FormRuleColumnSum           = "column_sum"
	FormRuleConditionalRequired = "conditional_required"
)

const (
	FormRuleOperatorEQ       = "eq"
	FormRuleOperatorNE       = "ne"
	FormRuleOperatorGT       = "gt"
	FormRuleOperatorGTE      = "gte"
	FormRuleOperatorLT       = "lt"
	FormRuleOperatorLTE      = "lte"
	FormRuleOperatorEmpty    = "empty"
	FormRuleOperatorNotEmpty = "not_empty"
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
	ValidationAutomation               = "automation_invalid"
	ValidationTimer                    = "timer_invalid"
	ValidationNotification             = "notification_invalid"
	ValidationGatewayMode              = "gateway_mode_invalid"
	ValidationBranchCount              = "branch_count_invalid"
	ValidationBranchConditionRequired  = "branch_condition_required"
	ValidationMultipleDefaultBranches  = "branch_default_multiple"
	ValidationConditionInvalid         = "condition_invalid"
	ValidationFormFieldKey             = "form_field_key_invalid"
	ValidationFormFieldDuplicate       = "form_field_duplicate"
	ValidationFormFieldType            = "form_field_type_invalid"
	ValidationFormFieldOptions         = "form_field_options_invalid"
	ValidationFormFieldOptionSource    = "form_field_option_source_invalid"
	ValidationFormFieldRange           = "form_field_range_invalid"
	ValidationFormFieldSpan            = "form_field_span_invalid"
	ValidationFormFieldColumns         = "form_field_columns_invalid"
	ValidationFormFieldRows            = "form_field_rows_invalid"
	ValidationFormFieldVisibleRows     = "form_field_visible_rows_invalid"
	ValidationFormFieldLayout          = "form_field_layout_invalid"
	ValidationFormFieldHelp            = "form_field_help_invalid"
	ValidationFormFieldRules           = "form_field_rules_invalid"
	ValidationFieldPermissionField     = "field_permission_field_invalid"
	ValidationFieldPermissionAccess    = "field_permission_access_invalid"
	ValidationFieldPermissionAction    = "field_permission_action_invalid"
	ValidationFieldPermissionDuplicate = "field_permission_duplicate"
	ValidationInitiator                = "initiator_invalid"
	ValidationStartAvailability        = "start_availability_invalid"
	ValidationStartLimit               = "start_limit_invalid"
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
	Key            string               `json:"key"`
	Label          string               `json:"label"`
	Type           string               `json:"type"`
	Required       bool                 `json:"required,omitempty"`
	Default        interface{}          `json:"default,omitempty"`
	Placeholder    string               `json:"placeholder,omitempty"`
	MaxLength      int                  `json:"maxLength,omitempty"`
	Min            *float64             `json:"min,omitempty"`
	Max            *float64             `json:"max,omitempty"`
	Options        []FormOption         `json:"options,omitempty"`
	OptionSource   *FormOptionSource    `json:"optionSource,omitempty"`
	Span           int                  `json:"span,omitempty"`
	RowKey         string               `json:"rowKey,omitempty"`
	Columns        []FormField          `json:"columns,omitempty"`
	Fields         []FormField          `json:"fields,omitempty"`
	Content        string               `json:"content,omitempty"`
	Help           *FormHelp            `json:"help,omitempty"`
	MinRows        int                  `json:"minRows,omitempty"`
	MaxRows        int                  `json:"maxRows,omitempty"`
	MinVisibleRows int                  `json:"minVisibleRows,omitempty"`
	MaxVisibleRows int                  `json:"maxVisibleRows,omitempty"`
	Rules          []FormValidationRule `json:"rules,omitempty"`
}

type FormAttachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type FormValidationRule struct {
	ID        string             `json:"id"`
	Type      string             `json:"type"`
	Min       *float64           `json:"min,omitempty"`
	Max       *float64           `json:"max,omitempty"`
	Precision *int               `json:"precision,omitempty"`
	Pattern   string             `json:"pattern,omitempty"`
	Field     string             `json:"field,omitempty"`
	Column    string             `json:"column,omitempty"`
	Operator  string             `json:"operator,omitempty"`
	Value     *float64           `json:"value,omitempty"`
	When      *FormRuleCondition `json:"when,omitempty"`
	Message   string             `json:"message,omitempty"`
}

type FormRuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value,omitempty"`
}

type FormHelp struct {
	ButtonText string `json:"buttonText,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

type FormOption struct {
	Label    string       `json:"label"`
	Value    string       `json:"value"`
	Children []FormOption `json:"children,omitempty"`
}

type FormOptionSource struct {
	Type          string `json:"type"`
	URL           string `json:"url,omitempty"`
	Method        string `json:"method,omitempty"`
	ResponsePath  string `json:"responsePath,omitempty"`
	LabelField    string `json:"labelField,omitempty"`
	ValueField    string `json:"valueField,omitempty"`
	ChildrenField string `json:"childrenField,omitempty"`
}

type FieldPermission struct {
	Field   string   `json:"field"`
	Access  string   `json:"access"`
	Actions []string `json:"actions,omitempty"`
}

type Node struct {
	ID                 string                   `json:"id"`
	Type               string                   `json:"type"`
	Name               string                   `json:"name"`
	Position           *Position                `json:"position,omitempty"`
	ApprovalMode       string                   `json:"approvalMode,omitempty"`
	Assignee           *Assignee                `json:"assignee,omitempty"`
	CompletionRate     int                      `json:"completionRate,omitempty"`
	GatewayMode        string                   `json:"gatewayMode,omitempty"`
	FormPermissions    []FieldPermission        `json:"formPermissions,omitempty"`
	Automation         *AutomationConfig        `json:"automation,omitempty"`
	Timer              *TimerConfig             `json:"timer,omitempty"`
	Notification       *NotificationConfig      `json:"notification,omitempty"`
	ResultNotification *NotificationConfig      `json:"resultNotification,omitempty"`
	Initiator          *InitiatorConfig         `json:"initiator,omitempty"`
	Availability       *StartAvailabilityConfig `json:"availability,omitempty"`
	StartLimit         *StartLimitConfig        `json:"startLimit,omitempty"`
}

type NotificationConfig struct {
	Enabled  bool     `json:"enabled"`
	Channels []string `json:"channels,omitempty"`
	Title    string   `json:"title,omitempty"`
	Content  string   `json:"content,omitempty"`
}

type InitiatorConfig struct {
	Scope           string `json:"scope"`
	UserIDs         []uint `json:"userIds,omitempty"`
	DepartmentIDs   []uint `json:"departmentIds,omitempty"`
	ExcludedUserIDs []uint `json:"excludedUserIds,omitempty"`
}

type StartAvailabilityConfig struct {
	Mode               string `json:"mode"`
	Timezone           string `json:"timezone,omitempty"`
	StartsAt           int64  `json:"startsAt,omitempty"`
	EndsAt             int64  `json:"endsAt,omitempty"`
	EffectiveStartDate string `json:"effectiveStartDate,omitempty"`
	EffectiveEndDate   string `json:"effectiveEndDate,omitempty"`
	Weekdays           []int  `json:"weekdays,omitempty"`
	MonthDays          []int  `json:"monthDays,omitempty"`
	LastDayOfMonth     bool   `json:"lastDayOfMonth,omitempty"`
	DailyStartTime     string `json:"dailyStartTime,omitempty"`
	DailyEndTime       string `json:"dailyEndTime,omitempty"`
}

type StartLimitConfig struct {
	Mode     string `json:"mode"`
	Period   string `json:"period,omitempty"`
	MaxCount int    `json:"maxCount,omitempty"`
}

type StartLimitWindow struct {
	PeriodKey string
	StartsAt  int64
	EndsAt    int64
}

type AutomationConfig struct {
	Type      string                 `json:"type"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type TimerConfig struct {
	DelaySeconds int64 `json:"delaySeconds"`
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
	ID           string     `json:"id"`
	Source       string     `json:"source"`
	Target       string     `json:"target"`
	SourceHandle string     `json:"sourceHandle,omitempty"`
	TargetHandle string     `json:"targetHandle,omitempty"`
	Name         string     `json:"name,omitempty"`
	Default      bool       `json:"default,omitempty"`
	Condition    *Condition `json:"condition,omitempty"`
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
