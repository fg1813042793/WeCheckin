package domain

import (
	"errors"
	"fmt"
	"testing"

	workflowcore "wecheckin/backend/internal/workflow"
)

type staticResolver map[string][]string

func (resolver staticResolver) Resolve(request AssigneeRequest) ([]string, error) {
	assignees := resolver[request.Node.ID]
	if len(assignees) == 0 {
		return nil, fmt.Errorf("node %s has no assignees", request.Node.ID)
	}
	return assignees, nil
}

type sequenceIDs int

func (ids *sequenceIDs) NewID(prefix string) string {
	*ids++
	return fmt.Sprintf("%s-%d", prefix, *ids)
}

func TestEngineStartsLinearApprovalAndCompletesInstance(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSingle, 0)
	engine := newTestEngine(staticResolver{"approve": {"user-7"}})

	state, err := engine.Start(definition, StartRequest{
		DefinitionID: 8, DefinitionVersion: 2, StarterID: "user-1",
		BusinessType: "leave", BusinessKey: "leave-100",
	})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	pending := state.PendingTasks()
	if state.Instance.Status != InstanceStatusRunning || len(pending) != 1 || pending[0].AssigneeID != "user-7" {
		t.Fatalf("unexpected started state: instance=%#v tasks=%#v", state.Instance, pending)
	}

	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: pending[0].ID, ActorID: "user-7", Action: TaskActionApprove, Comment: "同意",
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if state.Instance.Status != InstanceStatusCompleted || len(state.PendingTasks()) != 0 {
		t.Fatalf("workflow did not complete: instance=%#v tasks=%#v", state.Instance, state.Tasks)
	}
}

func TestEngineExclusiveGatewaySelectsConditionAndDefaultBranch(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "expense_review",
		Name:          "费用审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "gateway", Type: workflowcore.NodeTypeExclusive, Name: "金额判断", GatewayMode: workflowcore.GatewayModeSplit},
			approvalNode("manager", workflowcore.ApprovalModeSingle, 0),
			approvalNode("finance", workflowcore.ApprovalModeSingle, 0),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "gateway"},
			{ID: "e2", Source: "gateway", Target: "finance", Condition: &workflowcore.Condition{Field: "amount", Operator: workflowcore.ConditionGTE, Value: 1000}},
			{ID: "e3", Source: "gateway", Target: "manager", Default: true},
			{ID: "e4", Source: "manager", Target: "end"},
			{ID: "e5", Source: "finance", Target: "end"},
		},
	}
	resolver := staticResolver{"manager": {"manager-1"}, "finance": {"finance-1"}}

	high, err := newTestEngine(resolver).Start(definition, StartRequest{StarterID: "u1", Variables: map[string]interface{}{"amount": 1200}})
	if err != nil {
		t.Fatalf("start high amount: %v", err)
	}
	if got := high.PendingTasks(); len(got) != 1 || got[0].NodeID != "finance" {
		t.Fatalf("high amount selected wrong branch: %#v", got)
	}

	low, err := newTestEngine(resolver).Start(definition, StartRequest{StarterID: "u1", Variables: map[string]interface{}{"amount": 99}})
	if err != nil {
		t.Fatalf("start low amount: %v", err)
	}
	if got := low.PendingTasks(); len(got) != 1 || got[0].NodeID != "manager" {
		t.Fatalf("low amount did not select default branch: %#v", got)
	}
}

