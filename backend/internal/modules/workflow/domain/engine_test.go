package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"wecheckin/backend/internal/workflowcore"
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

func TestEngineTaskCreatedHistoryUsesTriggeringActor(t *testing.T) {
	definition := returnableDefinition()
	engine := newTestEngine(staticResolver{
		"draft": {"starter"}, "manager": {"manager-1"}, "hr": {"hr-1"},
	})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter", OperatorID: "operator"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	created := historyByType(state.History, HistoryTaskCreated)
	if created == nil || created.ActorID != "operator" {
		t.Fatalf("initial task-created history = %#v, want operator", created)
	}

	completePendingTask(t, engine, definition, state, "starter", TaskActionSubmit)
	created = lastHistoryByType(state.History, HistoryTaskCreated)
	if created == nil || created.ActorID != "starter" {
		t.Fatalf("next task-created history = %#v, want starter", created)
	}
}

func TestEngineRejectsTaskWithImages(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSingle, 0)
	engine := newTestEngine(staticResolver{"approve": {"user-7"}})
	state, err := engine.Start(definition, StartRequest{DefinitionID: 8, DefinitionVersion: 2, StarterID: "user-1"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	image := workflowcore.FormAttachment{
		ID: "uploads/workflow/2026/09/04/reject.png", Name: "reject.png",
		URL: "/uploads/workflow/2026/09/04/reject.png", MimeType: "image/png", Size: 1024,
	}
	taskID := state.PendingTasks()[0].ID
	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: taskID, ActorID: "user-7", Action: TaskActionReject,
		Comment: "材料不完整", Images: []workflowcore.FormAttachment{image},
	}); err != nil {
		t.Fatalf("reject task: %v", err)
	}
	if len(state.Tasks[0].Images) != 1 || state.Tasks[0].Images[0].ID != image.ID {
		t.Fatalf("task images = %#v", state.Tasks[0].Images)
	}
	var rejected *HistoryEvent
	for index := range state.History {
		if state.History[index].Type == HistoryTaskRejected {
			rejected = &state.History[index]
			break
		}
	}
	if rejected == nil || len(rejected.Images) != 1 || rejected.Images[0].URL != image.URL {
		t.Fatalf("rejected history = %#v", rejected)
	}
}

func TestEngineReturnsApprovalToPreviousHumanNode(t *testing.T) {
	definition := returnableDefinition()
	engine := newTestEngine(staticResolver{
		"draft": {"starter"}, "manager": {"manager-1"}, "hr": {"hr-1"},
	})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	completePendingTask(t, engine, definition, state, "starter", TaskActionSubmit)
	completePendingTask(t, engine, definition, state, "manager-1", TaskActionApprove)

	current := state.PendingTasks()[0]
	image := workflowcore.FormAttachment{
		ID: "uploads/workflow/2026/09/04/return.png", Name: "return.png",
		URL: "/uploads/workflow/2026/09/04/return.png", MimeType: "image/png", Size: 1024,
	}
	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: current.ID, ActorID: "hr-1", Action: TaskActionReturn,
		Comment: "请主管重新确认", Images: []workflowcore.FormAttachment{image},
	}); err != nil {
		t.Fatalf("return task: %v", err)
	}

	if state.Instance.Status != InstanceStatusRunning {
		t.Fatalf("instance status = %q, want running", state.Instance.Status)
	}
	returned := taskByID(state.Tasks, current.ID)
	if returned == nil || returned.Status != TaskStatusReturned || returned.Action != TaskActionReturn {
		t.Fatalf("returned task = %#v", returned)
	}
	if len(returned.Images) != 1 || returned.Images[0].ID != image.ID {
		t.Fatalf("returned task images = %#v", returned.Images)
	}
	pending := state.PendingTasks()
	if len(pending) != 1 || pending[0].NodeID != "manager" || pending[0].AssigneeID != "manager-1" || pending[0].ID == current.ID {
		t.Fatalf("pending tasks after return = %#v", pending)
	}
	if countHistoryType(state.History, HistoryTaskReturned) != 1 {
		t.Fatalf("returned history = %#v", state.History)
	}
}

