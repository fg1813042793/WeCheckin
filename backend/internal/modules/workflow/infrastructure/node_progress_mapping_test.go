package infrastructure

import (
	"testing"

	"wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/workflowcore"
)

func TestPopulateInstanceNodeProgressUsesLoadedDefinition(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start-v2", Type: workflowcore.NodeTypeStart},
			{ID: "approve-v2", Type: workflowcore.NodeTypeApproval},
			{ID: "end-v2", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start-v2", Target: "approve-v2"},
			{ID: "e2", Source: "approve-v2", Target: "end-v2"},
		},
	}
	detail := &application.InstanceDetail{
		Instance: application.InstanceSummary{DefinitionVersion: 2, Status: "running", StartTime: 1000},
		Tasks:    []application.TaskSummary{{ID: "task-1", NodeID: "approve-v2", Status: "pending"}},
	}

	populateInstanceNodeProgress(detail, definition)

	if len(detail.NodeProgress) != 3 ||
		detail.NodeProgress[0].NodeID != "start-v2" ||
		detail.NodeProgress[1].Status != application.NodeProgressProcessing ||
		detail.NodeProgress[2].Status != application.NodeProgressNotStarted {
		t.Fatalf("node progress = %#v", detail.NodeProgress)
	}
}
