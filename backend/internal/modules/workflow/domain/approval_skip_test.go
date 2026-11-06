package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

type skipTestResolver struct {
	users map[string][]string
	err   error
}

func (resolver skipTestResolver) Resolve(_ context.Context, request AssigneeRequest) ([]string, error) {
	return resolver.users[request.Node.ID], resolver.err
}

type skipTestLayerResolver struct {
	skipTestResolver
	layers []ApprovalLayer
	err    error
}

func (resolver skipTestLayerResolver) ResolveApprovalLayers(context.Context, AssigneeRequest) ([]ApprovalLayer, error) {
	return resolver.layers, resolver.err
}

func enableApprovalSkip(t *testing.T, node *workflowcore.Node) {
	t.Helper()
	if err := json.Unmarshal([]byte(`{"skipIfNoAssignee":true}`), node); err != nil {
		t.Fatal(err)
	}
}

func skipTestDefinition(t *testing.T, mode string, enabled, downstream bool) workflowcore.Definition {
	t.Helper()
	definition := linearDefinition(mode, 100)
	if enabled {
		enableApprovalSkip(t, &definition.Nodes[1])
	}
	definition.Nodes[1].Notification = notificationConfig("待审批", "请审批")
	if downstream {
		next := approvalNode("next", workflowcore.ApprovalModeSingle, 100)
		next.Notification = notificationConfig("下一步", "请审批")
		definition.Nodes = append(definition.Nodes, next)
		definition.Edges[1].Target = "next"
		definition.Edges = append(definition.Edges, workflowcore.Edge{ID: "e3", Source: "next", Target: "end"})
	}
	return definition
}

func assertApprovalSkipped(t *testing.T, state *State, nodeID string) {
	t.Helper()
	count := 0
	for _, task := range state.Tasks {
		if task.NodeID == nodeID {
			t.Fatalf("skipped node has task: %#v", task)
		}
	}
	for _, event := range state.History {
		if event.NodeID != nodeID {
			continue
		}
		if event.Type != HistoryEventType("node_skipped") || event.ActorID != "" || event.TaskID != "" || !strings.Contains(event.Message, "未解析到审批人") {
			t.Fatalf("invalid skip audit: %#v", event)
		}
		count++
	}
	if count != 1 {
		t.Fatalf("skip history count = %d, want 1: %#v", count, state.History)
	}
	for _, intent := range state.NotificationIntents {
		if intent.NodeID == nodeID {
			t.Fatalf("skipped node emitted notification: %#v", intent)
		}
	}
}

func TestApprovalSkipEmptyCandidates(t *testing.T) {
	for _, mode := range []string{workflowcore.ApprovalModeSingle, workflowcore.ApprovalModeSequential, workflowcore.ApprovalModeParallel, workflowcore.ApprovalModeCountersign} {
		for _, enabled := range []bool{false, true} {
			for _, downstream := range []bool{false, true} {
				for _, users := range [][]string{nil, {}, {"", " \t", "\n"}} {
					t.Run(fmt.Sprintf("%s/enabled=%t/downstream=%t/users=%q", mode, enabled, downstream, users), func(t *testing.T) {
						definition := skipTestDefinition(t, mode, enabled, downstream)
						engine := newTestEngine(skipTestResolver{users: map[string][]string{"approve": users, "next": {"next-user"}}})
						state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
						if !enabled {
							if !errors.Is(err, ErrAssigneeUnavailable) || state != nil {
								t.Fatalf("default must fail: state=%#v err=%v", state, err)
							}
							return
						}
						if err != nil {
							t.Fatal(err)
						}
						assertApprovalSkipped(t, state, "approve")
						if downstream {
							pending := state.PendingTasks()
							if state.Instance.Status != InstanceStatusRunning || len(pending) != 1 || pending[0].NodeID != "next" || pending[0].AssigneeID != "next-user" {
								t.Fatalf("downstream tasks = %#v", state)
							}
							if len(state.NotificationIntents) != 1 || state.NotificationIntents[0].Kind != NotificationKindTaskArrived {
								t.Fatalf("downstream arrival = %#v", state.NotificationIntents)
							}
							completePendingTask(t, engine, definition, state, "next-user", TaskActionApprove)
						}
						if state.Instance.Status != InstanceStatusCompleted || countHistoryType(state.History, HistoryInstanceCompleted) != 1 {
							t.Fatalf("workflow did not complete exactly once: %#v", state)
						}
					})
				}
			}
		}
	}
}

