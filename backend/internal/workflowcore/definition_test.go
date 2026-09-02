package workflowcore

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestValidateDefinitionAcceptsLinearApproval(t *testing.T) {
	definition := validLinearDefinition()

	errors := ValidateDefinition(definition)
	if len(errors) != 0 {
		t.Fatalf("expected valid definition, got errors: %#v", errors)
	}
}

func TestDefinitionJSONPreservesEdgeHandles(t *testing.T) {
	definition := validLinearDefinition()
	definition.Edges[0].SourceHandle = "right"
	definition.Edges[0].TargetHandle = "left"

	encoded, err := json.Marshal(definition)
	if err != nil {
		t.Fatalf("encode definition: %v", err)
	}
	var decoded Definition
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode definition: %v", err)
	}
	if decoded.Edges[0].SourceHandle != "right" || decoded.Edges[0].TargetHandle != "left" {
		t.Fatalf("expected edge handles to survive JSON round trip, got %#v", decoded.Edges[0])
	}
}

func TestValidateDefinitionAcceptsInitiatorAssigneeWithoutValue(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[1].Assignee = &Assignee{Type: AssigneeTypeInitiator}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected initiator assignee to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionAcceptsInitiatorDepartments(t *testing.T) {
	tests := []struct {
		name      string
		initiator InitiatorConfig
	}{
		{name: "specified users", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified, UserIDs: []uint{7, 8}}},
		{name: "specified departments", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified, DepartmentIDs: []uint{3, 5}}},
		{name: "users and departments", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified, UserIDs: []uint{7}, DepartmentIDs: []uint{3}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].Initiator = &test.initiator
			if errors := ValidateDefinition(definition); len(errors) != 0 {
				t.Fatalf("expected valid initiator range, got %#v", errors)
			}
		})
	}
}

func TestValidateDefinitionRejectsInvalidInitiatorRanges(t *testing.T) {
	tests := []struct {
		name      string
		initiator InitiatorConfig
	}{
		{name: "empty specified range", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified}},
		{name: "zero department id", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified, DepartmentIDs: []uint{0}}},
		{name: "duplicate department id", initiator: InitiatorConfig{Scope: InitiatorScopeSpecified, DepartmentIDs: []uint{3, 3}}},
		{name: "all users with departments", initiator: InitiatorConfig{Scope: InitiatorScopeAll, DepartmentIDs: []uint{3}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].Initiator = &test.initiator
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationInitiator) {
				t.Fatalf("expected %s, got %#v", ValidationInitiator, errors)
			}
		})
	}
}

func TestValidateDefinitionAcceptsStartAvailabilityModes(t *testing.T) {
	tests := []struct {
		name         string
		availability *StartAvailabilityConfig
	}{
		{name: "missing defaults to always"},
		{name: "always", availability: &StartAvailabilityConfig{Mode: StartAvailabilityAlways}},
		{name: "fixed", availability: &StartAvailabilityConfig{
			Mode: StartAvailabilityFixed, Timezone: "Asia/Shanghai", StartsAt: 1788224400000, EndsAt: 1788310800000,
		}},
		{name: "weekly", availability: &StartAvailabilityConfig{
			Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1, 3, 5},
			DailyStartTime: "09:00", DailyEndTime: "18:00", EffectiveStartDate: "2026-09-01", EffectiveEndDate: "2026-12-31",
		}},
		{name: "monthly", availability: &StartAvailabilityConfig{
			Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{1, 15, 31}, LastDayOfMonth: true,
			DailyStartTime: "09:00", DailyEndTime: "18:00",
		}},
		{name: "monthly last day only", availability: &StartAvailabilityConfig{
			Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", LastDayOfMonth: true,
			DailyStartTime: "09:00", DailyEndTime: "18:00",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].Availability = test.availability
			if errors := ValidateDefinition(definition); len(errors) != 0 {
				t.Fatalf("expected valid start availability, got %#v", errors)
			}
		})
	}
}

func TestValidateDefinitionRejectsInvalidStartAvailability(t *testing.T) {
	tests := []struct {
		name         string
		availability StartAvailabilityConfig
	}{
		{name: "unknown mode", availability: StartAvailabilityConfig{Mode: "daily"}},
		{name: "fixed reversed", availability: StartAvailabilityConfig{Mode: StartAvailabilityFixed, StartsAt: 200, EndsAt: 100}},
		{name: "weekly without weekday", availability: StartAvailabilityConfig{Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", DailyStartTime: "09:00", DailyEndTime: "18:00"}},
		{name: "weekly duplicate weekday", availability: StartAvailabilityConfig{Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1, 1}, DailyStartTime: "09:00", DailyEndTime: "18:00"}},
		{name: "monthly without day", availability: StartAvailabilityConfig{Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", DailyStartTime: "09:00", DailyEndTime: "18:00"}},
		{name: "monthly invalid day", availability: StartAvailabilityConfig{Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{32}, DailyStartTime: "09:00", DailyEndTime: "18:00"}},
		{name: "invalid daily range", availability: StartAvailabilityConfig{Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1}, DailyStartTime: "18:00", DailyEndTime: "09:00"}},
		{name: "invalid effective range", availability: StartAvailabilityConfig{Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{1}, DailyStartTime: "09:00", DailyEndTime: "18:00", EffectiveStartDate: "2026-12-31", EffectiveEndDate: "2026-01-01"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].Availability = &test.availability
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationStartAvailability) {
				t.Fatalf("expected %s, got %#v", ValidationStartAvailability, errors)
			}
		})
	}
}

