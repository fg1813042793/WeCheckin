package infrastructure

import (
	"encoding/json"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/workflowcore"
)

func TestPublishedDefinitionIncludesLogoURL(t *testing.T) {
	result := publishedDefinition(
		workflowmodel.Definition{LogoURL: "https://static.example.com/uploads/workflow-logos/performance.png"},
		workflowcore.Definition{},
		5,
		publishedAssigneeLabels{},
		false,
	)

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal published definition: %v", err)
	}
	if !strings.Contains(string(encoded), `"logoUrl":"https://static.example.com/uploads/workflow-logos/performance.png"`) {
		t.Fatalf("published definition response missing logoUrl: %s", encoded)
	}
}

func TestBuildPublishedWorkflowGraphDisplaysReadableAssignees(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "user-approval", Type: workflowcore.NodeTypeApproval, Name: "指定人员审批", Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: "7,9"}},
			{ID: "manager-approval", Type: workflowcore.NodeTypeApproval, Name: "上级审批", Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager, Value: "direct_manager"}},
			{ID: "identity-approval", Type: workflowcore.NodeTypeApproval, Name: "部门负责人审批", Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeOrgIdentity, Value: "department:3:department_leader"}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{{ID: "flow-1", Source: "start", Target: "user-approval", Name: "提交"}},
	}

	nodes, edges := buildPublishedWorkflowGraph(definition, publishedAssigneeLabels{
		users:       map[uint]string{7: "张三", 9: "李四"},
		identities:  map[string]string{"department_leader": "部门负责人"},
		departments: map[uint]string{3: "产品部"},
	})

	if len(nodes) != 5 || len(edges) != 1 {
		t.Fatalf("graph size = (%d nodes, %d edges), want (5, 1)", len(nodes), len(edges))
	}
	if nodes[1].AssigneeDisplay != "张三、李四" {
		t.Fatalf("user assignee display = %q", nodes[1].AssigneeDisplay)
	}
	if nodes[2].AssigneeDisplay != "发起人的直属上级" {
		t.Fatalf("manager assignee display = %q", nodes[2].AssigneeDisplay)
	}
	if nodes[3].AssigneeDisplay != "产品部 · 部门负责人" {
		t.Fatalf("org identity display = %q", nodes[3].AssigneeDisplay)
	}
	if edges[0].Source != "start" || edges[0].Target != "user-approval" || edges[0].Name != "提交" {
		t.Fatalf("published edge = %#v", edges[0])
	}
}