func TestApprovalSkipNonemptyCandidatesStillRequireApproval(t *testing.T) {
	for _, mode := range []string{workflowcore.ApprovalModeSingle, workflowcore.ApprovalModeSequential, workflowcore.ApprovalModeParallel, workflowcore.ApprovalModeCountersign} {
		t.Run(mode, func(t *testing.T) {
			definition := skipTestDefinition(t, mode, true, false)
			engine := newTestEngine(skipTestResolver{users: map[string][]string{"approve": {"", " u1 ", "u1", "\t", "u2"}}})
			state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
			if err != nil {
				t.Fatal(err)
			}
			wantTasks, wantPending := 2, 2
			if mode == workflowcore.ApprovalModeSingle {
				wantTasks, wantPending = 1, 1
			} else if mode == workflowcore.ApprovalModeSequential {
				wantPending = 1
			}
			if state.Instance.Status != InstanceStatusRunning || len(state.Tasks) != wantTasks || len(state.PendingTasks()) != wantPending || len(state.NotificationIntents) != wantPending || state.Tasks[0].AssigneeID != "u1" {
				t.Fatalf("normal approval changed: %#v", state)
			}
			if hasHistoryType(state.History, HistoryEventType("node_skipped")) {
				t.Fatal("nonempty approval skipped")
			}
			for len(state.PendingTasks()) > 0 {
				task := state.PendingTasks()[0]
				if err := engine.Complete(context.Background(), definition, state, CompleteRequest{TaskID: task.ID, ActorID: task.AssigneeID, Action: TaskActionApprove}); err != nil {
					t.Fatal(err)
				}
			}
			if state.Instance.Status != InstanceStatusCompleted {
				t.Fatalf("status = %s", state.Instance.Status)
			}
		})
	}
}

func TestApprovalSkipPropagatesResolverErrors(t *testing.T) {
	for _, cause := range []error{ErrAssigneeUnavailable, fmt.Errorf("wrapped: %w", ErrAssigneeUnavailable), errors.New("database unavailable"), context.Canceled} {
		for _, users := range [][]string{nil, {"u1"}} {
			definition := skipTestDefinition(t, workflowcore.ApprovalModeSingle, true, false)
			engine := newTestEngine(skipTestResolver{users: map[string][]string{"approve": users}, err: cause})
			state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
			if !errors.Is(err, cause) || state != nil {
				t.Fatalf("resolver error swallowed: state=%#v err=%v, want %v", state, err, cause)
			}
		}
	}
}

func TestApprovalSkipRejectsInvalidConfiguration(t *testing.T) {
	for _, mutate := range []func(*workflowcore.Node){
		func(node *workflowcore.Node) { node.Assignee = nil },
		func(node *workflowcore.Node) {
			node.Assignee = &workflowcore.Assignee{Type: workflowcore.AssigneeTypeUser, Value: " "}
		},
		func(node *workflowcore.Node) { node.Assignee.Type = "invalid" },
		func(node *workflowcore.Node) { node.ApprovalMode = "invalid" },
		func(node *workflowcore.Node) { node.CompletionRate = 0 },
		func(node *workflowcore.Node) { node.Type = workflowcore.NodeTypeHandle },
	} {
		definition := skipTestDefinition(t, workflowcore.ApprovalModeCountersign, true, false)
		mutate(&definition.Nodes[1])
		state, err := newTestEngine(skipTestResolver{}).Start(context.Background(), definition, StartRequest{StarterID: "starter"})
		var validationErrors workflowcore.ValidationErrors
		if !errors.As(err, &validationErrors) || state != nil {
			t.Fatalf("invalid config bypassed: state=%#v err=%v", state, err)
		}
	}
}