func TestEvaluateStartAvailability(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	at := func(year int, month time.Month, day, hour, minute int) time.Time {
		return time.Date(year, month, day, hour, minute, 0, 0, location)
	}

	fixedStart := at(2026, time.September, 1, 9, 0)
	fixed := &StartAvailabilityConfig{Mode: StartAvailabilityFixed, Timezone: "Asia/Shanghai", StartsAt: fixedStart.UnixMilli(), EndsAt: fixedStart.Add(8 * time.Hour).UnixMilli()}
	if state := EvaluateStartAvailability(fixed, fixedStart.Add(-time.Minute)); state != StartAvailabilityStateNotStarted {
		t.Fatalf("fixed state before start = %q", state)
	}
	if state := EvaluateStartAvailability(fixed, fixedStart); state != StartAvailabilityStateAvailable {
		t.Fatalf("fixed state at start = %q", state)
	}
	if state := EvaluateStartAvailability(fixed, fixedStart.Add(8*time.Hour)); state != StartAvailabilityStateExpired {
		t.Fatalf("fixed state at end = %q", state)
	}

	weekly := &StartAvailabilityConfig{Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{2}, DailyStartTime: "09:00", DailyEndTime: "18:00"}
	if state := EvaluateStartAvailability(weekly, at(2026, time.September, 1, 10, 0)); state != StartAvailabilityStateAvailable {
		t.Fatalf("weekly state in window = %q", state)
	}
	if state := EvaluateStartAvailability(weekly, at(2026, time.September, 1, 18, 0)); state != StartAvailabilityStateOutsideWindow {
		t.Fatalf("weekly state at end = %q", state)
	}

	monthly := &StartAvailabilityConfig{Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{31}, DailyStartTime: "09:00", DailyEndTime: "18:00"}
	if state := EvaluateStartAvailability(monthly, at(2026, time.September, 30, 10, 0)); state != StartAvailabilityStateOutsideWindow {
		t.Fatalf("monthly day 31 must skip short month, got %q", state)
	}
	monthly.LastDayOfMonth = true
	if state := EvaluateStartAvailability(monthly, at(2026, time.September, 30, 10, 0)); state != StartAvailabilityStateAvailable {
		t.Fatalf("monthly last day state = %q", state)
	}
}

func TestValidateDefinitionAcceptsExtendedWorkflowNodes(t *testing.T) {
	definition := extendedNodeDefinition(t)
	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected extended nodes to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionRejectsInvalidExtendedNodeConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		nodeJSON   string
		expectCode string
	}{
		{name: "handle without assignee", nodeJSON: `{"id":"work","type":"handle","name":"填写资料"}`, expectCode: "assignee_required"},
		{name: "cc without assignee", nodeJSON: `{"id":"work","type":"cc","name":"通知相关人"}`, expectCode: "assignee_required"},
		{name: "automation without variables", nodeJSON: `{"id":"work","type":"automation","name":"写入变量","automation":{"type":"set_variables","variables":{}}}`, expectCode: "automation_invalid"},
		{name: "timer without positive delay", nodeJSON: `{"id":"work","type":"timer","name":"等待","timer":{"delaySeconds":0}}`, expectCode: "timer_invalid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var definition Definition
			source := `{"schemaVersion":1,"key":"extended_test","name":"扩展节点测试","nodes":[{"id":"start","type":"start","name":"开始"},` + test.nodeJSON + `,{"id":"end","type":"end","name":"结束"}],"edges":[{"id":"e1","source":"start","target":"work"},{"id":"e2","source":"work","target":"end"}]}`
			if err := json.Unmarshal([]byte(source), &definition); err != nil {
				t.Fatalf("decode definition: %v", err)
			}
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, test.expectCode) {
				t.Fatalf("expected validation code %q, got %#v", test.expectCode, errors)
			}
		})
	}
}

func TestValidateDefinitionAcceptsGenericNotificationConfiguration(t *testing.T) {
	definition := Definition{
		SchemaVersion: CurrentSchemaVersion,
		Key:           "notification_flow",
		Name:          "通知流程",
		Nodes: []Node{
			{ID: "start", Type: NodeTypeStart, Name: "开始"},
			{
				ID: "notify", Type: NodeTypeNotify, Name: "通知财务",
				Assignee: &Assignee{Type: AssigneeTypeRole, Value: "finance"},
				Notification: &NotificationConfig{
					Enabled: true, Channels: []string{NotificationChannelInApp, NotificationChannelDingTalkOA},
					Title: "{{workflowName}}", Content: "{{starterName}} 发起的流程已到达 {{nodeName}}，实例 {{instanceId}}",
				},
			},
			{ID: "end", Type: NodeTypeEnd, Name: "结束"},
		},
		Edges: []Edge{{ID: "e1", Source: "start", Target: "notify"}, {ID: "e2", Source: "notify", Target: "end"}},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("valid notification configuration rejected: %#v", errors)
	}
}

