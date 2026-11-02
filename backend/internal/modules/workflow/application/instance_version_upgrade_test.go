package application

import (
	"context"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestAnalyzeInstanceVersionUpgradeAllowsStableRuntimeReferences(t *testing.T) {
	source, target, state := compatibleUpgradeFixture()
	preview := analyzeInstanceVersionUpgrade(source, target, state, nil, 2)
	if !preview.Compatible || preview.CurrentVersion != 1 || preview.TargetVersion != 2 || len(preview.Blockers) != 0 {
		t.Fatalf("preview = %#v", preview)
	}
}

func TestAnalyzeInstanceVersionUpgradeBlocksUnsafeChanges(t *testing.T) {
	tests := []struct {
		name string
		edit func(*workflowcore.Definition, *workflowdomain.State) *workflowdomain.FormRevisionRequest
		code string
	}{
		{name: "terminal", edit: func(_ *workflowcore.Definition, state *workflowdomain.State) *workflowdomain.FormRevisionRequest {
			state.Instance.Status = workflowdomain.InstanceStatusCompleted
			return nil
		}, code: "instance_not_running"},
		{name: "missing active node", edit: func(target *workflowcore.Definition, _ *workflowdomain.State) *workflowdomain.FormRevisionRequest {
			target.Nodes = target.Nodes[:1]
			return nil
		}, code: "active_node_missing"},
		{name: "node type changed", edit: func(target *workflowcore.Definition, _ *workflowdomain.State) *workflowdomain.FormRevisionRequest {
			target.Nodes[1].Type = workflowcore.NodeTypeHandle
			return nil
		}, code: "active_node_type_changed"},
		{name: "form field removed", edit: func(target *workflowcore.Definition, _ *workflowdomain.State) *workflowdomain.FormRevisionRequest {
			target.Form = nil
			return nil
		}, code: "form_field_missing"},
		{name: "active revision", edit: func(_ *workflowcore.Definition, _ *workflowdomain.State) *workflowdomain.FormRevisionRequest {
			return &workflowdomain.FormRevisionRequest{ID: "revision-1", Status: workflowdomain.FormRevisionStatusPending}
		}, code: "active_form_revision"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, target, state := compatibleUpgradeFixture()
			activeRevision := test.edit(&target, state)
			preview := analyzeInstanceVersionUpgrade(source, target, state, activeRevision, 2)
			if preview.Compatible || !hasUpgradeBlocker(preview.Blockers, test.code) {
				t.Fatalf("preview = %#v", preview)
			}
		})
	}
}

func TestUpgradeInstanceVersionRechecksAndAudits(t *testing.T) {
	source, target, state := compatibleUpgradeFixture()
	store := &fakeStore{definition: source, targetDefinition: target, targetVersion: 2, state: state}
	service := NewService(store, fixedResolver{"42"}, &sequenceIDs{})

	result, err := service.UpgradeInstanceVersion(context.Background(), UpgradeInstanceVersionRequest{
		InstanceID: "instance-1", ActorID: "99", ExpectedFromVersion: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentVersion != 2 || store.updatedDefinitionVersionFrom != 1 || store.updatedDefinitionVersionTo != 2 || !store.updatedVersionInTransaction {
		t.Fatalf("result/store = %#v / %#v", result, store)
	}
	if store.appendedHistory.Type != workflowdomain.HistoryInstanceVersionUpgraded || store.appendedHistory.ActorID != "99" {
		t.Fatalf("history = %#v", store.appendedHistory)
	}
}

func compatibleUpgradeFixture() (workflowcore.Definition, workflowcore.Definition, *workflowdomain.State) {
	source := workflowcore.Definition{
		Key:   "performance",
		Form:  []workflowcore.FormField{{Key: "grade", Type: workflowcore.FormFieldTypeText}},
		Nodes: []workflowcore.Node{{ID: "start", Type: workflowcore.NodeTypeStart}, {ID: "approve", Type: workflowcore.NodeTypeApproval}},
	}
	target := source
	target.Nodes = append([]workflowcore.Node(nil), source.Nodes...)
	target.Form = append([]workflowcore.FormField(nil), source.Form...)
	state := &workflowdomain.State{
		Instance: workflowdomain.ProcessInstance{ID: "instance-1", DefinitionID: 7, DefinitionVersion: 1, Status: workflowdomain.InstanceStatusRunning},
		Tokens:   []workflowdomain.Token{{ID: "token-1", NodeID: "approve", Status: workflowdomain.TokenStatusWaiting}},
		Tasks:    []workflowdomain.Task{{ID: "task-1", NodeID: "approve", Status: workflowdomain.TaskStatusPending}},
		FormData: map[string]interface{}{"grade": "C"},
	}
	return source, target, state
}

func hasUpgradeBlocker(blockers []InstanceVersionUpgradeBlocker, code string) bool {
	for _, blocker := range blockers {
		if blocker.Code == code {
			return true
		}
	}
	return false
}