func TestApprovalSkipLayeredResolution(t *testing.T) {
	for _, mode := range []string{workflowcore.ApprovalModeSingle, workflowcore.ApprovalModeSequential, workflowcore.ApprovalModeParallel, workflowcore.ApprovalModeCountersign} {
		for _, policy := range []string{workflowcore.DepartmentApprovalChainMissingSkip, workflowcore.DepartmentApprovalChainMissingError} {
			for _, enabled := range []bool{false, true} {
				for _, test := range []struct {
					name   string
					layers []ApprovalLayer
					err    error
				}{
					{name: "nil layers"},
					{name: "all layers empty", layers: []ApprovalLayer{{DepartmentID: 1}, {DepartmentID: 2, AssigneeIDs: []string{"", " \t"}}}},
					{name: "mixed layers", layers: []ApprovalLayer{{DepartmentID: 1}, {DepartmentID: 2, AssigneeIDs: []string{" u1 ", "u1"}}, {DepartmentID: 3, AssigneeIDs: []string{"u2"}}}},
					{name: "missing policy error", err: errors.New("部门测试部未配置可用的负责人")},
					{name: "typed resolver error", err: ErrAssigneeUnavailable},
					{name: "database error after partial result", layers: []ApprovalLayer{{AssigneeIDs: []string{"u1"}}}, err: errors.New("database unavailable")},
				} {
					t.Run(fmt.Sprintf("%s/%s/enabled=%t/%s", mode, policy, enabled, test.name), func(t *testing.T) {
						definition := skipTestDefinition(t, mode, enabled, false)
						node := &definition.Nodes[1]
						node.Assignee = &workflowcore.Assignee{Type: workflowcore.AssigneeTypeOrgIdentity, Value: "starter_department:department_leader"}
						node.DepartmentApprovalChain = &workflowcore.DepartmentApprovalChainConfig{Enabled: true, StopMode: workflowcore.DepartmentApprovalChainStopRoot, MissingAssigneePolicy: policy}
						engine := newTestEngine(skipTestLayerResolver{skipTestResolver: skipTestResolver{err: errors.New("flat resolver must not be called")}, layers: test.layers, err: test.err})
						state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
						if test.err != nil {
							if !errors.Is(err, test.err) || state != nil {
								t.Fatalf("layer resolver error swallowed: %v", err)
							}
							return
						}
						if test.name != "mixed layers" {
							if !enabled {
								if !errors.Is(err, ErrAssigneeUnavailable) {
									t.Fatalf("disabled skip error = %v", err)
								}
								return
							}
							if err != nil {
								t.Fatal(err)
							}
							assertApprovalSkipped(t, state, "approve")
							if state.Instance.Status != InstanceStatusCompleted {
								t.Fatalf("empty chain did not complete: %#v", state)
							}
							return
						}
						if err != nil {
							t.Fatal(err)
						}
						if len(state.Tasks) != 2 || len(state.PendingTasks()) != 1 || state.Tasks[0].DepartmentID != 2 || state.Tasks[1].Status != TaskStatusWaiting || hasHistoryType(state.History, HistoryEventType("node_skipped")) {
							t.Fatalf("nonempty chain changed: %#v", state)
						}
						completePendingTask(t, engine, definition, state, "u1", TaskActionApprove)
						if state.PendingTasks()[0].DepartmentID != 3 {
							t.Fatal("next layer not activated")
						}
						completePendingTask(t, engine, definition, state, "u2", TaskActionApprove)
						if state.Instance.Status != InstanceStatusCompleted {
							t.Fatal("chain not completed")
						}
					})
				}
			}
		}
	}
}