func TestEngineParallelJoinWaitsForEveryBranch(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "parallel_review",
		Name:          "并行审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "split", Type: workflowcore.NodeTypeParallel, Name: "并行开始", GatewayMode: workflowcore.GatewayModeSplit},
			approvalNode("legal", workflowcore.ApprovalModeSingle, 0),
			approvalNode("finance", workflowcore.ApprovalModeSingle, 0),
			{ID: "join", Type: workflowcore.NodeTypeParallel, Name: "并行汇聚", GatewayMode: workflowcore.GatewayModeJoin},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "split"},
			{ID: "e2", Source: "split", Target: "legal"},
			{ID: "e3", Source: "split", Target: "finance"},
			{ID: "e4", Source: "legal", Target: "join"},
			{ID: "e5", Source: "finance", Target: "join"},
			{ID: "e6", Source: "join", Target: "end"},
		},
	}
	engine := newTestEngine(staticResolver{"legal": {"legal-1"}, "finance": {"finance-1"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "u1"})
	if err != nil {
		t.Fatalf("start parallel workflow: %v", err)
	}
	tasks := state.PendingTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected two parallel tasks, got %#v", tasks)
	}
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: tasks[0].ID, ActorID: tasks[0].AssigneeID, Action: TaskActionApprove}); err != nil {
		t.Fatalf("complete first branch: %v", err)
	}
	if state.Instance.Status != InstanceStatusRunning || len(state.PendingTasks()) != 1 {
		t.Fatalf("parallel join did not wait: instance=%#v tasks=%#v", state.Instance, state.Tasks)
	}
	remaining := state.PendingTasks()[0]
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: remaining.ID, ActorID: remaining.AssigneeID, Action: TaskActionApprove}); err != nil {
		t.Fatalf("complete second branch: %v", err)
	}
	if state.Instance.Status != InstanceStatusCompleted {
		t.Fatalf("parallel workflow did not complete: %#v", state.Instance)
	}
}

func TestEngineSequentialApprovalActivatesOneTaskAtATime(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSequential, 0)
	engine := newTestEngine(staticResolver{"approve": {"u1", "u2", "u3"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start sequential workflow: %v", err)
	}
	for index, assignee := range []string{"u1", "u2", "u3"} {
		pending := state.PendingTasks()
		if len(pending) != 1 || pending[0].AssigneeID != assignee {
			t.Fatalf("step %d expected %s, got %#v", index, assignee, pending)
		}
		if err := engine.Complete(definition, state, CompleteRequest{TaskID: pending[0].ID, ActorID: assignee, Action: TaskActionApprove}); err != nil {
			t.Fatalf("complete sequential task %d: %v", index, err)
		}
	}
	if state.Instance.Status != InstanceStatusCompleted {
		t.Fatalf("sequential workflow did not complete: %#v", state.Instance)
	}
}

func TestEngineCountersignCompletesAtThresholdAndCancelsRemainder(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeCountersign, 60)
	engine := newTestEngine(staticResolver{"approve": {"u1", "u2", "u3"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start countersign workflow: %v", err)
	}
	tasks := state.PendingTasks()
	if len(tasks) != 3 {
		t.Fatalf("expected three countersign tasks, got %#v", tasks)
	}
	for _, task := range tasks[:2] {
		if err := engine.Complete(definition, state, CompleteRequest{TaskID: task.ID, ActorID: task.AssigneeID, Action: TaskActionApprove}); err != nil {
			t.Fatalf("complete countersign task: %v", err)
		}
	}
	if state.Instance.Status != InstanceStatusCompleted || len(state.PendingTasks()) != 0 {
		t.Fatalf("countersign did not complete at threshold: instance=%#v tasks=%#v", state.Instance, state.Tasks)
	}
	cancelled := 0
	for _, task := range state.Tasks {
		if task.Status == TaskStatusCancelled {
			cancelled++
		}
	}
	if cancelled != 1 {
		t.Fatalf("expected one cancelled task, got %#v", state.Tasks)
	}
}

func TestEngineRejectsTaskFromDifferentActor(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSingle, 0)
	engine := newTestEngine(staticResolver{"approve": {"owner"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	err = engine.Complete(definition, state, CompleteRequest{TaskID: state.PendingTasks()[0].ID, ActorID: "other", Action: TaskActionApprove})
	if err == nil {
		t.Fatal("expected assignee validation error")
	}
}

func TestEngineKeepsAndUpdatesFormData(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSingle, 0)
	engine := newTestEngine(staticResolver{"approve": {"owner"}})
	state, err := engine.Start(definition, StartRequest{
		StarterID: "starter",
		FormData:  map[string]interface{}{"reason": "出差", "amount": 100},
	})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	if state.FormData["reason"] != "出差" || state.FormData["amount"] != 100 {
		t.Fatalf("form data was not kept: %#v", state.FormData)
	}
	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: state.PendingTasks()[0].ID, ActorID: "owner", Action: TaskActionApprove,
		FormData: map[string]interface{}{"amount": 120},
	}); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if state.FormData["reason"] != "出差" || state.FormData["amount"] != 120 {
		t.Fatalf("form patch was not merged: %#v", state.FormData)
	}
}

func TestEngineWithdrawsOnlyUntouchedInstanceByStarter(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSingle, 0)
	engine := newTestEngine(staticResolver{"approve": {"owner"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}

	if err := engine.Withdraw(state, "other", "不需要了"); !errors.Is(err, ErrInstanceStarterMismatch) {
		t.Fatalf("withdraw by other user error = %v, want ErrInstanceStarterMismatch", err)
	}
	if err := engine.Withdraw(state, "starter", "不需要了"); err != nil {
		t.Fatalf("withdraw by starter: %v", err)
	}
	if state.Instance.Status != InstanceStatusCancelled || len(state.PendingTasks()) != 0 {
		t.Fatalf("withdraw did not terminate instance: %#v %#v", state.Instance, state.Tasks)
	}
	if !hasHistoryType(state.History, HistoryInstanceWithdrawn) {
		t.Fatalf("withdraw history missing: %#v", state.History)
	}
}

func TestEngineRejectsWithdrawAfterTaskHandled(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "two_steps",
		Name:          "两级审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			approvalNode("first", workflowcore.ApprovalModeSingle, 0),
			approvalNode("second", workflowcore.ApprovalModeSingle, 0),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "first"},
			{ID: "e2", Source: "first", Target: "second"},
			{ID: "e3", Source: "second", Target: "end"},
		},
	}
	engine := newTestEngine(staticResolver{"first": {"u1"}, "second": {"u2"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	first := state.PendingTasks()[0]
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: first.ID, ActorID: "u1", Action: TaskActionApprove}); err != nil {
		t.Fatalf("complete first task: %v", err)
	}
	if err := engine.Withdraw(state, "starter", "撤回"); !errors.Is(err, ErrInstanceAlreadyHandled) {
		t.Fatalf("withdraw error = %v, want ErrInstanceAlreadyHandled", err)
	}
}