func TestValidateDefinitionRejectsInvalidNotificationConfiguration(t *testing.T) {
	valid := func() Definition {
		return Definition{
			SchemaVersion: CurrentSchemaVersion,
			Key:           "notification_flow",
			Name:          "通知流程",
			Nodes: []Node{
				{ID: "start", Type: NodeTypeStart, Name: "开始"},
				{
					ID: "notify", Type: NodeTypeNotify, Name: "通知财务",
					Assignee:     &Assignee{Type: AssigneeTypeRole, Value: "finance"},
					Notification: &NotificationConfig{Enabled: true, Channels: []string{NotificationChannelInApp}, Title: "流程通知", Content: "流程已到达 {{nodeName}}"},
				},
				{ID: "end", Type: NodeTypeEnd, Name: "结束"},
			},
			Edges: []Edge{{ID: "e1", Source: "start", Target: "notify"}, {ID: "e2", Source: "notify", Target: "end"}},
		}
	}
	tests := []struct {
		name   string
		mutate func(*Node)
	}{
		{name: "missing config", mutate: func(node *Node) { node.Notification = nil }},
		{name: "disabled notify", mutate: func(node *Node) { node.Notification.Enabled = false }},
		{name: "missing channels", mutate: func(node *Node) { node.Notification.Channels = nil }},
		{name: "duplicate channels", mutate: func(node *Node) {
			node.Notification.Channels = []string{NotificationChannelInApp, NotificationChannelInApp}
		}},
		{name: "unknown channel", mutate: func(node *Node) { node.Notification.Channels = []string{"email"} }},
		{name: "blank title", mutate: func(node *Node) { node.Notification.Title = " " }},
		{name: "long title", mutate: func(node *Node) { node.Notification.Title = strings.Repeat("题", 257) }},
		{name: "long content", mutate: func(node *Node) { node.Notification.Content = strings.Repeat("文", 2001) }},
		{name: "unknown template token", mutate: func(node *Node) { node.Notification.Content = "{{form.secret}}" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := valid()
			test.mutate(&definition.Nodes[1])
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationNotification) {
				t.Fatalf("expected notification validation error, got %#v", errors)
			}
		})
	}
}

func TestValidateDefinitionKeepsLegacyNodesWithoutNotificationConfiguration(t *testing.T) {
	definition := extendedNodeDefinition(t)
	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("legacy nodes without notification config must remain valid: %#v", errors)
	}
}

func TestValidateDefinitionRejectsInvalidStructure(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Definition)
		expectCode string
	}{
		{
			name: "missing start",
			mutate: func(definition *Definition) {
				definition.Nodes = definition.Nodes[1:]
				definition.Edges = definition.Edges[1:]
			},
			expectCode: ValidationMissingStart,
		},
		{
			name: "duplicate start",
			mutate: func(definition *Definition) {
				definition.Nodes = append(definition.Nodes, Node{ID: "start_2", Type: NodeTypeStart, Name: "另一个开始"})
			},
			expectCode: ValidationMultipleStarts,
		},
		{
			name: "approval without assignee",
			mutate: func(definition *Definition) {
				definition.Nodes[1].Assignee = nil
			},
			expectCode: ValidationAssigneeRequired,
		},
		{
			name: "unreachable node",
			mutate: func(definition *Definition) {
				definition.Nodes = append(definition.Nodes, Node{
					ID:           "orphan",
					Type:         NodeTypeApproval,
					Name:         "孤立审批",
					ApprovalMode: ApprovalModeSingle,
					Assignee:     &Assignee{Type: AssigneeTypeUser, Value: "u1"},
				})
			},
			expectCode: ValidationUnreachableNode,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			test.mutate(&definition)
			errors := ValidateDefinition(definition)
			if !hasValidationCode(errors, test.expectCode) {
				t.Fatalf("expected validation code %q, got %#v", test.expectCode, errors)
			}
		})
	}
}

func TestValidateDefinitionRequiresCompleteExclusiveBranches(t *testing.T) {
	definition := Definition{
		SchemaVersion: 1,
		Key:           "score_approval",
		Name:          "分数审批",
		Nodes: []Node{
			{ID: "start", Type: NodeTypeStart, Name: "开始"},
			{ID: "decision", Type: NodeTypeExclusive, Name: "分数判断", GatewayMode: GatewayModeSplit},
			{ID: "approved", Type: NodeTypeEnd, Name: "通过"},
			{ID: "rejected", Type: NodeTypeEnd, Name: "退回"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "decision"},
			{ID: "e2", Source: "decision", Target: "approved", Condition: &Condition{Field: "score", Operator: ConditionGTE, Value: 80}},
			{ID: "e3", Source: "decision", Target: "rejected"},
		},
	}

	errors := ValidateDefinition(definition)
	if !hasValidationCode(errors, ValidationBranchConditionRequired) {
		t.Fatalf("expected missing branch condition error, got %#v", errors)
	}

	definition.Edges[2].Default = true
	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected valid default branch, got %#v", errors)
	}
}

func TestCompileBPMNProducesParseableFlowableProcess(t *testing.T) {
	definition := validLinearDefinition()
	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("compile BPMN: %v", err)
	}
	var parsed interface{}
	if err := xml.Unmarshal(bpmn, &parsed); err != nil {
		t.Fatalf("generated BPMN is not XML: %v\n%s", err, bpmn)
	}
	text := string(bpmn)
	for _, expected := range []string{
		`<process id="leave_approval"`,
		`<startEvent id="start"`,
		`<userTask id="manager"`,
		`flowable:assignee="${workflowApprover_manager}"`,
		`<endEvent id="end"`,
		`<sequenceFlow id="e1" sourceRef="start" targetRef="manager"`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated BPMN missing %q\n%s", expected, text)
		}
	}
}

