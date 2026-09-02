package infrastructure

import (
	"encoding/json"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestTaskModelRoundTripPreservesRuntimeStatus(t *testing.T) {
	statuses := []workflowdomain.TaskStatus{
		workflowdomain.TaskStatusWaiting,
		workflowdomain.TaskStatusPending,
		workflowdomain.TaskStatusCompleted,
		workflowdomain.TaskStatusApproved,
		workflowdomain.TaskStatusRejected,
		workflowdomain.TaskStatusCancelled,
	}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			task := workflowdomain.Task{
				ID: "task-1", TokenID: "token-1", NodeID: "approve", NodeName: "审批",
				GroupKey: "group-1", AssigneeID: "42", ApprovalMode: "countersign",
				CompletionRate: 67, Sequence: 2, Total: 3, Status: status,
				Action: workflowdomain.TaskActionApprove, Comment: "同意",
			}
			model := taskToModel("instance-1", task, 1234, nil)
			actual := taskFromModel(model)

			if actual.Status != status {
				t.Fatalf("status = %q, want %q", actual.Status, status)
			}
			if actual.CompletionRate != 67 {
				t.Fatalf("completion rate = %d, want 67", actual.CompletionRate)
			}
		})
	}
}

func TestCompletedHandleTaskPersistsSubmitActorAndTime(t *testing.T) {
	task := workflowdomain.Task{
		ID: "task-1", TokenID: "token-1", NodeID: "handle", NodeName: "填写资料",
		AssigneeID: "42", ApprovalMode: workflowcore.ApprovalModeSingle,
		Status: workflowdomain.TaskStatusCompleted, Action: workflowdomain.TaskActionSubmit,
	}
	history := []workflowdomain.HistoryEvent{{Type: workflowdomain.HistoryTaskSubmitted, TaskID: task.ID, ActorID: "42"}}
	model := taskToModel("instance-1", task, 1234, handledActors(history))

	if model.HandledBy != "42" || model.HandledAt != 1234 || model.Action != "submit" || model.Status != "completed" {
		t.Fatalf("completed handle task model = %#v", model)
	}
}

func TestDefinitionNodeTypesIncludesExtendedNodes(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{
		{ID: "start", Type: workflowcore.NodeTypeStart},
		{ID: "handle", Type: workflowcore.NodeTypeHandle},
		{ID: "timer", Type: workflowcore.NodeTypeTimer},
	}}
	nodeTypes := definitionNodeTypes(definition)
	if nodeTypes["handle"] != workflowcore.NodeTypeHandle || nodeTypes["timer"] != workflowcore.NodeTypeTimer {
		t.Fatalf("node types = %#v", nodeTypes)
	}
}

func TestDefinitionInitiatorPreservesDepartmentScopeWithDefensiveCopies(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{{
		ID: "start", Type: workflowcore.NodeTypeStart,
		Initiator: &workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8}, DepartmentIDs: []uint{3, 5},
		},
	}}}

	initiator := definitionInitiator(definition)
	definition.Nodes[0].Initiator.UserIDs[0] = 99
	definition.Nodes[0].Initiator.DepartmentIDs[0] = 99

	if len(initiator.UserIDs) != 2 || initiator.UserIDs[0] != 7 || initiator.UserIDs[1] != 8 {
		t.Fatalf("initiator users = %#v", initiator.UserIDs)
	}
	if len(initiator.DepartmentIDs) != 2 || initiator.DepartmentIDs[0] != 3 || initiator.DepartmentIDs[1] != 5 {
		t.Fatalf("initiator departments = %#v", initiator.DepartmentIDs)
	}
}

func TestDefinitionStartAvailabilityPreservesRecurringListsWithDefensiveCopies(t *testing.T) {
	definition := workflowcore.Definition{Nodes: []workflowcore.Node{{
		ID: "start", Type: workflowcore.NodeTypeStart,
		Availability: &workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityMonthly, Timezone: "Asia/Shanghai",
			MonthDays: []int{1, 15}, LastDayOfMonth: true, DailyStartTime: "09:00", DailyEndTime: "18:00",
		},
	}}}

	availability := definitionStartAvailability(definition)
	definition.Nodes[0].Availability.MonthDays[0] = 30

	if availability.Mode != workflowcore.StartAvailabilityMonthly || len(availability.MonthDays) != 2 || availability.MonthDays[0] != 1 || !availability.LastDayOfMonth {
		t.Fatalf("start availability = %#v", availability)
	}
}

func TestDecodeVariableUsesJSONNumber(t *testing.T) {
	value, err := decodeVariable(`{"count":9007199254740993,"ratio":1.5}`)
	if err != nil {
		t.Fatalf("decodeVariable() error = %v", err)
	}
	object, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("decoded type = %T", value)
	}
	if object["count"] != json.Number("9007199254740993") {
		t.Fatalf("count = %#v, want json.Number", object["count"])
	}
}

func TestStateFromModelsInitializesEmptyVariables(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{ID: "instance-1", StarterID: "7", OperatorID: "42", Status: string(workflowdomain.InstanceStatusRunning)},
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if state.Variables == nil {
		t.Fatal("variables must be initialized")
	}
	if state.FormData == nil {
		t.Fatal("form data must be initialized")
	}
	if state.Instance.StarterID != "7" || state.Instance.OperatorID != "42" {
		t.Fatalf("instance identities = starter %q operator %q", state.Instance.StarterID, state.Instance.OperatorID)
	}
}

func TestStateFromModelsRestoresFormData(t *testing.T) {
	state, err := stateFromModels(
		workflowmodel.ProcessInstance{
			ID: "instance-1", Status: string(workflowdomain.InstanceStatusRunning),
			FormDataJSON: `{"reason":"出差","days":2}`,
		},
		nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("stateFromModels() error = %v", err)
	}
	if state.FormData["reason"] != "出差" || state.FormData["days"] != json.Number("2") {
		t.Fatalf("form data = %#v", state.FormData)
	}
}
