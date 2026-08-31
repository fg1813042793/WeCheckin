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