func TestCompileBPMNIncludesExtendedWorkflowNodes(t *testing.T) {
	bpmn, err := CompileBPMN(extendedNodeDefinition(t))
	if err != nil {
		t.Fatalf("compile BPMN: %v", err)
	}
	text := string(bpmn)
	for _, expected := range []string{
		`<userTask id="handle" name="填写资料"`,
		`<serviceTask id="cc" name="通知相关人"`,
		`<serviceTask id="automation" name="写入变量"`,
		`<intermediateCatchEvent id="timer" name="等待 30 秒">`,
		`<timeDuration>PT30S</timeDuration>`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated BPMN missing %q\n%s", expected, text)
		}
	}
}

func TestCompileBPMNIncludesNotifyNodeAndNotificationAttributes(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[1].Notification = &NotificationConfig{
		Enabled: true, Channels: []string{NotificationChannelInApp, NotificationChannelDingTalkOA},
		Title: "{{workflowName}}", Content: "请处理 {{nodeName}}",
	}
	notify := Node{
		ID: "notify", Type: NodeTypeNotify, Name: "结果通知",
		Assignee:     &Assignee{Type: AssigneeTypeVariable, Value: "notifyUserIds"},
		Notification: &NotificationConfig{Enabled: true, Channels: []string{NotificationChannelInApp}, Title: "结果", Content: "流程 {{instanceId}} 已结束"},
	}
	definition.Nodes = append(definition.Nodes[:2], notify, definition.Nodes[2])
	definition.Edges = []Edge{
		{ID: "e1", Source: "start", Target: "manager"},
		{ID: "e2", Source: "manager", Target: "notify"},
		{ID: "e3", Source: "notify", Target: "end"},
	}

	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("compile notification BPMN: %v", err)
	}
	text := string(bpmn)
	for _, expected := range []string{
		`<serviceTask id="notify" name="结果通知"`,
		`flowable:topic="wecheckin-notify"`,
		`flowable:notificationEnabled="true"`,
		`flowable:notificationChannels="in_app,dingtalk_oa"`,
		`flowable:notificationTitle="{{workflowName}}"`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated notification BPMN missing %q\n%s", expected, text)
		}
	}
}

func TestCompileBPMNUsesGoResolvedVariablesForMultiApprovers(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[1].ID = "manager-review.1"
	definition.Nodes[1].ApprovalMode = ApprovalModeParallel
	definition.Nodes[1].Assignee = &Assignee{Type: AssigneeTypeRole, Value: "finance"}
	definition.Edges[0].Target = "manager-review.1"
	definition.Edges[1].Source = "manager-review.1"

	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("compile BPMN: %v", err)
	}
	text := string(bpmn)
	if !strings.Contains(text, `flowable:collection="${workflowApprovers_manager_review_1}"`) {
		t.Fatalf("expected node-scoped approver collection, got\n%s", text)
	}
	if strings.Contains(text, "workflowAssigneeResolver") {
		t.Fatalf("BPMN must not depend on a custom Java bean, got\n%s", text)
	}
}

func TestCompileBPMNIncludesGatewayConditionsAndDefaultFlow(t *testing.T) {
	definition := Definition{
		SchemaVersion: 1,
		Key:           "score_route",
		Name:          "分数路由",
		Nodes: []Node{
			{ID: "start", Type: NodeTypeStart, Name: "开始"},
			{ID: "decision", Type: NodeTypeExclusive, Name: "判断", GatewayMode: GatewayModeSplit},
			{ID: "passed", Type: NodeTypeEnd, Name: "通过"},
			{ID: "fallback", Type: NodeTypeEnd, Name: "默认"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "decision"},
			{ID: "e2", Source: "decision", Target: "passed", Condition: &Condition{Field: "score", Operator: ConditionGTE, Value: 80}},
			{ID: "e3", Source: "decision", Target: "fallback", Default: true},
		},
	}

	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("compile BPMN: %v", err)
	}
	text := string(bpmn)
	for _, expected := range []string{
		`<exclusiveGateway id="decision" name="判断" default="e3"`,
		`<conditionExpression xsi:type="tFormalExpression">${score &gt;= 80}</conditionExpression>`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated BPMN missing %q\n%s", expected, text)
		}
	}
}

func TestCompileBPMNIncludesCountersignConfiguration(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[1].ApprovalMode = ApprovalModeCountersign
	definition.Nodes[1].Assignee = &Assignee{Type: AssigneeTypeVariable, Value: "reviewers"}
	definition.Nodes[1].CompletionRate = 70

	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("compile BPMN: %v", err)
	}
	text := string(bpmn)
	for _, expected := range []string{
		`flowable:assignee="${assignee}"`,
		`<multiInstanceLoopCharacteristics isSequential="false" flowable:collection="${reviewers}" flowable:elementVariable="assignee">`,
		`<completionCondition>${nrOfCompletedInstances / nrOfInstances &gt;= 0.7}</completionCondition>`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("generated BPMN missing %q\n%s", expected, text)
		}
	}
}

func TestValidateDefinitionAcceptsFormSchemaAndNodePermissions(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "reason", Label: "申请原因", Type: FormFieldTypeTextarea, Required: true, MaxLength: 500},
		{Key: "days", Label: "天数", Type: FormFieldTypeNumber, Min: numberPointer(0.5), Max: numberPointer(30)},
		{Key: "leaveType", Label: "请假类型", Type: FormFieldTypeSelect, Options: []FormOption{{Label: "年假", Value: "annual"}}},
	}
	definition.Nodes[1].FormPermissions = []FieldPermission{
		{Field: "reason", Access: FieldAccessRead},
		{Field: "days", Access: FieldAccessWrite},
		{Field: "leaveType", Access: FieldAccessHidden},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected valid OA form definition, got %#v", errors)
	}
}

