package application

import "testing"

func TestNodeProgressExplicitSkipAndRevisits(t *testing.T) {
	for _, test := range []struct {
		name    string
		history []HistorySummary
		tasks   []TaskSummary
		tokens  []TokenSummary
		want    string
	}{
		{name: "explicit skip during running", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}}, want: NodeProgressSkipped},
		{name: "new skip replaces older approval", history: []HistorySummary{{EventType: "task_approved", NodeID: "finance"}, {EventType: "node_skipped", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "approved"}}, want: NodeProgressSkipped},
		{name: "new skip replaces older return", history: []HistorySummary{{EventType: "task_returned", NodeID: "finance"}, {EventType: "node_skipped", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "returned"}}, tokens: []TokenSummary{{NodeID: "finance", Status: "cancelled"}}, want: NodeProgressSkipped},
		{name: "new pending visit replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_created", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "pending"}}, want: NodeProgressProcessing},
		{name: "active token replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}}, tokens: []TokenSummary{{NodeID: "finance", Status: "waiting"}}, want: NodeProgressProcessing},
		{name: "later completion replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_created", NodeID: "finance"}, {EventType: "task_approved", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "approved"}}, want: NodeProgressCompleted},
		{name: "later history completion replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_approved", NodeID: "finance"}}, want: NodeProgressCompleted},
		{name: "completed revisit supersedes old cancelled token", history: []HistorySummary{{EventType: "task_returned", NodeID: "finance"}, {EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_created", NodeID: "finance"}, {EventType: "task_approved", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "returned"}, {NodeID: "finance", Status: "approved"}}, tokens: []TokenSummary{{NodeID: "finance", Status: "cancelled"}}, want: NodeProgressCompleted},
		{name: "countersign revisit preserves completion", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_created", NodeID: "finance"}, {EventType: "task_approved", NodeID: "finance"}, {EventType: "task_cancelled", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "approved"}, {NodeID: "finance", Status: "cancelled"}}, want: NodeProgressCompleted},
		{name: "history pending revisit replaces skip and old completion", history: []HistorySummary{{EventType: "task_approved", NodeID: "finance"}, {EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_created", NodeID: "finance"}}, want: NodeProgressProcessing},
		{name: "later rejection replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_rejected", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "rejected"}}, want: NodeProgressTerminated},
		{name: "later return replaces skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "task_returned", NodeID: "finance"}}, tasks: []TaskSummary{{NodeID: "finance", Status: "returned"}}, want: NodeProgressReturned},
		{name: "later comment preserves skip", history: []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}, {EventType: "instance_commented", NodeID: "finance"}}, want: NodeProgressSkipped},
	} {
		t.Run(test.name, func(t *testing.T) {
			status := progressStatusByNode(BuildNodeProgress(conditionalDefinition(), InstanceSummary{Status: "running"}, test.tokens, test.tasks, test.history))
			if status["finance"] != test.want {
				t.Fatalf("finance = %s, want %s: %#v", status["finance"], test.want, status)
			}
			if status["split"] != NodeProgressCompleted || status["hr"] != NodeProgressNotStarted || status["end"] != NodeProgressNotStarted {
				t.Fatalf("skip traversal evidence or unvisited nodes incorrect: %#v", status)
			}
		})
	}
}

func TestNodeProgressExplicitSkipSurvivesInstanceTermination(t *testing.T) {
	for _, instanceStatus := range []string{"running", "completed", "voided", "rejected", "cancelled"} {
		t.Run(instanceStatus, func(t *testing.T) {
			status := progressStatusByNode(BuildNodeProgress(conditionalDefinition(), InstanceSummary{Status: instanceStatus}, nil, nil, []HistorySummary{{EventType: "node_skipped", NodeID: "finance"}}))
			if status["finance"] != NodeProgressSkipped || status["split"] != NodeProgressCompleted {
				t.Fatalf("explicit skip lost: %#v", status)
			}
		})
	}
}