func TestEngineReturnsApprovalToSpecifiedVisitedHumanNode(t *testing.T) {
	definition := returnableDefinition()
	engine := newTestEngine(staticResolver{
		"draft": {"starter"}, "manager": {"manager-1"}, "hr": {"hr-1"},
	})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	completePendingTask(t, engine, definition, state, "starter", TaskActionSubmit)
	completePendingTask(t, engine, definition, state, "manager-1", TaskActionApprove)

	current := state.PendingTasks()[0]
	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: current.ID, ActorID: "hr-1", Action: TaskActionReturn,
		ReturnTargetNodeID: "draft", Comment: "请发起人修改",
	}); err != nil {
		t.Fatalf("return task to specified node: %v", err)
	}
	pending := state.PendingTasks()
	if len(pending) != 1 || pending[0].NodeID != "draft" || pending[0].AssigneeID != "starter" {
		t.Fatalf("pending tasks after specified return = %#v", pending)
	}
}

func TestEngineRejectsInvalidReturnTargetWithoutMutatingState(t *testing.T) {
	definition := returnableDefinition()
	engine := newTestEngine(staticResolver{
		"draft": {"starter"}, "manager": {"manager-1"}, "hr": {"hr-1"},
	})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	completePendingTask(t, engine, definition, state, "starter", TaskActionSubmit)
	current := state.PendingTasks()[0]

	err = engine.Complete(definition, state, CompleteRequest{
		TaskID: current.ID, ActorID: "manager-1", Action: TaskActionReturn,
		ReturnTargetNodeID: "hr", Comment: "非法目标",
	})
	if !errors.Is(err, ErrReturnTargetInvalid) {
		t.Fatalf("return error = %v, want ErrReturnTargetInvalid", err)
	}
	if state.Instance.Status != InstanceStatusRunning || len(state.PendingTasks()) != 1 || state.PendingTasks()[0].ID != current.ID {
		t.Fatalf("invalid return mutated state: %#v", state)
	}
}