func TestValidateDefinitionAcceptsTreeOptionsAndAPISource(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{
			Key:   "region",
			Label: "所属区域",
			Type:  FormFieldTypeSelect,
			Options: []FormOption{
				{Label: "华东", Value: "east", Children: []FormOption{
					{Label: "上海", Value: "shanghai"},
					{Label: "杭州", Value: "hangzhou"},
				}},
				{Label: "华南", Value: "south", Children: []FormOption{{Label: "深圳", Value: "shenzhen"}}},
			},
		},
		{
			Key:   "department",
			Label: "所属部门",
			Type:  FormFieldTypeMultiSelect,
			OptionSource: &FormOptionSource{
				Type:          OptionSourceAPI,
				URL:           "/api/v2/admin/departments/tree",
				Method:        "GET",
				ResponsePath:  "data",
				LabelField:    "name",
				ValueField:    "id",
				ChildrenField: "children",
			},
		},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected tree options and API source to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionAcceptsCommonFormFieldSpans(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "legacy", Label: "旧字段", Type: FormFieldTypeText},
		{Key: "quarter", Label: "四分之一行", Type: FormFieldTypeText, Span: 6},
		{Key: "third", Label: "三分之一行", Type: FormFieldTypeText, Span: 8},
		{Key: "half", Label: "二分之一行", Type: FormFieldTypeText, Span: 12},
		{Key: "full", Label: "整行", Type: FormFieldTypeTextarea, Span: 24},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected common form field spans to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionAcceptsDetailListFormFieldAndRowActions(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{objectiveDetailListField()}
	definition.Nodes[1].FormPermissions = []FieldPermission{
		{Field: "objectives", Access: FieldAccessWrite, Actions: []string{FieldActionAdd, FieldActionDelete}},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected detail list field definition to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionAcceptsFormLayoutComponents(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{{
		Key: "basic_group", Label: "基本信息", Type: FormFieldTypeGroup, Span: 24,
		Help: &FormHelp{ButtonText: "查看分组说明", Title: "基本信息说明", Content: "请如实填写。"},
		Fields: []FormField{
			{Key: "basic_label", Label: "申请信息", Type: FormFieldTypeLabel, Span: 24},
			{Key: "basic_tip", Label: "填写提示", Type: FormFieldTypeDescription, Content: "请核对后提交。", Span: 24},
			{Key: "reason", Label: "申请事由", Type: FormFieldTypeTextarea, Required: true,
				Help: &FormHelp{Title: "事由说明", Content: "请输入完整事由。"}},
			{Key: "rules", Label: "查看填写规则", Type: FormFieldTypeButton, Span: 24,
				Help: &FormHelp{Title: "填写规则", Content: "仅填写本次申请。"}},
			objectiveDetailListField(),
		},
	}}
	definition.Nodes[0].FormPermissions = []FieldPermission{
		{Field: "reason", Access: FieldAccessWrite},
		{Field: "objectives", Access: FieldAccessWrite, Actions: []string{FieldActionAdd, FieldActionDelete}},
	}

	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected layout form to be valid, got %#v", errors)
	}
}

func TestValidateDefinitionRejectsInvalidFormSchema(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Definition)
		expectCode string
	}{
		{
			name: "duplicate field key",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{
					{Key: "reason", Label: "原因", Type: FormFieldTypeText},
					{Key: "reason", Label: "补充原因", Type: FormFieldTypeTextarea},
				}
			},
			expectCode: ValidationFormFieldDuplicate,
		},
		{
			name: "nested form group",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "outer", Label: "外层", Type: FormFieldTypeGroup, Fields: []FormField{
					{Key: "inner", Label: "内层", Type: FormFieldTypeGroup, Fields: []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText}}},
				}}}
			},
			expectCode: ValidationFormFieldLayout,
		},
		{
			name: "duplicate key across group",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{
					{Key: "reason", Label: "原因", Type: FormFieldTypeText},
					{Key: "group", Label: "分组", Type: FormFieldTypeGroup, Fields: []FormField{{Key: "reason", Label: "组内原因", Type: FormFieldTypeTextarea}}},
				}
			},
			expectCode: ValidationFormFieldDuplicate,
		},
		{
			name: "empty description content",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "tip", Label: "提示", Type: FormFieldTypeDescription}}
			},
			expectCode: ValidationFormFieldLayout,
		},
		{
			name: "button without help",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "rules", Label: "查看规则", Type: FormFieldTypeButton}}
			},
			expectCode: ValidationFormFieldHelp,
		},
		{
			name: "data field with empty help",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText, Help: &FormHelp{Title: "原因说明"}}}
			},
			expectCode: ValidationFormFieldHelp,
		},
		{
			name: "fields on data field",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText, Fields: []FormField{{Key: "nested", Label: "子字段", Type: FormFieldTypeText}}}}
			},
			expectCode: ValidationFormFieldLayout,
		},
		{
			name: "select without options",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "type", Label: "类型", Type: FormFieldTypeSelect}}
			},
			expectCode: ValidationFormFieldOptions,
		},
		{
			name: "tree option duplicate value",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{
					Key: "region", Label: "区域", Type: FormFieldTypeSelect,
					Options: []FormOption{{Label: "华东", Value: "east", Children: []FormOption{{Label: "上海", Value: "east"}}}},
				}}
			},
			expectCode: ValidationFormFieldOptions,
		},
		{
			name: "api source external URL",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{
					Key: "department", Label: "部门", Type: FormFieldTypeSelect,
					OptionSource: &FormOptionSource{Type: OptionSourceAPI, URL: "https://example.com/options", Method: "GET", ResponsePath: "data", LabelField: "name", ValueField: "id"},
				}}
			},
			expectCode: ValidationFormFieldOptionSource,
		},
		{
			name: "api source on non-option field",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{
					Key: "reason", Label: "原因", Type: FormFieldTypeText,
					OptionSource: &FormOptionSource{Type: OptionSourceAPI, URL: "/api/v2/admin/options", Method: "GET", ResponsePath: "data", LabelField: "name", ValueField: "id"},
				}}
			},
			expectCode: ValidationFormFieldOptionSource,
		},
		{
			name: "permission references unknown field",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText}}
				definition.Nodes[1].FormPermissions = []FieldPermission{{Field: "missing", Access: FieldAccessWrite}}
			},
			expectCode: ValidationFieldPermissionField,
		},
		{
			name: "invalid permission access",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText}}
				definition.Nodes[1].FormPermissions = []FieldPermission{{Field: "reason", Access: "delete"}}
			},
			expectCode: ValidationFieldPermissionAccess,
		},
		{
			name: "invalid field span",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "reason", Label: "原因", Type: FormFieldTypeText, Span: 7}}
			},
			expectCode: ValidationFormFieldSpan,
		},
		{
			name: "detail list without columns",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "objectives", Label: "我的目标", Type: FormFieldTypeDetailList}}
			},
			expectCode: ValidationFormFieldColumns,
		},
		{
			name: "invalid detail list row bounds",
			mutate: func(definition *Definition) {
				field := objectiveDetailListField()
				field.MinRows = 3
				field.MaxRows = 2
				definition.Form = []FormField{field}
			},
			expectCode: ValidationFormFieldRows,
		},
		{
			name: "detail list row key conflicts with column",
			mutate: func(definition *Definition) {
				field := objectiveDetailListField()
				field.RowKey = "target"
				definition.Form = []FormField{field}
			},
			expectCode: ValidationFormFieldColumns,
		},
		{
			name: "invalid detail list action",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{objectiveDetailListField()}
				definition.Nodes[1].FormPermissions = []FieldPermission{{Field: "objectives", Access: FieldAccessWrite, Actions: []string{"archive"}}}
			},
			expectCode: ValidationFieldPermissionAction,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			test.mutate(&definition)
			errors := ValidateDefinition(definition)
			if !hasValidationCode(errors, test.expectCode) {
				t.Fatalf("expected validation code %q, got %#v", test.expectCode, errors)
			}
		})
	}
}

