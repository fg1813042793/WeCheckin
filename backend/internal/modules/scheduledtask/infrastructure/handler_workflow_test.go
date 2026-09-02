package infrastructure

import (
	"context"
	"testing"

	"wecheckin/backend/internal/modules/scheduledtask/application"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestWorkflowHandlerStartsWorkflowWithRunIdempotencyKey(t *testing.T) {
	starter := &fakeWorkflowStarter{states: []*workflowdomain.State{{Instance: workflowdomain.ProcessInstance{ID: "instance-1"}}}}
	handler := NewWorkflowHandler(starter)
	run := application.RunContext{
		RunID: "run-42",
		Task: application.TaskSnapshot{HandlerConfigJSON: `{
			"definitionId":9,"versionPolicy":"fixed_version","fixedVersion":3,
			"starterId":"7","formData":{"days":2},"variables":{"source":"timer"}
		}`},
	}
	result, err := handler.Execute(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	request := starter.requests[0]
	if request.BusinessType != "scheduled_task" || request.BusinessKey != "run-42" || !request.Idempotent {
		t.Fatalf("business idempotency = %#v", request)
	}
	if request.DefinitionID != 9 || request.DefinitionVersion != 3 || request.StarterID != "7" || request.OperatorID != "7" {
		t.Fatalf("workflow request = %#v", request)
	}
	if request.FormData["days"] != float64(2) || result.Summary != "instance-1" {
		t.Fatalf("form/result = %#v / %#v", request.FormData, result)
	}
}

func TestWorkflowHandlerStartsIndependentInstanceForEachStarter(t *testing.T) {
	starter := &fakeWorkflowStarter{states: []*workflowdomain.State{
		{Instance: workflowdomain.ProcessInstance{ID: "instance-7"}},
		{Instance: workflowdomain.ProcessInstance{ID: "instance-8"}},
	}}
	handler := NewWorkflowHandler(starter)
	run := application.RunContext{
		RunID: "run-42",
		Task: application.TaskSnapshot{HandlerConfigJSON: `{
			"definitionId":9,"starterIds":[7,8],"formData":{},"variables":{}
		}`},
	}
	result, err := handler.Execute(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	if len(starter.requests) != 2 {
		t.Fatalf("workflow requests = %d, want 2", len(starter.requests))
	}
	for index, expected := range []struct {
		starterID   string
		businessKey string
	}{
		{starterID: "7", businessKey: "run-42:user:7"},
		{starterID: "8", businessKey: "run-42:user:8"},
	} {
		request := starter.requests[index]
		if request.StarterID != expected.starterID || request.OperatorID != expected.starterID || request.BusinessKey != expected.businessKey || !request.Idempotent {
			t.Fatalf("workflow request %d = %#v", index, request)
		}
	}
	instanceIDs, ok := result.Data["instanceIds"].([]string)
	if !ok || len(instanceIDs) != 2 || instanceIDs[0] != "instance-7" || instanceIDs[1] != "instance-8" {
		t.Fatalf("workflow result = %#v", result)
	}
}

type fakeWorkflowStarter struct {
	requests []workflowapp.StartInstanceRequest
	states   []*workflowdomain.State
}

func (starter *fakeWorkflowStarter) StartInstance(_ context.Context, request workflowapp.StartInstanceRequest) (*workflowdomain.State, error) {
	starter.requests = append(starter.requests, request)
	return starter.states[len(starter.requests)-1], nil
}
