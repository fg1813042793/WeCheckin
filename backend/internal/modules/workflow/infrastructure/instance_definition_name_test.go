package infrastructure

import (
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
)

func TestInstanceSummariesIncludeDefinitionName(t *testing.T) {
	rows := []workflowmodel.ProcessInstance{{
		ID:            "instance-1",
		DefinitionID:  7,
		DefinitionKey: "performance-review",
	}}

	list := instanceSummariesWithDefinitionNames(rows, nil, map[uint]string{7: "绩效单"})
	if len(list) != 1 {
		t.Fatalf("summary count = %d, want 1", len(list))
	}
	if list[0].DefinitionName != "绩效单" {
		t.Fatalf("definition name = %q, want %q", list[0].DefinitionName, "绩效单")
	}
}

func TestInstanceSummariesPreferDefinitionNameSnapshot(t *testing.T) {
	rows := []workflowmodel.ProcessInstance{{
		ID: "instance-1", DefinitionID: 7, DefinitionKey: "performance-review",
		DefinitionNameSnapshot: "绩效考评单",
	}}

	list := instanceSummariesWithDefinitionNames(rows, nil, map[uint]string{7: "绩效考评单（当前管理名称）"})
	if len(list) != 1 || list[0].DefinitionName != "绩效考评单" {
		t.Fatalf("summary must preserve start-time display name snapshot, got %#v", list)
	}
}
