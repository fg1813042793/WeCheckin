package httpclient

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

type runtimeServiceStub struct {
	startRequest    workflowapp.StartInstanceRequest
	completeRequest workflowapp.CompleteTaskRequest
	withdrawRequest workflowapp.WithdrawInstanceRequest
	instanceQuery   workflowapp.InstanceQuery
	taskQuery       workflowapp.TaskQuery
	actorID         string
	instanceID      string
	startCalls      int
	completeCalls   int
	withdrawCalls   int
}

func (stub *runtimeServiceStub) ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error) {
	return []workflowapp.PublishedDefinition{}, nil
}

func (stub *runtimeServiceStub) GetPublishedDefinition(_ context.Context, definitionID uint) (*workflowapp.PublishedDefinition, error) {
	return &workflowapp.PublishedDefinition{ID: definitionID}, nil
}

func (stub *runtimeServiceStub) StartInstance(_ context.Context, request workflowapp.StartInstanceRequest) (*workflowdomain.State, error) {
	stub.startCalls++
	stub.startRequest = request
	return &workflowdomain.State{}, nil
}

func (stub *runtimeServiceStub) CompleteTask(_ context.Context, request workflowapp.CompleteTaskRequest) (*workflowdomain.State, error) {
	stub.completeCalls++
	stub.completeRequest = request
	return &workflowdomain.State{}, nil
}

func (stub *runtimeServiceStub) WithdrawInstance(_ context.Context, request workflowapp.WithdrawInstanceRequest) (*workflowdomain.State, error) {
	stub.withdrawCalls++
	stub.withdrawRequest = request
	return &workflowdomain.State{}, nil
}

func (stub *runtimeServiceStub) ListMyInstances(_ context.Context, actorID string, query workflowapp.InstanceQuery) (*workflowapp.InstanceList, error) {
	stub.actorID = actorID
	stub.instanceQuery = query
	return &workflowapp.InstanceList{}, nil
}

func (stub *runtimeServiceStub) GetMyInstance(_ context.Context, actorID, instanceID string) (*workflowapp.InstanceDetail, error) {
	stub.actorID = actorID
	stub.instanceID = instanceID
	return &workflowapp.InstanceDetail{}, nil
}

func (stub *runtimeServiceStub) ListMyTasks(_ context.Context, actorID string, query workflowapp.TaskQuery) (*workflowapp.TaskList, error) {
	stub.actorID = actorID
	stub.taskQuery = query
	return &workflowapp.TaskList{}, nil
}

func TestStartInstanceUsesAuthenticatedUserAndKeepsFormData(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Request.SetBodyString(`{
		"definitionId": 7,
		"businessType": "leave",
		"businessKey": "LEAVE-1001",
		"starterId": "forged-user",
		"variables": {"days": 2},
		"formData": {"reason": "annual leave"}
	}`)

	handler.StartInstance(context.Background(), c)

	if stub.startCalls != 1 {
		t.Fatalf("expected one start call, got %d", stub.startCalls)
	}
	if stub.startRequest.StarterID != "42" {
		t.Fatalf("starter must come from authenticated user, got %q", stub.startRequest.StarterID)
	}
	if stub.startRequest.FormData["reason"] != "annual leave" {
		t.Fatalf("unexpected form data: %+v", stub.startRequest.FormData)
	}
}

func TestCompleteTaskUsesRouteTaskAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "task-1"})
	c.Request.SetBodyString(`{
		"taskId": "forged-task",
		"actorId": "forged-user",
		"action": "approve",
		"formData": {"managerComment": "approved"}
	}`)

	handler.CompleteTask(context.Background(), c)

	if stub.completeCalls != 1 {
		t.Fatalf("expected one complete call, got %d", stub.completeCalls)
	}
	if stub.completeRequest.TaskID != "task-1" || stub.completeRequest.ActorID != "42" {
		t.Fatalf("route task and authenticated actor must win: %+v", stub.completeRequest)
	}
}

func TestMyQueriesIgnoreForgedActorFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Request.SetRequestURI("/api/v2/workflows/instances?starterId=999&status=running&page=2&pageSize=10")

	handler.ListMyInstances(context.Background(), c)

	if stub.actorID != "42" || stub.instanceQuery.StarterID != "" {
		t.Fatalf("instance actor must be supplied separately by auth: actor=%q query=%+v", stub.actorID, stub.instanceQuery)
	}

	c = newUserContext(42)
	c.Request.SetRequestURI("/api/v2/workflows/tasks?assigneeId=999&status=pending&page=3&pageSize=15")
	handler.ListMyTasks(context.Background(), c)
	if stub.actorID != "42" || stub.taskQuery.AssigneeID != "" {
		t.Fatalf("task actor must be supplied separately by auth: actor=%q query=%+v", stub.actorID, stub.taskQuery)
	}
}

func TestWithdrawUsesRouteInstanceAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"instanceId":"forged-instance","actorId":"forged-user","reason":"mistake"}`)

	handler.WithdrawInstance(context.Background(), c)

	if stub.withdrawCalls != 1 {
		t.Fatalf("expected one withdraw call, got %d", stub.withdrawCalls)
	}
	if stub.withdrawRequest.InstanceID != "instance-1" || stub.withdrawRequest.ActorID != "42" {
		t.Fatalf("route instance and authenticated actor must win: %+v", stub.withdrawRequest)
	}
}

func newUserContext(userID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("user", &model.User{ID: userID})
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	return c
}