func TestEngineRejectsReturnWithoutPreviousHumanNode(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "first_approval",
		Name:          "首节点审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			approvalNode("manager", workflowcore.ApprovalModeSingle, 0),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "manager"},
			{ID: "e2", Source: "manager", Target: "end"},
		},
	}
	engine := newTestEngine(staticResolver{"manager": {"manager-1"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	current := state.PendingTasks()[0]

	err = engine.Complete(definition, state, CompleteRequest{
		TaskID: current.ID, ActorID: "manager-1", Action: TaskActionReturn, Comment: "退回",
	})
	if !errors.Is(err, ErrReturnTargetUnavailable) {
		t.Fatalf("return error = %v, want ErrReturnTargetUnavailable", err)
	}
	if state.Instance.Status != InstanceStatusRunning || len(state.PendingTasks()) != 1 || state.PendingTasks()[0].ID != current.ID {
		t.Fatalf("unavailable return mutated state: %#v", state)
	}
}

func TestEngineRejectsReturnAfterParallelTraversalWithoutMutatingState(t *testing.T) {
	definition := returnableDefinition()
	engine := newTestEngine(staticResolver{
		"draft": {"starter"}, "manager": {"manager-1"}, "hr": {"hr-1"},
	})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}
	completePendingTask(t, engine, definition, state, "starter", TaskActionSubmit)
	current := state.PendingTasks()[0]
	state.Tokens[0].BranchGroup = "parallel-1"
	state.Tokens[0].BranchTotal = 2

	err = engine.Complete(definition, state, CompleteRequest{
		TaskID: current.ID, ActorID: "manager-1", Action: TaskActionReturn,
		ReturnTargetNodeID: "draft", Comment: "退回",
	})
	if !errors.Is(err, ErrReturnParallelUnsupported) {
		t.Fatalf("return error = %v, want ErrReturnParallelUnsupported", err)
	}
	if state.Instance.Status != InstanceStatusRunning || len(state.PendingTasks()) != 1 || state.PendingTasks()[0].ID != current.ID {
		t.Fatalf("parallel return mutated state: %#v", state)
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

	missing, err := newTestEngine(resolver).Start(definition, StartRequest{StarterID: "u1"})
	if err != nil {
		t.Fatalf("start workflow without condition value: %v", err)
	}
	if got := missing.PendingTasks(); len(got) != 1 || got[0].NodeID != "manager" {
		t.Fatalf("missing condition value did not select default branch: %#v", got)
	}
}

func TestEngineExclusiveGatewayReadsConditionFromFormData(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "expense_form_review",
		Name:          "费用表单审批",
		Form: []workflowcore.FormField{
			{Key: "amount", Label: "申请金额", Type: workflowcore.FormFieldTypeAmount},
		},
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

	high, err := newTestEngine(resolver).Start(definition, StartRequest{
		StarterID: "u1",
		FormData:  map[string]interface{}{"amount": 1200},
	})
	if err != nil {
		t.Fatalf("start high amount form workflow: %v", err)
	}
	if got := high.PendingTasks(); len(got) != 1 || got[0].NodeID != "finance" {
		t.Fatalf("form amount selected wrong branch: %#v", got)
	}

	low, err := newTestEngine(resolver).Start(definition, StartRequest{
		StarterID: "u1",
		FormData:  map[string]interface{}{"amount": 99},
	})
	if err != nil {
		t.Fatalf("start low amount form workflow: %v", err)
	}
	if got := low.PendingTasks(); len(got) != 1 || got[0].NodeID != "manager" {
		t.Fatalf("low form amount did not select default branch: %#v", got)
	}
}

func TestEngineExclusiveGatewayReadsUpdatedFormDataAfterApproval(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "updated_form_review",
		Name:          "更新表单审批",
		Form: []workflowcore.FormField{
			{Key: "status", Label: "资料状态", Type: workflowcore.FormFieldTypeText},
		},
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			approvalNode("fill", workflowcore.ApprovalModeSingle, 0),
			{ID: "gateway", Type: workflowcore.NodeTypeExclusive, Name: "资料判断", GatewayMode: workflowcore.GatewayModeSplit},
			approvalNode("manager", workflowcore.ApprovalModeSingle, 0),
			approvalNode("finance", workflowcore.ApprovalModeSingle, 0),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "fill"},
			{ID: "e2", Source: "fill", Target: "gateway"},
			{ID: "e3", Source: "gateway", Target: "finance", Condition: &workflowcore.Condition{Field: "status", Operator: workflowcore.ConditionEQ, Value: "ready"}},
			{ID: "e4", Source: "gateway", Target: "manager", Default: true},
			{ID: "e5", Source: "manager", Target: "end"},
			{ID: "e6", Source: "finance", Target: "end"},
		},
	}
	resolver := staticResolver{"fill": {"owner"}, "manager": {"manager-1"}, "finance": {"finance-1"}}
	engine := newTestEngine(resolver)
	state, err := engine.Start(definition, StartRequest{
		StarterID: "u1",
		FormData:  map[string]interface{}{"status": "draft"},
	})
	if err != nil {
		t.Fatalf("start form workflow: %v", err)
	}

	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID:   state.PendingTasks()[0].ID,
		ActorID:  "owner",
		Action:   TaskActionApprove,
		FormData: map[string]interface{}{"status": "ready"},
	}); err != nil {
		t.Fatalf("complete fill task: %v", err)
	}
	if got := state.PendingTasks(); len(got) != 1 || got[0].NodeID != "finance" {
		t.Fatalf("updated form status selected wrong branch: %#v", got)
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

func TestEngineHandleTaskOnlyAcceptsSubmit(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "profile_update",
		Name:          "资料填写",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "handle", Type: workflowcore.NodeTypeHandle, Name: "填写资料", Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeVariable, Value: "targetUserId"}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "handle"},
			{ID: "e2", Source: "handle", Target: "end"},
		},
	}
	engine := newTestEngine(staticResolver{"handle": {"user-9"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "admin-1"})
	if err != nil {
		t.Fatalf("start handle workflow: %v", err)
	}
	pending := state.PendingTasks()
	if len(pending) != 1 || pending[0].AssigneeID != "user-9" {
		t.Fatalf("unexpected handle tasks: %#v", pending)
	}
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: pending[0].ID, ActorID: "user-9", Action: TaskActionApprove}); !errors.Is(err, ErrInvalidTaskAction) {
		t.Fatalf("approve handle task error = %v, want ErrInvalidTaskAction", err)
	}
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: pending[0].ID, ActorID: "user-9", Action: TaskAction("submit")}); err != nil {
		t.Fatalf("submit handle task: %v", err)
	}
	if state.Tasks[0].Status != TaskStatus("completed") || state.Instance.Status != InstanceStatusCompleted {
		t.Fatalf("handle workflow not completed: task=%#v instance=%#v", state.Tasks[0], state.Instance)
	}
}

