package workflowcore

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompletedRevisionCapabilityUsesOneNode(t *testing.T) {
	definition := completedRevisionDefinition()

	capability, err := CompletedRevisionCapabilityForNode(definition, "manager")
	if err != nil {
		t.Fatalf("CompletedRevisionCapabilityForNode() error = %v", err)
	}
	if capability.NodeID != "manager" || capability.NodeName != "主管审批" {
		t.Fatalf("capability source = %#v", capability)
	}
	if !reflect.DeepEqual(capability.DirectFields, []string{"remark"}) {
		t.Fatalf("direct fields = %#v", capability.DirectFields)
	}
	if accessForCompletedRevisionField(capability.FieldPermissions, "amount") != FieldAccessRead {
		t.Fatal("routing field must be read-only")
	}
	if accessForCompletedRevisionField(capability.FieldPermissions, "total") != FieldAccessRead {
		t.Fatal("calculation field must be read-only")
	}
	if accessForCompletedRevisionField(capability.FieldPermissions, "hrComment") != FieldAccessRead {
		t.Fatal("another node's writable field must not be merged")
	}
	if accessForCompletedRevisionField(capability.FieldPermissions, "remark") != FieldAccessWrite {
		t.Fatal("source node writable field must remain writable")
	}
}

func TestValidateCompletedRevisionPatchRejectsRoutingAndCalculationFields(t *testing.T) {
	definition := completedRevisionDefinition()
	current := map[string]interface{}{"amount": float64(100), "remark": "原说明", "total": float64(100)}

	for _, patch := range []map[string]interface{}{
		{"amount": float64(200)},
		{"total": float64(200)},
		{"hrComment": "跨节点修改"},
	} {
		if err := ValidateCompletedRevisionPatch(definition, "manager", current, patch); err == nil {
			t.Fatalf("patch %#v must be rejected", patch)
		}
	}
	if err := ValidateCompletedRevisionPatch(definition, "manager", current, map[string]interface{}{"remark": "修正说明"}); err != nil {
		t.Fatalf("direct field patch error = %v", err)
	}
}

func TestBuildCompletedRevisionPlanUsesActualDownstreamPath(t *testing.T) {
	definition := completedRevisionDefinition()

	stages, err := BuildCompletedRevisionPlan(
		definition,
		[]string{"manager", "finance", "legal", "hr"},
		"manager",
	)
	if err != nil {
		t.Fatalf("BuildCompletedRevisionPlan() error = %v", err)
	}
	want := []RevisionPlanStage{
		{Stage: 1, Nodes: []string{"finance", "legal"}},
		{Stage: 2, Nodes: []string{"hr"}},
	}
	if !reflect.DeepEqual(stages, want) {
		t.Fatalf("stages = %#v, want %#v", stages, want)
	}
}

func TestCompletedRevisionModeRequiresReviewForAnyNonDirectField(t *testing.T) {
	capability := CompletedRevisionCapability{DirectFields: []string{"remark", "phone"}}
	if mode := CompletedRevisionMode(capability, map[string]interface{}{"remark": "修正"}); mode != CompletedRevisionModeDirect {
		t.Fatalf("direct mode = %s", mode)
	}
	if mode := CompletedRevisionMode(capability, map[string]interface{}{"remark": "修正", "amount": 200}); mode != CompletedRevisionModeDownstreamReview {
		t.Fatalf("review mode = %s", mode)
	}
}