func TestValidateFormDataSupportsLayoutGroups(t *testing.T) {
	fields := []FormField{{
		Key: "group", Label: "分组", Type: FormFieldTypeGroup,
		Fields: []FormField{
			{Key: "tip", Label: "提示", Type: FormFieldTypeDescription, Content: "请填写"},
			{Key: "rules", Label: "查看规则", Type: FormFieldTypeButton, Help: &FormHelp{Title: "规则", Content: "请如实填写"}},
			{Key: "reason", Label: "原因", Type: FormFieldTypeText, Required: true},
		},
	}}

	if err := ValidateFormData(fields, map[string]interface{}{}, false); err == nil {
		t.Fatal("missing required grouped field must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"reason": "出差"}, false); err != nil {
		t.Fatalf("valid grouped form data rejected: %v", err)
	}
	if err := ValidateFormData(fields, map[string]interface{}{"tip": "篡改"}, false); err == nil {
		t.Fatal("display component key must not be accepted as form data")
	}
}

func TestValidateNodeFormPatchSupportsGroupedFields(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{{Key: "group", Label: "分组", Type: FormFieldTypeGroup, Fields: []FormField{
		{Key: "reason", Label: "原因", Type: FormFieldTypeText, Required: true},
	}}}
	definition.Nodes[1].FormPermissions = []FieldPermission{{Field: "reason", Access: FieldAccessWrite}}

	if err := ValidateNodeFormPatch(definition, "manager", map[string]interface{}{"reason": "出差"}, map[string]interface{}{"reason": "外出"}); err != nil {
		t.Fatalf("grouped writable field patch rejected: %v", err)
	}
}

func TestValidateFormDataSupportsTreeOptionsAndRemoteOptionFields(t *testing.T) {
	fields := []FormField{
		{
			Key:   "region",
			Label: "所属区域",
			Type:  FormFieldTypeSelect,
			Options: []FormOption{{Label: "华东", Value: "east", Children: []FormOption{
				{Label: "上海", Value: "shanghai"},
				{Label: "杭州", Value: "hangzhou"},
			}}},
		},
		{
			Key:   "department",
			Label: "所属部门",
			Type:  FormFieldTypeMultiSelect,
			OptionSource: &FormOptionSource{
				Type:          OptionSourceAPI,
				URL:           "/api/v2/admin/departments/tree",
				Method:        "GET",
				ResponsePath:  "data",
				LabelField:    "name",
				ValueField:    "id",
				ChildrenField: "children",
			},
		},
	}

	if err := ValidateFormData(fields, map[string]interface{}{"region": "shanghai", "department": []interface{}{"dept-1", "dept-2"}}, false); err != nil {
		t.Fatalf("valid tree and remote option form data rejected: %v", err)
	}
	if err := ValidateFormData(fields, map[string]interface{}{"region": "beijing", "department": []interface{}{"dept-1"}}, false); err == nil {
		t.Fatal("static tree select value outside configured options must be rejected")
	}
}

func TestValidateFormDataSupportsDetailListFields(t *testing.T) {
	fields := []FormField{objectiveDetailListField()}
	valid := map[string]interface{}{
		"objectives": []interface{}{
			map[string]interface{}{"id": "obj-1", "target": "提升续费率", "weight": 40, "result": "完成核心客户拜访"},
			map[string]interface{}{"id": "obj-2", "target": "优化交付质量", "weight": 60, "result": "沉淀验收清单"},
		},
	}

	if err := ValidateFormData(fields, valid, false); err != nil {
		t.Fatalf("valid detail list data rejected: %v", err)
	}
	if err := ValidateFormData(fields, map[string]interface{}{"objectives": []interface{}{}}, true); err == nil {
		t.Fatal("submitted detail list below min rows must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"objectives": []interface{}{map[string]interface{}{"id": "obj-1", "weight": 40}}}, false); err == nil {
		t.Fatal("detail row missing required column must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"objectives": []interface{}{map[string]interface{}{"id": "obj-1", "target": "A", "unknown": "B"}}}, false); err == nil {
		t.Fatal("detail row with unknown column must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"objectives": []interface{}{map[string]interface{}{"target": "A"}}}, false); err == nil {
		t.Fatal("detail row without row key must be rejected")
	}
}

func TestValidateFormDataEnforcesRequiredTypeAndRange(t *testing.T) {
	fields := []FormField{
		{Key: "reason", Label: "原因", Type: FormFieldTypeText, Required: true},
		{Key: "days", Label: "天数", Type: FormFieldTypeNumber, Min: numberPointer(1), Max: numberPointer(30)},
	}

	if err := ValidateFormData(fields, map[string]interface{}{"reason": "年假", "days": 2}, false); err != nil {
		t.Fatalf("valid form data rejected: %v", err)
	}
	if err := ValidateFormData(fields, map[string]interface{}{"days": 2}, false); err == nil {
		t.Fatal("missing required field must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"reason": "年假", "days": 31}, false); err == nil {
		t.Fatal("out-of-range number must be rejected")
	}
	if err := ValidateFormData(fields, map[string]interface{}{"reason": "年假", "unknown": true}, false); err == nil {
		t.Fatal("unknown form field must be rejected")
	}
}

func TestValidateFormDataSupportsCommonOAFields(t *testing.T) {
	fields := []FormField{
		{Key: "amount", Label: "报销金额", Type: FormFieldTypeAmount, Min: numberPointer(0), Max: numberPointer(100000)},
		{Key: "phone", Label: "联系电话", Type: FormFieldTypePhone},
		{Key: "email", Label: "邮箱", Type: FormFieldTypeEmail},
		{Key: "method", Label: "出行方式", Type: FormFieldTypeRadio, Options: []FormOption{{Label: "飞机", Value: "plane"}}},
		{Key: "benefits", Label: "福利", Type: FormFieldTypeCheckbox, Options: []FormOption{{Label: "餐补", Value: "meal"}, {Label: "交通", Value: "traffic"}}},
		{Key: "startTime", Label: "开始时间", Type: FormFieldTypeTime},
		{Key: "tripPeriod", Label: "出差日期", Type: FormFieldTypeDateRange},
		{Key: "companions", Label: "同行人", Type: FormFieldTypeUserMulti},
		{Key: "departments", Label: "协同部门", Type: FormFieldTypeDepartmentMulti},
	}
	definition := validLinearDefinition()
	definition.Form = fields
	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("common OA fields must be accepted in form schema: %#v", errors)
	}
	data := map[string]interface{}{
		"amount":      1280.5,
		"phone":       "13800138000",
		"email":       "oa@example.com",
		"method":      "plane",
		"benefits":    []interface{}{"meal", "traffic"},
		"startTime":   "09:30",
		"tripPeriod":  []interface{}{"2026-09-01", "2026-09-03"},
		"companions":  []interface{}{"1001", "1002"},
		"departments": []interface{}{"2001", "2002"},
	}

	if err := ValidateFormData(fields, data, false); err != nil {
		t.Fatalf("valid common OA form data rejected: %v", err)
	}
	for name, value := range map[string]interface{}{
		"phone":      "123",
		"email":      "invalid-email",
		"method":     "train",
		"benefits":   []interface{}{"unknown"},
		"startTime":  "25:00",
		"tripPeriod": []interface{}{"2026-09-01"},
	} {
		invalid := make(map[string]interface{}, len(data))
		for key, item := range data {
			invalid[key] = item
		}
		invalid[name] = value
		if err := ValidateFormData(fields, invalid, false); err == nil {
			t.Fatalf("invalid %s value must be rejected", name)
		}
	}
}

func TestValidateNodeFormPatchOnlyAllowsWritableFields(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "reason", Label: "原因", Type: FormFieldTypeText},
		{Key: "amount", Label: "核定金额", Type: FormFieldTypeNumber},
	}
	definition.Nodes[1].FormPermissions = []FieldPermission{
		{Field: "reason", Access: FieldAccessRead},
		{Field: "amount", Access: FieldAccessWrite},
	}

	if err := ValidateNodeFormPatch(definition, "manager", nil, map[string]interface{}{"amount": 120}); err != nil {
		t.Fatalf("writable field patch rejected: %v", err)
	}
	if err := ValidateNodeFormPatch(definition, "manager", nil, map[string]interface{}{"reason": "changed"}); err == nil {
		t.Fatal("read-only field patch must be rejected")
	}
}

func TestValidateNodeFormPatchRejectsInvalidMergedFormData(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "reason", Label: "原因", Type: FormFieldTypeText, Required: true},
	}
	definition.Nodes[1].FormPermissions = []FieldPermission{
		{Field: "reason", Access: FieldAccessWrite},
	}
	current := map[string]interface{}{"reason": "出差"}

	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"reason": ""}); err == nil {
		t.Fatal("required field must not be cleared by node form patch")
	}
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"reason": "调整为外出"}); err != nil {
		t.Fatalf("valid merged form patch rejected: %v", err)
	}
}