func TestEngineCCAndAutomationRunWithoutBlocking(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "notify_and_mark",
		Name:          "通知并标记",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "cc", Type: workflowcore.NodeTypeCC, Name: "抄送相关人",
				Assignee:     &workflowcore.Assignee{Type: workflowcore.AssigneeTypeVariable, Value: "ccUserIds"},
				Notification: notificationConfig("流程抄送", "{{starterName}} 发起的流程已抄送给你"),
			},
			{ID: "automation", Type: workflowcore.NodeTypeAutomation, Name: "写入变量", Automation: &workflowcore.AutomationConfig{Type: workflowcore.AutomationTypeSetVariables, Variables: map[string]interface{}{"processed": true}}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "cc"},
			{ID: "e2", Source: "cc", Target: "automation"},
			{ID: "e3", Source: "automation", Target: "end"},
		},
	}
	state, err := newTestEngine(staticResolver{"cc": {"user-2", "user-3"}}).Start(definition, StartRequest{StarterID: "admin-1"})
	if err != nil {
		t.Fatalf("start automatic workflow: %v", err)
	}
	if state.Instance.Status != InstanceStatusCompleted || len(state.Tasks) != 0 || state.Variables["processed"] != true {
		t.Fatalf("unexpected automatic state: instance=%#v tasks=%#v variables=%#v", state.Instance, state.Tasks, state.Variables)
	}
	if countHistoryType(state.History, HistoryEventType("node_cc")) != 2 {
		t.Fatalf("cc history = %#v", state.History)
	}
	if countHistoryType(state.History, HistoryEventType("node_automated")) != 1 {
		t.Fatalf("automation history = %#v", state.History)
	}
	if len(state.Participants) != 2 || state.Participants[0].Role != ParticipantRoleCC || state.Participants[1].Role != ParticipantRoleCC {
		t.Fatalf("cc participants = %#v", state.Participants)
	}
	if len(state.NotificationIntents) != 2 {
		t.Fatalf("cc notification intents = %#v", state.NotificationIntents)
	}
	for _, intent := range state.NotificationIntents {
		if intent.Kind != NotificationKindNodeCC || intent.NodeID != "cc" || intent.TaskID != "" {
			t.Fatalf("unexpected cc notification intent: %#v", intent)
		}
	}
}

func TestEngineNotifyNodeCreatesMessageIntentWithoutParticipant(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "notify_only",
		Name:          "纯通知流程",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "notify", Type: workflowcore.NodeTypeNotify, Name: "通知观察人",
				Assignee:     &workflowcore.Assignee{Type: workflowcore.AssigneeTypeVariable, Value: "observerIds"},
				Notification: notificationConfig("流程提醒", "流程已到达 {{nodeName}}"),
			},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{{ID: "e1", Source: "start", Target: "notify"}, {ID: "e2", Source: "notify", Target: "end"}},
	}
	state, err := newTestEngine(staticResolver{"notify": {"user-2", "user-2", "user-3"}}).Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start notify workflow: %v", err)
	}
	if state.Instance.Status != InstanceStatusCompleted || len(state.Tasks) != 0 || len(state.Participants) != 0 {
		t.Fatalf("notify node must not block or grant access: %#v", state)
	}
	if countHistoryType(state.History, HistoryNodeNotify) != 2 || len(state.NotificationIntents) != 2 {
		t.Fatalf("notify effects = history %#v intents %#v", state.History, state.NotificationIntents)
	}
	for _, intent := range state.NotificationIntents {
		if intent.Kind != NotificationKindNodeNotify || intent.NodeID != "notify" {
			t.Fatalf("unexpected notify intent: %#v", intent)
		}
	}
}

func TestEngineTaskArrivalNotificationsFollowTaskActivation(t *testing.T) {
	definition := linearDefinition(workflowcore.ApprovalModeSequential, 0)
	definition.Nodes[1].Notification = notificationConfig("待办提醒", "请处理 {{nodeName}}")
	engine := newTestEngine(staticResolver{"approve": {"u1", "u2", "u3"}})
	state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
	if err != nil {
		t.Fatalf("start sequential workflow: %v", err)
	}
	if len(state.NotificationIntents) != 1 || state.NotificationIntents[0].RecipientUserID != "u1" {
		t.Fatalf("initial task notifications = %#v", state.NotificationIntents)
	}
	first := state.PendingTasks()[0]
	if err := engine.Complete(definition, state, CompleteRequest{TaskID: first.ID, ActorID: "u1", Action: TaskActionApprove}); err != nil {
		t.Fatalf("complete first task: %v", err)
	}
	if len(state.NotificationIntents) != 2 || state.NotificationIntents[1].RecipientUserID != "u2" || state.NotificationIntents[1].Kind != NotificationKindTaskArrived {
		t.Fatalf("activated task notifications = %#v", state.NotificationIntents)
	}
}