func TestValidateDefinitionRejectsInvalidCompletedRevisionDirectFields(t *testing.T) {
	definition := completedRevisionDefinition()
	definition.Nodes[1].PostHandleEdit.CompletedRevision.DirectFields = []string{"amount", "total", "unknown"}

	errors := ValidateDefinition(definition)
	count := 0
	for _, validationError := range errors {
		if validationError.Code == ValidationCompletedRevision && validationError.NodeID == "manager" {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("completed revision validation errors = %d, all errors = %#v", count, errors)
	}
}

func TestCompileBPMNIncludesCompletedRevisionAttributes(t *testing.T) {
	definition := completedRevisionDefinition()
	bpmn, err := CompileBPMN(definition)
	if err != nil {
		t.Fatalf("CompileBPMN() error = %v", err)
	}

	for _, token := range []string{
		`flowable:postHandleEditEnabled="true"`,
		`flowable:completedRevisionEnabled="true"`,
		`flowable:completedRevisionDirectFields="remark"`,
	} {
		if !strings.Contains(string(bpmn), token) {
			t.Fatalf("BPMN missing %s: %s", token, bpmn)
		}
	}
}

func completedRevisionDefinition() Definition {
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "amount", Label: "金额", Type: FormFieldTypeNumber},
		{Key: "remark", Label: "说明", Type: FormFieldTypeTextarea},
		{Key: "hrComment", Label: "HR 意见", Type: FormFieldTypeTextarea},
		{Key: "total", Label: "合计", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "[amount]"}},
	}
	definition.Nodes = []Node{
		{ID: "start", Type: NodeTypeStart, Name: "开始", Initiator: &InitiatorConfig{Scope: InitiatorScopeAll}},
		{
			ID: "manager", Type: NodeTypeApproval, Name: "主管审批",
			ApprovalMode: ApprovalModeSingle,
			Assignee:     &Assignee{Type: AssigneeTypeManager, Value: "direct_manager"},
			FormPermissions: []FieldPermission{
				{Field: "amount", Access: FieldAccessWrite},
				{Field: "remark", Access: FieldAccessWrite},
				{Field: "hrComment", Access: FieldAccessRead},
				{Field: "total", Access: FieldAccessRead},
			},
			PostHandleEdit: &PostHandleEditConfig{
				Enabled: true,
				CompletedRevision: &CompletedRevisionConfig{
					Enabled:      true,
					DirectFields: []string{"remark"},
				},
			},
		},
		{ID: "route", Type: NodeTypeExclusive, Name: "金额判断", GatewayMode: GatewayModeSplit},
		{ID: "parallel-split", Type: NodeTypeParallel, Name: "并行开始", GatewayMode: GatewayModeSplit},
		{
			ID: "finance", Type: NodeTypeApproval, Name: "财务审批",
			ApprovalMode: ApprovalModeSingle,
			Assignee:     &Assignee{Type: AssigneeTypeUser, Value: "31"},
		},
		{
			ID: "legal", Type: NodeTypeApproval, Name: "法务审批",
			ApprovalMode: ApprovalModeSingle,
			Assignee:     &Assignee{Type: AssigneeTypeUser, Value: "32"},
		},
		{ID: "parallel-join", Type: NodeTypeParallel, Name: "并行汇合", GatewayMode: GatewayModeJoin},
		{
			ID: "hr", Type: NodeTypeApproval, Name: "HR 审批",
			ApprovalMode: ApprovalModeSingle,
			Assignee:     &Assignee{Type: AssigneeTypeUser, Value: "33"},
			FormPermissions: []FieldPermission{
				{Field: "hrComment", Access: FieldAccessWrite},
			},
		},
		{ID: "audit", Type: NodeTypeApproval, Name: "审计审批", ApprovalMode: ApprovalModeSingle, Assignee: &Assignee{Type: AssigneeTypeUser, Value: "34"}},
		{ID: "end", Type: NodeTypeEnd, Name: "结束"},
	}
	definition.Edges = []Edge{
		{ID: "e1", Source: "start", Target: "manager"},
		{ID: "e2", Source: "manager", Target: "route"},
		{ID: "e3", Source: "route", Target: "parallel-split", Condition: &Condition{Field: "amount", Operator: ConditionGTE, Value: float64(0)}},
		{ID: "e4", Source: "parallel-split", Target: "finance"},
		{ID: "e5", Source: "parallel-split", Target: "legal"},
		{ID: "e6", Source: "finance", Target: "parallel-join"},
		{ID: "e7", Source: "legal", Target: "parallel-join"},
		{ID: "e8", Source: "parallel-join", Target: "hr"},
		{ID: "e9", Source: "hr", Target: "end"},
		{ID: "e10", Source: "route", Target: "audit", Default: true},
		{ID: "e11", Source: "audit", Target: "end"},
	}
	return definition
}

func accessForCompletedRevisionField(permissions []FieldPermission, field string) string {
	for _, permission := range permissions {
		if permission.Field == field {
			return permission.Access
		}
	}
	return ""
}
