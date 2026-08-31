package workflow

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestValidateDefinitionAcceptsLinearApproval(t *testing.T) {
	definition := validLinearDefinition()

	errors := ValidateDefinition(definition)
	if len(errors) != 0 {
		t.Fatalf("expected valid definition, got errors: %#v", errors)
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
			name: "select without options",
			mutate: func(definition *Definition) {
				definition.Form = []FormField{{Key: "type", Label: "类型", Type: FormFieldTypeSelect}}
			},
			expectCode: ValidationFormFieldOptions,
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

	if err := ValidateNodeFormPatch(definition, "manager", map[string]interface{}{"amount": 120}); err != nil {
		t.Fatalf("writable field patch rejected: %v", err)
	}
	if err := ValidateNodeFormPatch(definition, "manager", map[string]interface{}{"reason": "changed"}); err == nil {
		t.Fatal("read-only field patch must be rejected")
	}
}

func numberPointer(value float64) *float64 { return &value }

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