func TestEngineApprovalResultNotificationsTargetStarter(t *testing.T) {
	newState := func(mode string, assignees []string) (workflowcore.Definition, *Engine, *State) {
		t.Helper()
		definition := linearDefinition(mode, 0)
		definition.Nodes[1].ResultNotification = notificationConfig(
			"{{workflowName}}审批结果",
			"你发起的流程在“{{nodeName}}”节点{{result}}",
		)
		engine := newTestEngine(staticResolver{"approve": assignees})
		state, err := engine.Start(definition, StartRequest{StarterID: "starter"})
		if err != nil {
			t.Fatalf("start workflow: %v", err)
		}
		return definition, engine, state
	}

	t.Run("approved", func(t *testing.T) {
		definition, engine, state := newState(workflowcore.ApprovalModeSingle, []string{"approver"})
		task := state.PendingTasks()[0]
		if err := engine.Complete(definition, state, CompleteRequest{
			TaskID: task.ID, ActorID: task.AssigneeID, Action: TaskActionApprove,
		}); err != nil {
			t.Fatalf("approve task: %v", err)
		}
		assertApprovalResultNotification(t, state.NotificationIntents, NotificationKindApprovalResultApproved, "starter", task.ID)
	})

	t.Run("rejected", func(t *testing.T) {
		definition, engine, state := newState(workflowcore.ApprovalModeSingle, []string{"approver"})
		task := state.PendingTasks()[0]
		if err := engine.Complete(definition, state, CompleteRequest{
			TaskID: task.ID, ActorID: task.AssigneeID, Action: TaskActionReject, Comment: "材料不完整",
		}); err != nil {
			t.Fatalf("reject task: %v", err)
		}
		assertApprovalResultNotification(t, state.NotificationIntents, NotificationKindApprovalResultRejected, "starter", task.ID)
	})

	t.Run("sequential waits for node completion", func(t *testing.T) {
		definition, engine, state := newState(workflowcore.ApprovalModeSequential, []string{"approver-1", "approver-2"})
		first := state.PendingTasks()[0]
		if err := engine.Complete(definition, state, CompleteRequest{
			TaskID: first.ID, ActorID: first.AssigneeID, Action: TaskActionApprove,
		}); err != nil {
			t.Fatalf("approve first task: %v", err)
		}
		if len(state.NotificationIntents) != 0 {
			t.Fatalf("partial sequential approval notifications = %#v", state.NotificationIntents)
		}
		second := state.PendingTasks()[0]
		if err := engine.Complete(definition, state, CompleteRequest{
			TaskID: second.ID, ActorID: second.AssigneeID, Action: TaskActionApprove,
		}); err != nil {
			t.Fatalf("approve second task: %v", err)
		}
		assertApprovalResultNotification(t, state.NotificationIntents, NotificationKindApprovalResultApproved, "starter", second.ID)
	})
}

func TestEngineTimerWaitsUntilDueAndResumesOnce(t *testing.T) {
	definition := workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "delayed_finish",
		Name:          "延时完成",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{ID: "timer", Type: workflowcore.NodeTypeTimer, Name: "等待", Timer: &workflowcore.TimerConfig{DelaySeconds: 30}},
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "timer"},
			{ID: "e2", Source: "timer", Target: "end"},
		},
	}
	startedAt := time.Now().Unix()
	engine := newTestEngine(staticResolver{})
	state, err := engine.Start(definition, StartRequest{StarterID: "admin-1"})
	if err != nil {
		t.Fatalf("start timer workflow: %v", err)
	}
	if state.Instance.Status != InstanceStatusRunning || len(state.Tokens) != 1 || state.Tokens[0].Status != TokenStatusWaiting {
		t.Fatalf("timer did not wait: instance=%#v tokens=%#v", state.Instance, state.Tokens)
	}
	resumer, ok := interface{}(engine).(interface {
		ResumeTimers(workflowcore.Definition, *State, int64) (int, error)
	})
	if !ok {
		t.Fatal("timer resume support missing")
	}
	advanced, err := resumer.ResumeTimers(definition, state, startedAt+29)
	if err != nil || advanced != 0 || state.Instance.Status != InstanceStatusRunning {
		t.Fatalf("timer advanced before due: advanced=%d err=%v state=%#v", advanced, err, state.Instance)
	}
	advanced, err = resumer.ResumeTimers(definition, state, startedAt+31)
	if err != nil || advanced != 1 || state.Instance.Status != InstanceStatusCompleted {
		t.Fatalf("timer did not advance after due: advanced=%d err=%v state=%#v", advanced, err, state.Instance)
	}
	advanced, err = resumer.ResumeTimers(definition, state, startedAt+60)
	if err != nil || advanced != 0 {
		t.Fatalf("timer resumed twice: advanced=%d err=%v", advanced, err)
	}
}