func TestValidateNodeFormPatchEnforcesDetailListRowActions(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{objectiveDetailListField()}
	definition.Nodes[1].FormPermissions = []FieldPermission{
		{Field: "objectives", Access: FieldAccessWrite},
	}
	current := map[string]interface{}{
		"objectives": []interface{}{
			map[string]interface{}{"id": "obj-1", "target": "提升续费率", "weight": 40, "result": "待跟进"},
			map[string]interface{}{"id": "obj-2", "target": "优化交付质量", "weight": 60, "result": "待验收"},
		},
	}

	editRows := []interface{}{
		map[string]interface{}{"id": "obj-1", "target": "提升续费率", "weight": 40, "result": "已拜访"},
		map[string]interface{}{"id": "obj-2", "target": "优化交付质量", "weight": 60, "result": "待验收"},
	}
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"objectives": editRows}); err != nil {
		t.Fatalf("editing existing detail rows should be allowed by write access: %v", err)
	}

	addRows := append(append([]interface{}{}, editRows...), map[string]interface{}{"id": "obj-3", "target": "新增专项", "weight": 10, "result": ""})
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"objectives": addRows}); err == nil {
		t.Fatal("adding detail rows without add action must be rejected")
	}

	definition.Nodes[1].FormPermissions[0].Actions = []string{FieldActionAdd}
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"objectives": addRows}); err != nil {
		t.Fatalf("adding detail rows with add action should be allowed: %v", err)
	}
	deleteRows := []interface{}{editRows[0]}
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"objectives": deleteRows}); err == nil {
		t.Fatal("deleting detail rows without delete action must be rejected")
	}

	definition.Nodes[1].FormPermissions[0].Actions = []string{FieldActionAdd, FieldActionDelete}
	if err := ValidateNodeFormPatch(definition, "manager", current, map[string]interface{}{"objectives": deleteRows}); err != nil {
		t.Fatalf("deleting detail rows with delete action should be allowed: %v", err)
	}
}

