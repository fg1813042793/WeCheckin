package infrastructure

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestResolveReturnsCanceledContextBeforeResolvingAssignees(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resolver := NewAssigneeResolver(nil)

	actual, err := resolver.Resolve(ctx, workflowdomain.AssigneeRequest{
		Node: workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "9"}},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Resolve() error = %v, want context.Canceled", err)
	}
	if actual != nil {
		t.Fatalf("Resolve() = %#v, want nil", actual)
	}
}

func TestResolvedAssigneeDisplayNamesPreserveResolverOrder(t *testing.T) {
	actual := resolvedAssigneeDisplayNames(
		[]string{"9", "7", "9", "invalid"},
		[]model.User{
			{ID: 7, Name: "主管张三", Account: "zhangsan"},
			{ID: 9, Name: "主管李四", Account: "lisi"},
		},
	)
	if !reflect.DeepEqual(actual, []string{"主管李四", "主管张三"}) {
		t.Fatalf("resolved display names = %#v", actual)
	}
}

func TestResolveVariableAssigneesSupportsCommonValueShapes(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]interface{}
		want      []string
	}{
		{name: "comma string", variables: map[string]interface{}{"reviewers": "7, 8,7"}, want: []string{"7", "8"}},
		{name: "string slice", variables: map[string]interface{}{"reviewers": []string{"8", "9"}}, want: []string{"8", "9"}},
		{name: "interface slice", variables: map[string]interface{}{"reviewers": []interface{}{"10", float64(11)}}, want: []string{"10", "11"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewAssigneeResolver(nil)
			actual, err := resolver.Resolve(context.Background(), workflowdomain.AssigneeRequest{
				Node:      workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeVariable, Value: "reviewers"}},
				Variables: test.variables,
			})
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if !reflect.DeepEqual(actual, test.want) {
				t.Fatalf("Resolve() = %#v, want %#v", actual, test.want)
			}
		})
	}
}

func TestResolveManagerUsesGenericWorkflowVariable(t *testing.T) {
	resolver := NewAssigneeResolver(nil)
	actual, err := resolver.Resolve(context.Background(), workflowdomain.AssigneeRequest{
		Node:      workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager}},
		Variables: map[string]interface{}{"managerId": "21"},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !reflect.DeepEqual(actual, []string{"21"}) {
		t.Fatalf("Resolve() = %#v, want [21]", actual)
	}
}

func TestResolveUserAssigneePreservesConfiguredOrder(t *testing.T) {
	resolver := NewAssigneeResolver(nil)
	actual, err := resolver.Resolve(context.Background(), workflowdomain.AssigneeRequest{
		Node: workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "9, 4, 9, 2"}},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !reflect.DeepEqual(actual, []string{"9", "4", "2"}) {
		t.Fatalf("Resolve() = %#v, want [9 4 2]", actual)
	}
}

func TestResolveInitiatorAssigneeUsesBusinessStarter(t *testing.T) {
	resolver := NewAssigneeResolver(nil)
	actual, err := resolver.Resolve(context.Background(), workflowdomain.AssigneeRequest{
		Instance: workflowdomain.ProcessInstance{StarterID: "27", OperatorID: "3"},
		Node:     workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeInitiator}},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !reflect.DeepEqual(actual, []string{"27"}) {
		t.Fatalf("Resolve() = %#v, want business starter [27]", actual)
	}
}

func TestParseOrgIdentityValueScopes(t *testing.T) {
	tests := []struct {
		raw            string
		wantScope      string
		wantDepartment uint
		wantIdentity   string
	}{
		{raw: "department_leader", wantScope: "starter_department", wantIdentity: "department_leader"},
		{raw: "starter_department:group_leader", wantScope: "starter_department", wantIdentity: "group_leader"},
		{raw: "department:17:hrbp", wantScope: "department", wantDepartment: 17, wantIdentity: "hrbp"},
	}
	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			scope, departmentID, identityCode := parseOrgIdentityValue(test.raw)
			if scope != test.wantScope || departmentID != test.wantDepartment || identityCode != test.wantIdentity {
				t.Fatalf("parseOrgIdentityValue() = (%q, %d, %q), want (%q, %d, %q)",
					scope, departmentID, identityCode, test.wantScope, test.wantDepartment, test.wantIdentity)
			}
		})
	}
}

func TestBuildDepartmentApprovalPathUsesDynamicAncestors(t *testing.T) {
	departments := []model.Department{
		{ID: 1, Name: "集团", ParentID: 0},
		{ID: 2, Name: "事业部", ParentID: 1},
		{ID: 3, Name: "产品部", ParentID: 2},
	}
	path, err := buildDepartmentApprovalPath(departments, 3, workflowcore.DepartmentApprovalChainStopRoot, 0)
	if err != nil {
		t.Fatalf("build root approval path: %v", err)
	}
	if got := departmentPathIDs(path); !reflect.DeepEqual(got, []uint{3, 2, 1}) {
		t.Fatalf("root approval path = %#v", got)
	}

	path, err = buildDepartmentApprovalPath(departments, 3, workflowcore.DepartmentApprovalChainStopDepartment, 2)
	if err != nil {
		t.Fatalf("build bounded approval path: %v", err)
	}
	if got := departmentPathIDs(path); !reflect.DeepEqual(got, []uint{3, 2}) {
		t.Fatalf("bounded approval path = %#v", got)
	}

	if _, err := buildDepartmentApprovalPath(departments, 3, workflowcore.DepartmentApprovalChainStopDepartment, 99); err == nil {
		t.Fatal("non-ancestor stop department should fail")
	}
}

func TestBuildDepartmentApprovalPathRejectsCycle(t *testing.T) {
	departments := []model.Department{{ID: 1, ParentID: 2}, {ID: 2, ParentID: 1}}
	if _, err := buildDepartmentApprovalPath(departments, 1, workflowcore.DepartmentApprovalChainStopRoot, 0); err == nil {
		t.Fatal("cyclic department path should fail")
	}
}

func TestFilterApprovalLayerAssigneesSkipsStarterAndCrossLayerDuplicates(t *testing.T) {
	seen := map[string]struct{}{"9": {}}
	actual := filterApprovalLayerAssignees([]string{"7", "8", "9", "8"}, "7", true, seen)
	if !reflect.DeepEqual(actual, []string{"8"}) {
		t.Fatalf("filtered layer assignees = %#v", actual)
	}
	if _, exists := seen["8"]; !exists {
		t.Fatal("new layer assignee should be recorded as seen")
	}
}
