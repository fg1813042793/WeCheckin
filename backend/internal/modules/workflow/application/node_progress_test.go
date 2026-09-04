package application

import (
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestBuildNodeProgressLinearFlow(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "发起"},
			{ID: "approve", Type: workflowcore.NodeTypeApproval, Name: "主管审批"},
			{ID: "handle", Type: workflowcore.NodeTypeHandle, Name: "归档"},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "approve"},
			{ID: "e2", Source: "approve", Target: "handle"},
			{ID: "e3", Source: "handle", Target: "end"},
		},
	}
	actual := BuildNodeProgress(
		definition,
		InstanceSummary{Status: "running", StartTime: 1000},
		[]TokenSummary{{ID: "token-1", NodeID: "handle", Status: "waiting"}},
		[]TaskSummary{
			{ID: "task-1", NodeID: "approve", Status: "approved"},
			{ID: "task-2", NodeID: "handle", Status: "pending"},
		},
		nil,
	)

	assertNodeProgress(t, actual, []nodeProgressExpectation{
		{nodeID: "start", status: NodeProgressCompleted},
		{nodeID: "approve", status: NodeProgressCompleted},
		{nodeID: "handle", status: NodeProgressProcessing},
		{nodeID: "end", status: NodeProgressNotStarted},
	})
}

func TestBuildNodeProgressMarksUnselectedConditionBranchSkipped(t *testing.T) {
	actual := BuildNodeProgress(
		conditionalDefinition(),
		InstanceSummary{Status: "completed", StartTime: 1000, EndTime: 2000},
		[]TokenSummary{{ID: "token-1", NodeID: "end", Status: "completed"}},
		[]TaskSummary{{ID: "task-finance", NodeID: "finance", Status: "approved"}},
		[]HistorySummary{{ID: "history-finance", EventType: "task_approved", NodeID: "finance", EventTime: 1500}},
	)
	status := progressStatusByNode(actual)
	for _, nodeID := range []string{"start", "split", "finance", "join", "end"} {
		if status[nodeID] != NodeProgressCompleted {
			t.Fatalf("node %s status = %q, want completed", nodeID, status[nodeID])
		}
	}
	if status["hr"] != NodeProgressSkipped {
		t.Fatalf("hr status = %q, want skipped", status["hr"])
	}
}

func TestBuildNodeProgressMarksRemainingNodesTerminatedAfterReject(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart},
			{ID: "approve", Type: workflowcore.NodeTypeApproval},
			{ID: "archive", Type: workflowcore.NodeTypeHandle},
			{ID: "end", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "approve"},
			{ID: "e2", Source: "approve", Target: "archive"},
			{ID: "e3", Source: "archive", Target: "end"},
		},
	}
	actual := BuildNodeProgress(
		definition,
		InstanceSummary{Status: "rejected", StartTime: 1000, EndTime: 2000},
		[]TokenSummary{{ID: "token-1", NodeID: "approve", Status: "cancelled"}},
		[]TaskSummary{{ID: "task-1", NodeID: "approve", Status: "rejected"}},
		[]HistorySummary{{ID: "history-1", EventType: "task_rejected", NodeID: "approve", EventTime: 1800}},
	)
	status := progressStatusByNode(actual)
	if status["start"] != NodeProgressCompleted {
		t.Fatalf("start status = %q, want completed", status["start"])
	}
	for _, nodeID := range []string{"approve", "archive", "end"} {
		if status[nodeID] != NodeProgressTerminated {
			t.Fatalf("node %s status = %q, want terminated", nodeID, status[nodeID])
		}
	}
}

