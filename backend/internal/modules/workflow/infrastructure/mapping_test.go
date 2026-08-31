package infrastructure

import (
	"encoding/json"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestTaskModelRoundTripPreservesRuntimeStatus(t *testing.T) {
	statuses := []workflowdomain.TaskStatus{
		workflowdomain.TaskStatusWaiting,
		workflowdomain.TaskStatusPending,
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
		workflowmodel.ProcessInstance{ID: "instance-1", Status: string(workflowdomain.InstanceStatusRunning)},
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
