package httpadmin

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/workflow/adminscope"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

type scopeResolverStub struct {
	scope adminscope.Scope
	err   error
}

func (stub scopeResolverStub) Resolve(context.Context, *model.Admin) (adminscope.Scope, error) {
	return stub.scope, stub.err
}

func TestScopedHandlerFiltersDefinitionAndManagementLists(t *testing.T) {
	stub := &runtimeServiceStub{}
	visibility := workflowapp.InstanceVisibility{Ready: true, DepartmentIDs: []uint{10, 11}}
	handler := NewRuntimeHandlerWithAccess(stub, scopeResolverStub{scope: adminscope.Scope{
		Visibility:    visibility,
		DefinitionIDs: []uint{7},
	}})

	definitions := newAdminContext(42)
	handler.ListDefinitions(context.Background(), definitions)
	if !strings.Contains(string(definitions.Response.Body()), `"id":7`) {
		t.Fatalf("visible definition missing: %s", definitions.Response.Body())
	}

	instances := newAdminContext(42)
	handler.ListInstances(context.Background(), instances)
	if stub.instanceQuery.Visibility == nil || len(stub.instanceQuery.Visibility.DepartmentIDs) != 2 {
		t.Fatalf("instance visibility = %#v", stub.instanceQuery.Visibility)
	}

	tasks := newAdminContext(42)
	handler.ListTasks(context.Background(), tasks)
	if stub.taskQuery.Visibility == nil || len(stub.taskQuery.Visibility.DepartmentIDs) != 2 {
		t.Fatalf("task visibility = %#v", stub.taskQuery.Visibility)
	}

	notifications := newAdminContext(42)
	handler.ListNotifications(context.Background(), notifications)
	if stub.notificationQuery.Visibility == nil || len(stub.notificationQuery.Visibility.DepartmentIDs) != 2 {
		t.Fatalf("notification visibility = %#v", stub.notificationQuery.Visibility)
	}
}

func TestScopedHandlerRejectsInvisibleDefinitionBeforeLoadingOrStarting(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandlerWithAccess(stub, scopeResolverStub{scope: adminscope.Scope{
		Visibility:    workflowapp.InstanceVisibility{Ready: true, UserIDs: []uint{42}},
		DefinitionIDs: []uint{8},
	}})

	detail := newAdminContext(42)
	detail.Params = append(detail.Params, param.Param{Key: "id", Value: "7"})
	handler.GetDefinition(context.Background(), detail)
	if stub.getDefinition != 0 {
		t.Fatalf("invisible definition loaded %d times", stub.getDefinition)
	}

	start := newAdminContext(42)
	start.Request.SetBodyString(`{"definitionId":7,"businessType":"leave","businessKey":"leave-1","starterId":"42"}`)
	handler.StartInstance(context.Background(), start)
	if stub.startCalls != 0 {
		t.Fatalf("invisible definition started %d times", stub.startCalls)
	}
}

func TestScopedHandlerRejectsInvisibleInstanceBeforeMutation(t *testing.T) {
	stub := &runtimeServiceStub{instanceList: &workflowapp.InstanceList{List: []workflowapp.InstanceSummary{}}}
	handler := NewRuntimeHandlerWithAccess(stub, scopeResolverStub{scope: adminscope.Scope{
		Visibility: workflowapp.InstanceVisibility{Ready: true, DepartmentIDs: []uint{10}},
	}})

	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"reason":"invalid"}`)
	handler.CancelInstance(context.Background(), c)
	if stub.cancelCalls != 0 {
		t.Fatalf("invisible instance cancel calls = %d", stub.cancelCalls)
	}
}

func TestAdminHandlerWithoutScopeResolverDeniesRequests(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := &RuntimeHandler{service: stub}
	c := newAdminContext(42)

	handler.ListInstances(context.Background(), c)

	if stub.instanceQuery.Visibility != nil {
		t.Fatalf("service called without scope resolver: %#v", stub.instanceQuery)
	}
	if !strings.Contains(string(c.Response.Body()), "流程操作失败") {
		t.Fatalf("unexpected missing-resolver response: %s", c.Response.Body())
	}
}