func TestBuildNodeProgressMarksReturnedNodeAndRestartedTarget(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart},
			{ID: "manager", Type: workflowcore.NodeTypeApproval},
			{ID: "hr", Type: workflowcore.NodeTypeApproval},
			{ID: "end", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "manager"},
			{ID: "e2", Source: "manager", Target: "hr"},
			{ID: "e3", Source: "hr", Target: "end"},
		},
	}
	status := progressStatusByNode(BuildNodeProgress(
		definition,
		InstanceSummary{Status: "running", StartTime: 1000},
		[]TokenSummary{
			{ID: "old-token", NodeID: "hr", Status: "cancelled"},
			{ID: "new-token", NodeID: "manager", Status: "waiting"},
		},
		[]TaskSummary{
			{ID: "manager-old", NodeID: "manager", Status: "approved"},
			{ID: "hr-returned", NodeID: "hr", Status: "returned"},
			{ID: "manager-new", NodeID: "manager", Status: "pending"},
		},
		[]HistorySummary{{ID: "history-returned", EventType: "task_returned", NodeID: "hr", EventTime: 1800}},
	))
	if status["manager"] != NodeProgressProcessing || status["hr"] != NodeProgressReturned {
		t.Fatalf("node progress after return = %#v", status)
	}
}

func TestBuildNodeProgressTracksParallelBranchesAndJoin(t *testing.T) {
	definition := parallelDefinition()
	cases := []struct {
		name     string
		tokens   []TokenSummary
		tasks    []TaskSummary
		expected map[string]string
	}{
		{
			name: "branches active",
			tokens: []TokenSummary{
				{ID: "left-token", NodeID: "left", Status: "waiting"},
				{ID: "right-token", NodeID: "right", Status: "waiting"},
			},
			tasks: []TaskSummary{
				{ID: "left-task", NodeID: "left", Status: "pending"},
				{ID: "right-task", NodeID: "right", Status: "pending"},
			},
			expected: map[string]string{
				"split": NodeProgressCompleted,
				"left":  NodeProgressProcessing,
				"right": NodeProgressProcessing,
				"join":  NodeProgressNotStarted,
			},
		},
		{
			name:   "joined",
			tokens: []TokenSummary{{ID: "next-token", NodeID: "next", Status: "waiting"}},
			tasks: []TaskSummary{
				{ID: "left-task", NodeID: "left", Status: "approved"},
				{ID: "right-task", NodeID: "right", Status: "approved"},
				{ID: "next-task", NodeID: "next", Status: "pending"},
			},
			expected: map[string]string{
				"split": NodeProgressCompleted,
				"left":  NodeProgressCompleted,
				"right": NodeProgressCompleted,
				"join":  NodeProgressCompleted,
				"next":  NodeProgressProcessing,
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			progress := BuildNodeProgress(
				definition,
				InstanceSummary{Status: "running", StartTime: 1000},
				test.tokens,
				test.tasks,
				nil,
			)
			status := progressStatusByNode(progress)
			for nodeID, expected := range test.expected {
				if status[nodeID] != expected {
					t.Fatalf("node %s status = %q, want %q", nodeID, status[nodeID], expected)
				}
			}
			assertNodeOrder(t, progress, []string{"start", "split", "left", "right", "join", "next", "end"})
		})
	}
}

func TestBuildNodeProgressMarksResumedTimerCompleted(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart},
			{ID: "timer", Type: workflowcore.NodeTypeTimer},
			{ID: "approve", Type: workflowcore.NodeTypeApproval},
			{ID: "end", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "timer"},
			{ID: "e2", Source: "timer", Target: "approve"},
			{ID: "e3", Source: "approve", Target: "end"},
		},
	}
	status := progressStatusByNode(BuildNodeProgress(
		definition,
		InstanceSummary{Status: "running", StartTime: 1000},
		[]TokenSummary{{ID: "token-1", NodeID: "approve", Status: "waiting"}},
		[]TaskSummary{{ID: "task-1", NodeID: "approve", Status: "pending"}},
		[]HistorySummary{
			{ID: "history-1", EventType: "timer_waiting", NodeID: "timer", EventTime: 1100},
			{ID: "history-2", EventType: "timer_resumed", NodeID: "timer", EventTime: 1200},
		},
	))
	if status["timer"] != NodeProgressCompleted {
		t.Fatalf("timer status = %q, want completed", status["timer"])
	}
}

