package infrastructure

import (
	"reflect"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	workflowcore "wecheckin/backend/internal/workflow"
)

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
			actual, err := resolver.Resolve(workflowdomain.AssigneeRequest{
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
	actual, err := resolver.Resolve(workflowdomain.AssigneeRequest{
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
	actual, err := resolver.Resolve(workflowdomain.AssigneeRequest{
		Node: workflowcore.Node{Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "9, 4, 9, 2"}},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !reflect.DeepEqual(actual, []string{"9", "4", "2"}) {
		t.Fatalf("Resolve() = %#v, want [9 4 2]", actual)
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
