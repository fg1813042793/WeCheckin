package domain

import (
	"context"
	"fmt"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

type layeredResolver struct {
	layers []ApprovalLayer
	calls  int
}

func (resolver *layeredResolver) Resolve(_ context.Context, _ AssigneeRequest) ([]string, error) {
	return nil, fmt.Errorf("flat resolver must not be used for layered approval")
}

func (resolver *layeredResolver) ResolveApprovalLayers(_ context.Context, _ AssigneeRequest) ([]ApprovalLayer, error) {
	resolver.calls++
	return resolver.layers, nil
}

func TestEngineLayeredApprovalActivatesOneDepartmentAtATime(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeParallel, 0)
	definition.Nodes[1].Assignee = &workflowcore.Assignee{
		Type: workflowcore.AssigneeTypeOrgIdentity, Value: "starter_department:department_leader",
	}
	definition.Nodes[1].DepartmentApprovalChain = &workflowcore.DepartmentApprovalChainConfig{
		Enabled: true, StopMode: workflowcore.DepartmentApprovalChainStopRoot,
		MissingAssigneePolicy: workflowcore.DepartmentApprovalChainMissingSkip,
	}
	resolver := &layeredResolver{layers: []ApprovalLayer{
		{DepartmentID: 30, DepartmentName: "产品组", AssigneeIDs: []string{"u1", "u2"}},
		{DepartmentID: 20, DepartmentName: "产品部", AssigneeIDs: []string{"u3"}},
	}}
	engine := newTestEngine(resolver)

	state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start layered workflow: %v", err)
	}
	if resolver.calls != 1 {
		t.Fatalf("layer resolver calls = %d, want 1", resolver.calls)
	}
	if len(state.Tasks) != 3 || len(state.PendingTasks()) != 2 {
		t.Fatalf("initial layered tasks = %#v", state.Tasks)
	}
	if state.Tasks[0].ApprovalChainKey == "" || state.Tasks[0].ApprovalChainKey != state.Tasks[2].ApprovalChainKey {
		t.Fatalf("approval chain keys = %#v", state.Tasks)
	}
	if state.Tasks[0].ApprovalLayer != 1 || state.Tasks[0].ApprovalLayerTotal != 2 || state.Tasks[0].DepartmentID != 30 || state.Tasks[0].DepartmentName != "产品组" {
		t.Fatalf("first layer task metadata = %#v", state.Tasks[0])
	}
	if state.Tasks[2].Status != TaskStatusWaiting || state.Tasks[2].ApprovalLayer != 2 || state.Tasks[2].DepartmentID != 20 {
		t.Fatalf("second layer task = %#v", state.Tasks[2])
	}

	for _, task := range append([]Task(nil), state.PendingTasks()...) {
		if err := engine.Complete(context.Background(), definition, state, CompleteRequest{
			TaskID: task.ID, ActorID: task.AssigneeID, Action: TaskActionApprove,
		}); err != nil {
			t.Fatalf("complete first layer task: %v", err)
		}
	}
	pending := state.PendingTasks()
	if len(pending) != 1 || pending[0].AssigneeID != "u3" || pending[0].ApprovalLayer != 2 {
		t.Fatalf("activated second layer = %#v", pending)
	}
	if err := engine.Complete(context.Background(), definition, state, CompleteRequest{
		TaskID: pending[0].ID, ActorID: "u3", Action: TaskActionApprove,
	}); err != nil {
		t.Fatalf("complete second layer: %v", err)
	}
	if state.Instance.Status != InstanceStatusCompleted {
		t.Fatalf("instance status = %s, want completed", state.Instance.Status)
	}
}
