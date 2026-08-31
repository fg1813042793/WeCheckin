package httpadmin

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
	cancelRequest   workflowapp.CancelInstanceRequest
	instanceQuery   workflowapp.InstanceQuery
	taskQuery       workflowapp.TaskQuery
	startCalls      int
	completeCalls   int
	cancelCalls     int
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

func (stub *runtimeServiceStub) CancelInstance(_ context.Context, request workflowapp.CancelInstanceRequest) (*workflowdomain.State, error) {
	stub.cancelCalls++
	stub.cancelRequest = request
	return &workflowdomain.State{}, nil
}

func (stub *runtimeServiceStub) ListInstances(_ context.Context, query workflowapp.InstanceQuery) (*workflowapp.InstanceList, error) {
	stub.instanceQuery = query
	return &workflowapp.InstanceList{Page: query.Page, PageSize: query.PageSize}, nil
}

func (stub *runtimeServiceStub) GetInstance(_ context.Context, instanceID string) (*workflowapp.InstanceDetail, error) {
	return &workflowapp.InstanceDetail{Instance: workflowapp.InstanceSummary{ID: instanceID}}, nil
}

func (stub *runtimeServiceStub) ListTasks(_ context.Context, query workflowapp.TaskQuery) (*workflowapp.TaskList, error) {
	stub.taskQuery = query
	return &workflowapp.TaskList{Page: query.Page, PageSize: query.PageSize}, nil
}

func TestStartInstanceUsesAuthenticatedAdminAsStarter(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Request.SetBodyString(`{
		"definitionId": 7,
		"definitionVersion": 3,
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
		t.Fatalf("starter must come from authenticated admin, got %q", stub.startRequest.StarterID)
	}
	if stub.startRequest.BusinessType != "leave" || stub.startRequest.BusinessKey != "LEAVE-1001" {
		t.Fatalf("unexpected business reference: %+v", stub.startRequest)
	}
	if stub.startRequest.FormData["reason"] != "annual leave" {
		t.Fatalf("unexpected form data: %+v", stub.startRequest.FormData)
	}
}

func TestCompleteTaskUsesAuthenticatedAdminAndRouteTaskID(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "task-1001"})
	c.Request.SetBodyString(`{
		"taskId": "forged-task",
		"actorId": "forged-user",
		"action": "approve",
		"comment": "同意",
		"variables": {"approvedAmount": 100}
	}`)

	handler.CompleteTask(context.Background(), c)

	if stub.completeCalls != 1 {
		t.Fatalf("expected one complete call, got %d", stub.completeCalls)
	}
	if stub.completeRequest.TaskID != "task-1001" {
		t.Fatalf("task id must come from route, got %q", stub.completeRequest.TaskID)
	}
	if stub.completeRequest.ActorID != "42" {
		t.Fatalf("actor must come from authenticated admin, got %q", stub.completeRequest.ActorID)
	}
}

func TestCancelInstanceUsesAuthenticatedAdminAndRouteInstanceID(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1001"})
	c.Request.SetBodyString(`{"instanceId":"forged-instance","actorId":"forged-user","reason":"invalid request"}`)

	handler.CancelInstance(context.Background(), c)

	if stub.cancelCalls != 1 {
		t.Fatalf("expected one cancel call, got %d", stub.cancelCalls)
	}
	if stub.cancelRequest.InstanceID != "instance-1001" || stub.cancelRequest.ActorID != "42" {
		t.Fatalf("route instance and authenticated actor must win: %+v", stub.cancelRequest)
	}
}

func TestListInstancesParsesStableFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	c.Request.SetRequestURI("/api/v2/admin/workflow-instances?definitionId=7&status=running&businessType=leave&businessKey=LEAVE-1001&starterId=42&page=2&pageSize=40")

	handler.ListInstances(context.Background(), c)

	want := workflowapp.InstanceQuery{
		DefinitionID: 7,
		Status:       "running", BusinessType: "leave", BusinessKey: "LEAVE-1001", StarterID: "42",
		Page: 2, PageSize: 40,
	}
	if stub.instanceQuery != want {
		t.Fatalf("unexpected instance query: got %+v want %+v", stub.instanceQuery, want)
	}
}

func TestListTasksParsesStableFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	c.Request.SetRequestURI("/api/v2/admin/workflow-tasks?instanceId=instance-1&assigneeId=42&status=pending&page=3&pageSize=15")

	handler.ListTasks(context.Background(), c)

	want := workflowapp.TaskQuery{
		InstanceID: "instance-1", AssigneeID: "42", Status: "pending", Page: 3, PageSize: 15,
	}
	if stub.taskQuery != want {
		t.Fatalf("unexpected task query: got %+v want %+v", stub.taskQuery, want)
	}
}

func TestStartInstanceRejectsUnauthenticatedRequest(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	c.Request.SetBodyString(`{"definitionId":7,"businessType":"leave","businessKey":"LEAVE-1001"}`)

	handler.StartInstance(context.Background(), c)

	if stub.startCalls != 0 {
		t.Fatal("service must not run without authenticated admin")
	}
}

func newAdminContext(adminID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("admin", &model.Admin{ID: adminID})
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	return c
}