func numberPointer(value float64) *float64 { return &value }

func extendedNodeDefinition(t *testing.T) Definition {
	t.Helper()
	const source = `{
		"schemaVersion": 1,
		"key": "extended_nodes",
		"name": "扩展节点流程",
		"form": [{"key":"content","label":"内容","type":"text"}],
		"nodes": [
			{"id":"start","type":"start","name":"开始"},
			{"id":"handle","type":"handle","name":"填写资料","assignee":{"type":"variable","value":"targetUserId"},"formPermissions":[{"field":"content","access":"write"}]},
			{"id":"cc","type":"cc","name":"通知相关人","assignee":{"type":"variable","value":"ccUserIds"}},
			{"id":"automation","type":"automation","name":"写入变量","automation":{"type":"set_variables","variables":{"processed":true}}},
			{"id":"timer","type":"timer","name":"等待 30 秒","timer":{"delaySeconds":30}},
			{"id":"end","type":"end","name":"结束"}
		],
		"edges": [
			{"id":"e1","source":"start","target":"handle"},
			{"id":"e2","source":"handle","target":"cc"},
			{"id":"e3","source":"cc","target":"automation"},
			{"id":"e4","source":"automation","target":"timer"},
			{"id":"e5","source":"timer","target":"end"}
		]
	}`
	var definition Definition
	if err := json.Unmarshal([]byte(source), &definition); err != nil {
		t.Fatalf("decode extended definition: %v", err)
	}
	return definition
}

func objectiveDetailListField() FormField {
	return FormField{
		Key:     "objectives",
		Label:   "我的目标",
		Type:    FormFieldTypeDetailList,
		RowKey:  "id",
		MinRows: 1,
		MaxRows: 20,
		Span:    24,
		Columns: []FormField{
			{Key: "target", Label: "目标", Type: FormFieldTypeTextarea, Required: true, MaxLength: 200},
			{Key: "weight", Label: "权重", Type: FormFieldTypeNumber, Min: numberPointer(0), Max: numberPointer(100)},
			{Key: "result", Label: "结果", Type: FormFieldTypeTextarea, MaxLength: 500},
		},
	}
}

func validLinearDefinition() Definition {
	return Definition{
		SchemaVersion: 1,
		Key:           "leave_approval",
		Name:          "请假审批",
		Nodes: []Node{
			{ID: "start", Type: NodeTypeStart, Name: "开始"},
			{
				ID:           "manager",
				Type:         NodeTypeApproval,
				Name:         "直属上级审批",
				ApprovalMode: ApprovalModeSingle,
				Assignee:     &Assignee{Type: AssigneeTypeManager, Value: "direct"},
			},
			{ID: "end", Type: NodeTypeEnd, Name: "结束"},
		},
		Edges: []Edge{
			{ID: "e1", Source: "start", Target: "manager"},
			{ID: "e2", Source: "manager", Target: "end"},
		},
	}
}

func hasValidationCode(errors []ValidationError, code string) bool {
	for _, item := range errors {
		if item.Code == code {
			return true
		}
	}
	return false
}