func TestApprovalSkipRequiresLayerResolverAndValidChain(t *testing.T) {
	definition := skipTestDefinition(t, workflowcore.ApprovalModeSingle, true, false)
	definition.Nodes[1].Assignee = &workflowcore.Assignee{Type: workflowcore.AssigneeTypeOrgIdentity, Value: "starter_department:department_leader"}
	definition.Nodes[1].DepartmentApprovalChain = &workflowcore.DepartmentApprovalChainConfig{Enabled: true, StopMode: workflowcore.DepartmentApprovalChainStopRoot, MissingAssigneePolicy: workflowcore.DepartmentApprovalChainMissingSkip}
	if _, err := newTestEngine(skipTestResolver{}).Start(context.Background(), definition, StartRequest{StarterID: "starter"}); err == nil || !strings.Contains(err.Error(), "不支持逐级部门审批") {
		t.Fatalf("unsupported resolver bypassed: %v", err)
	}
	definition.Nodes[1].DepartmentApprovalChain.MissingAssigneePolicy = "invalid"
	_, err := newTestEngine(skipTestLayerResolver{}).Start(context.Background(), definition, StartRequest{StarterID: "starter"})
	var issues workflowcore.ValidationErrors
	if !errors.As(err, &issues) {
		t.Fatalf("invalid chain bypassed: %v", err)
	}
}

func TestApprovalSkipParallelJoin(t *testing.T) {
	for _, rightUsers := range [][]string{nil, {"right-user"}} {
		definition := skipTestDefinition(t, workflowcore.ApprovalModeSingle, true, false)
		right := approvalNode("right", workflowcore.ApprovalModeSingle, 100)
		enableApprovalSkip(t, &right)
		definition.Nodes = append(definition.Nodes, right,
			workflowcore.Node{ID: "split", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeSplit},
			workflowcore.Node{ID: "join", Type: workflowcore.NodeTypeParallel, GatewayMode: workflowcore.GatewayModeJoin})
		definition.Edges = []workflowcore.Edge{
			{ID: "e1", Source: "start", Target: "split"},
			{ID: "e2", Source: "split", Target: "approve"},
			{ID: "e3", Source: "split", Target: "right"},
			{ID: "e4", Source: "approve", Target: "join"},
			{ID: "e5", Source: "right", Target: "join"},
			{ID: "e6", Source: "join", Target: "end"},
		}
		engine := newTestEngine(skipTestResolver{users: map[string][]string{"right": rightUsers}})
		state, err := engine.Start(context.Background(), definition, StartRequest{StarterID: "starter"})
		if err != nil {
			t.Fatal(err)
		}
		assertApprovalSkipped(t, state, "approve")
		if len(rightUsers) > 0 {
			if state.Instance.Status != InstanceStatusRunning || len(state.PendingTasks()) != 1 {
				t.Fatal("join must wait for nonempty branch")
			}
			completePendingTask(t, engine, definition, state, "right-user", TaskActionApprove)
		} else {
			assertApprovalSkipped(t, state, "right")
		}
		if state.Instance.Status != InstanceStatusCompleted || countHistoryType(state.History, HistoryInstanceCompleted) != 1 {
			t.Fatalf("join completion = %#v", state)
		}
	}
}

func TestApprovalSkipPreservesAutomaticTransitionLimit(t *testing.T) {
	definition := skipTestDefinition(t, workflowcore.ApprovalModeSingle, true, false)
	// Enter at the existing budget boundary to verify skipping carries the budget forward.
	state := &State{Instance: ProcessInstance{Status: InstanceStatusRunning}, Tokens: []Token{{ID: "token", NodeID: "approve", Status: TokenStatusActive}}}
	engine := newTestEngine(skipTestResolver{})
	if err := engine.advanceToken(context.Background(), definition, state, 0, maxAutomaticTransitions); !errors.Is(err, ErrTransitionLimit) {
		t.Fatalf("transition budget bypassed: %v", err)
	}
	assertApprovalSkipped(t, state, "approve")
}