func countHistoryType(history []HistoryEvent, eventType HistoryEventType) int {
	count := 0
	for _, event := range history {
		if event.Type == eventType {
			count++
		}
	}
	return count
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

func returnableDefinition() workflowcore.Definition {
	return workflowcore.Definition{
		SchemaVersion: workflowcore.CurrentSchemaVersion,
		Key:           "returnable_review",
		Name:          "可退回审批",
		Nodes: []workflowcore.Node{
			{ID: "start", Type: workflowcore.NodeTypeStart, Name: "开始"},
			{
				ID: "draft", Type: workflowcore.NodeTypeHandle, Name: "填写申请",
				Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeInitiator},
			},
			approvalNode("manager", workflowcore.ApprovalModeSingle, 0),
			approvalNode("hr", workflowcore.ApprovalModeSingle, 0),
			{ID: "end", Type: workflowcore.NodeTypeEnd, Name: "结束"},
		},
		Edges: []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "draft"},
			{ID: "e2", Source: "draft", Target: "manager"},
			{ID: "e3", Source: "manager", Target: "hr"},
			{ID: "e4", Source: "hr", Target: "end"},
		},
	}
}

func completePendingTask(
	t *testing.T,
	engine *Engine,
	definition workflowcore.Definition,
	state *State,
	actorID string,
	action TaskAction,
) {
	t.Helper()
	pending := state.PendingTasks()
	if len(pending) != 1 {
		t.Fatalf("pending tasks = %#v, want one", pending)
	}
	if err := engine.Complete(definition, state, CompleteRequest{
		TaskID: pending[0].ID, ActorID: actorID, Action: action,
	}); err != nil {
		t.Fatalf("complete task %s: %v", pending[0].ID, err)
	}
}

func taskByID(tasks []Task, taskID string) *Task {
	for index := range tasks {
		if tasks[index].ID == taskID {
			return &tasks[index]
		}
	}
	return nil
}

func historyByType(history []HistoryEvent, eventType HistoryEventType) *HistoryEvent {
	for index := range history {
		if history[index].Type == eventType {
			return &history[index]
		}
	}
	return nil
}

func lastHistoryByType(history []HistoryEvent, eventType HistoryEventType) *HistoryEvent {
	for index := len(history) - 1; index >= 0; index-- {
		if history[index].Type == eventType {
			return &history[index]
		}
	}
	return nil
}

func approvalNode(id, mode string, completionRate int) workflowcore.Node {
	return workflowcore.Node{
		ID: id, Type: workflowcore.NodeTypeApproval, Name: id,
		ApprovalMode: mode, CompletionRate: completionRate,
		Assignee: &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: id},
	}
}

func notificationConfig(title, content string) *workflowcore.NotificationConfig {
	return &workflowcore.NotificationConfig{
		Enabled: true, Channels: []string{workflowcore.NotificationChannelInApp, workflowcore.NotificationChannelDingTalkOA},
		Title: title, Content: content,
	}
}

func assertApprovalResultNotification(
	t *testing.T,
	intents []NotificationIntent,
	wantKind NotificationKind,
	wantRecipient, wantTaskID string,
) {
	t.Helper()
	if len(intents) != 1 {
		t.Fatalf("approval result notifications = %#v", intents)
	}
	intent := intents[0]
	if intent.Kind != wantKind || intent.RecipientUserID != wantRecipient || intent.TaskID != wantTaskID || intent.NodeID != "approve" {
		t.Fatalf("approval result notification = %#v", intent)
	}
	if !intent.Config.Enabled || len(intent.Config.Channels) != 2 || intent.Config.Content != "你发起的流程在“{{nodeName}}”节点{{result}}" {
		t.Fatalf("approval result notification config = %#v", intent.Config)
	}
}