func TestBuildNodeProgressDoesNotTerminateCompletedMultiPersonNode(t *testing.T) {
	definition := workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart},
			{ID: "approve", Type: workflowcore.NodeTypeApproval},
			{ID: "end", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "approve"},
			{ID: "e2", Source: "approve", Target: "end"},
		},
	}
	status := progressStatusByNode(BuildNodeProgress(
		definition,
		InstanceSummary{Status: "completed", StartTime: 1000, EndTime: 2000},
		[]TokenSummary{{ID: "token-1", NodeID: "end", Status: "completed"}},
		[]TaskSummary{
			{ID: "task-1", NodeID: "approve", Status: "approved"},
			{ID: "task-2", NodeID: "approve", Status: "cancelled"},
		},
		[]HistorySummary{
			{ID: "history-1", EventType: "task_approved", NodeID: "approve", TaskID: "task-1", EventTime: 1800},
			{ID: "history-2", EventType: "task_cancelled", NodeID: "approve", TaskID: "task-2", EventTime: 1800},
		},
	))
	if status["approve"] != NodeProgressCompleted {
		t.Fatalf("approve status = %q, want completed", status["approve"])
	}
}

func parallelDefinition() workflowcore.Definition {
	return workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Position: &workflowcore.Position{X: 150, Y: 0}},
			{ID: "split", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeSplit, Position: &workflowcore.Position{X: 150, Y: 100}},
			{ID: "right", Type: workflowcore.NodeTypeApproval, Position: &workflowcore.Position{X: 200, Y: 200}},
			{ID: "left", Type: workflowcore.NodeTypeApproval, Position: &workflowcore.Position{X: 100, Y: 200}},
			{ID: "join", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeJoin, Position: &workflowcore.Position{X: 150, Y: 300}},
			{ID: "next", Type: workflowcore.NodeTypeHandle, Position: &workflowcore.Position{X: 150, Y: 400}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Position: &workflowcore.Position{X: 150, Y: 500}},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "split"},
			{ID: "e2", Source: "split", Target: "left"},
			{ID: "e3", Source: "split", Target: "right"},
			{ID: "e4", Source: "left", Target: "join"},
			{ID: "e5", Source: "right", Target: "join"},
			{ID: "e6", Source: "join", Target: "next"},
			{ID: "e7", Source: "next", Target: "end"},
		},
	}
}

func assertNodeOrder(t *testing.T, progress []NodeProgressSummary, expected []string) {
	t.Helper()
	if len(progress) != len(expected) {
		t.Fatalf("progress length = %d, want %d", len(progress), len(expected))
	}
	for index, nodeID := range expected {
		if progress[index].NodeID != nodeID {
			t.Fatalf("progress[%d].NodeID = %q, want %q", index, progress[index].NodeID, nodeID)
		}
	}
}

func conditionalDefinition() workflowcore.Definition {
	return workflowcore.Definition{
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart},
			{ID: "split", Type: workflowcore.NodeTypeExclusive, GatewayMode: workflowcore.GatewayModeSplit},
			{ID: "finance", Type: workflowcore.NodeTypeApproval},
			{ID: "hr", Type: workflowcore.NodeTypeApproval},
			{ID: "join", Type: workflowcore.NodeTypeExclusive, GatewayMode: workflowcore.GatewayModeJoin},
			{ID: "end", Type: workflowcore.NodeTypeEnd},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "split"},
			{ID: "e2", Source: "split", Target: "finance"},
			{ID: "e3", Source: "split", Target: "hr"},
			{ID: "e4", Source: "finance", Target: "join"},
			{ID: "e5", Source: "hr", Target: "join"},
			{ID: "e6", Source: "join", Target: "end"},
		},
	}
}

func progressStatusByNode(progress []NodeProgressSummary) map[string]string {
	result := make(map[string]string, len(progress))
	for _, item := range progress {
		result[item.NodeID] = item.Status
	}
	return result
}

type nodeProgressExpectation struct {
	nodeID string
	status string
}

func assertNodeProgress(t *testing.T, actual []NodeProgressSummary, expected []nodeProgressExpectation) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("progress length = %d, want %d: %#v", len(actual), len(expected), actual)
	}
	for index := range expected {
		if actual[index].NodeID != expected[index].nodeID || actual[index].Status != expected[index].status {
			t.Fatalf(
				"progress[%d] = (%q, %q), want (%q, %q)",
				index,
				actual[index].NodeID,
				actual[index].Status,
				expected[index].nodeID,
				expected[index].status,
			)
		}
	}
}