func TestEngineAdminCancelTerminatesRunningInstance(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeParallel, 0)
	engine := newTestEngine(staticResolver{"approve": {"u1", "u2"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	if err := engine.Cancel(state, "admin-1", "业务单已作废"); err != nil {
		t.Fatalf("cancel workflow: %v", err)
	}
	if state.Instance.Status != InstanceStatusCancelled || len(state.PendingTasks()) != 0 {
		t.Fatalf("cancel did not terminate instance: %#v %#v", state.Instance, state.Tasks)
	}
	if !hasHistoryType(state.History, HistoryInstanceCancelled) {
		t.Fatalf("cancel history missing: %#v", state.History)
	}
}

func hasHistoryType(history []HistoryEvent, eventType HistoryEventType) bool {
	for _, event := range history {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func newTestEngine(resolver AssigneeResolver) *Engine {
	ids := sequenceIDs(0)
	return NewEngine(resolver, &ids)
}

func linearDefinition(mode string, completionRate int) workflowcore.Definition {
	return workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "linear_review",
		Name:          "线性审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			approvalNode("approve", mode, completionRate),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "approve"},
			{ID: "e2", Source: "approve", Target: "end"},
		},
	}
}

func approvalNode(id, mode string, completionRate int) workflowcore.Node {
	return workflowcore.Node{
		ID: id, Type: workflowcore.NodeTypeApproval, Name: id,
		ApprovalMode: mode, CompletionRate: completionRate,
		Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: id},
	}
}
