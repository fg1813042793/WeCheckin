package httpadmin

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

type runtimeServiceStub struct {
	startRequest           workflowapp.StartInstanceRequest
	completeRequest        workflowapp.CompleteTaskRequest
	cancelRequest          workflowapp.CancelInstanceRequest
	resumeInstanceID       string
	resumeActorID          string
	instanceQuery          workflowapp.InstanceQuery
	taskQuery              workflowapp.TaskQuery
	definitionID           uint
	listDefinitions        int
	getDefinition          int
	startCalls             int
	completeCalls          int
	cancelCalls            int
	resumeCalls            int
	notificationQuery      workflowapp.NotificationQuery
	notificationListCalls  int
	retryNotificationID    string
	retryNotificationCalls int
	dispatchDueLimit       int
	dispatchDueCalls       int
	deleteActorID          string
	deleteInstanceIDs      []string
	deleteCalls            int
	deleteTaskActorID      string
	deleteTaskID           string
	deleteTaskCalls        int
	listDefinitionsErr     error
}

func (stub *runtimeServiceStub) ListPublishedDefinitions(context.Context) ([]workflowapp.PublishedDefinition, error) {
	stub.listDefinitions++
	if stub.listDefinitionsErr != nil {
		return nil, stub.listDefinitionsErr
	}
	return []workflowapp.PublishedDefinition{{
		ID: 7, Key: "leave", Name: "请假审批", Version: 3,
		Form: []workflowcore.FormField{{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea}},
		FieldPermissions: map[string][]workflowcore.FieldPermission{
			"start": {{Field: "reason", Access: workflowcore.FieldAccessWrite}},
		},
		StartNodeID: "start",
		Initiator: workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{9},
		},
		Availability: workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1, 3, 5}, DailyStartTime: "09:00", DailyEndTime: "18:00",
		},
		AvailabilityStatus: workflowcore.StartAvailabilityStateAvailable,
	}}, nil
}

func TestListDefinitionsDoesNotExposeUnknownServiceError(t *testing.T) {
	stub := &runtimeServiceStub{listDefinitionsErr: errors.New("SELECT password FROM admins: secret")}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)

	handler.ListDefinitions(context.Background(), c)

	body := string(c.Response.Body())
	if !strings.Contains(body, "流程操作失败，请稍后重试") || strings.Contains(body, "password") || strings.Contains(body, "secret") {
		t.Fatalf("unsafe workflow error response: %s", body)
	}
}

func (stub *runtimeServiceStub) GetPublishedDefinition(_ context.Context, definitionID uint) (*workflowapp.PublishedDefinition, error) {
	stub.getDefinition++
	stub.definitionID = definitionID
	return &workflowapp.PublishedDefinition{
		ID: definitionID, Key: "leave", Name: "请假审批", Version: 3,
		Form: []workflowcore.FormField{{Key: "reason", Label: "申请原因", Type: workflowcore.FormFieldTypeTextarea}},
		FieldPermissions: map[string][]workflowcore.FieldPermission{
			"start": {{Field: "reason", Access: workflowcore.FieldAccessWrite}},
		},
		StartNodeID: "start",
		Initiator: workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{9},
		},
		Availability: workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1, 3, 5}, DailyStartTime: "09:00", DailyEndTime: "18:00",
		},
		AvailabilityStatus: workflowcore.StartAvailabilityStateAvailable,
	}, nil
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

func (stub *runtimeServiceStub) ResumeTimers(_ context.Context, instanceID, actorID string) (*workflowdomain.State, int, error) {
	stub.resumeCalls++
	stub.resumeInstanceID = instanceID
	stub.resumeActorID = actorID
	return &workflowdomain.State{}, 0, nil
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

func (stub *runtimeServiceStub) ListNotifications(_ context.Context, query workflowapp.NotificationQuery) (*workflowapp.NotificationList, error) {
	stub.notificationListCalls++
	stub.notificationQuery = query
	return &workflowapp.NotificationList{Page: query.Page, PageSize: query.PageSize}, nil
}

func (stub *runtimeServiceStub) RetryNotification(_ context.Context, id string) error {
	stub.retryNotificationCalls++
	stub.retryNotificationID = id
	return nil
}

func (stub *runtimeServiceStub) DispatchDueNotifications(_ context.Context, limit int) (int, error) {
	stub.dispatchDueCalls++
	stub.dispatchDueLimit = limit
	return 3, nil
}

func (stub *runtimeServiceStub) DeleteInstances(_ context.Context, actorID string, instanceIDs []string) (int, error) {
	stub.deleteCalls++
	stub.deleteActorID = actorID
	stub.deleteInstanceIDs = append([]string(nil), instanceIDs...)
	return len(instanceIDs), nil
}

func (stub *runtimeServiceStub) DeleteTask(_ context.Context, actorID, taskID string) error {
	stub.deleteTaskCalls++
	stub.deleteTaskActorID = actorID
	stub.deleteTaskID = taskID
	return nil
}

func TestListDefinitionsReturnsPublishedFormSchemasForAdmins(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)

	handler.ListDefinitions(context.Background(), c)

	if stub.listDefinitions != 1 {
		t.Fatalf("expected published definitions to be loaded once, got %d", stub.listDefinitions)
	}
	body := string(c.Response.Body())
	for _, snippet := range []string{
		`"key":"leave"`, `"form"`, `"reason"`, `"fieldPermissions"`, `"startNodeId":"start"`,
		`"initiator":{"scope":"specified","userIds":[7],"departmentIds":[3,5],"excludedUserIds":[9]}`,
		`"availability":{"mode":"weekly","timezone":"Asia/Shanghai"`, `"availabilityStatus":"available"`,
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("published definition response missing %s: %s", snippet, body)
		}
	}
}

func TestGetDefinitionReturnsPublishedFormSchemaForAdmins(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})

	handler.GetDefinition(context.Background(), c)

	if stub.getDefinition != 1 || stub.definitionID != 7 {
		t.Fatalf("expected definition 7 to be loaded once, got calls=%d id=%d", stub.getDefinition, stub.definitionID)
	}
	body := string(c.Response.Body())
	if !strings.Contains(body, `"form"`) || !strings.Contains(body, `"reason"`) {
		t.Fatalf("published definition detail should include form schema, got %s", body)
	}
}

func TestStartInstanceUsesSelectedStarterAndAuthenticatedAdminAsOperator(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Request.SetBodyString(`{
		"definitionId": 7,
		"definitionVersion": 3,
		"businessType": "leave",
		"businessKey": "LEAVE-1001",
		"starterId": "7",
		"variables": {"days": 2},
		"formData": {"reason": "annual leave"}
	}`)

	handler.StartInstance(context.Background(), c)

	if stub.startCalls != 1 {
		t.Fatalf("expected one start call, got %d", stub.startCalls)
	}
	if stub.startRequest.StarterID != "7" {
		t.Fatalf("starter must come from selected business user, got %q", stub.startRequest.StarterID)
	}
	if stub.startRequest.OperatorID != "42" {
		t.Fatalf("operator must come from authenticated admin, got %q", stub.startRequest.OperatorID)
	}
	if !stub.startRequest.AdminInitiated {
		t.Fatal("admin start must enforce administrator data scope")
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
			"action": "return",
			"comment": "同意",
			"returnTargetNodeId": "draft",
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
	if stub.completeRequest.ReturnTargetNodeID != "draft" {
		t.Fatalf("return target node = %q, want draft", stub.completeRequest.ReturnTargetNodeID)
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

func TestResumeTimersUsesAuthenticatedAdminAndRouteInstanceID(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newAdminContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1001"})
	resumer, ok := interface{}(handler).(interface {
		ResumeTimers(context.Context, *app.RequestContext)
	})
	if !ok {
		t.Fatal("admin timer resume handler missing")
	}

	resumer.ResumeTimers(context.Background(), c)

	if stub.resumeCalls != 1 || stub.resumeInstanceID != "instance-1001" || stub.resumeActorID != "42" {
		t.Fatalf("resume must use route instance and authenticated actor: calls=%d instance=%q actor=%q", stub.resumeCalls, stub.resumeInstanceID, stub.resumeActorID)
	}
}

func TestDeleteInstancesUsesAuthenticatedAdminAndStableIDs(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)

	single := newAdminContext(42)
	single.Params = append(single.Params, param.Param{Key: "id", Value: "instance-route"})
	single.Request.SetBodyString(`{"ids":["forged"]}`)
	handler.DeleteInstance(context.Background(), single)
	if stub.deleteCalls != 1 || stub.deleteActorID != "42" || len(stub.deleteInstanceIDs) != 1 || stub.deleteInstanceIDs[0] != "instance-route" {
		t.Fatalf("single delete = calls %d actor %q ids %#v", stub.deleteCalls, stub.deleteActorID, stub.deleteInstanceIDs)
	}

	batch := newAdminContext(43)
	batch.Request.SetBodyString(`{"ids":["instance-1","instance-2"]}`)
	handler.DeleteInstances(context.Background(), batch)
	if stub.deleteCalls != 2 || stub.deleteActorID != "43" || strings.Join(stub.deleteInstanceIDs, ",") != "instance-1,instance-2" {
		t.Fatalf("batch delete = calls %d actor %q ids %#v", stub.deleteCalls, stub.deleteActorID, stub.deleteInstanceIDs)
	}
	if !strings.Contains(string(batch.Response.Body()), `"deleted":2`) {
		t.Fatalf("batch delete response = %s", batch.Response.Body())
	}

	unauthenticated := app.NewContext(1)
	unauthenticated.Request.SetBodyString(`{"ids":["instance-3"]}`)
	handler.DeleteInstances(context.Background(), unauthenticated)
	if stub.deleteCalls != 2 {
		t.Fatalf("unauthenticated delete calls = %d", stub.deleteCalls)
	}
}

func TestDeleteTaskUsesAuthenticatedAdminAndRouteTaskID(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)

	request := newAdminContext(42)
	request.Params = append(request.Params, param.Param{Key: "id", Value: "task-route"})
	request.Request.SetBodyString(`{"id":"task-forged"}`)
	handler.DeleteTask(context.Background(), request)
	if stub.deleteTaskCalls != 1 || stub.deleteTaskActorID != "42" || stub.deleteTaskID != "task-route" {
		t.Fatalf("task delete = calls %d actor %q task %q", stub.deleteTaskCalls, stub.deleteTaskActorID, stub.deleteTaskID)
	}
	if !strings.Contains(string(request.Response.Body()), `"id":"task-route"`) {
		t.Fatalf("task delete response = %s", request.Response.Body())
	}

	unauthenticated := app.NewContext(1)
	unauthenticated.Params = append(unauthenticated.Params, param.Param{Key: "id", Value: "task-other"})
	handler.DeleteTask(context.Background(), unauthenticated)
	if stub.deleteTaskCalls != 1 {
		t.Fatalf("unauthenticated task delete calls = %d", stub.deleteTaskCalls)
	}
}

func TestNotificationManagementUsesAuthenticatedAdminAndStableInputs(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)

	listContext := newAdminContext(42)
	listContext.Request.SetRequestURI("/api/v2/admin/workflow-notifications?instanceId=instance-1&recipientUserId=7&kind=node_cc&channel=dingtalk_oa&status=failed&page=2&pageSize=15")
	handler.ListNotifications(context.Background(), listContext)
	query := stub.notificationQuery
	if stub.notificationListCalls != 1 || query.InstanceID != "instance-1" || query.RecipientUserID != "7" || query.Kind != "node_cc" || query.Channel != "dingtalk_oa" || query.Status != "failed" || query.Page != 2 || query.PageSize != 15 {
		t.Fatalf("notification query = %#v calls=%d", query, stub.notificationListCalls)
	}

	retryContext := newAdminContext(42)
	retryContext.Params = append(retryContext.Params, param.Param{Key: "id", Value: "outbox-route"})
	retryContext.Request.SetBodyString(`{"id":"outbox-forged"}`)
	handler.RetryNotification(context.Background(), retryContext)
	if stub.retryNotificationCalls != 1 || stub.retryNotificationID != "outbox-route" {
		t.Fatalf("retry notification = calls %d id %q", stub.retryNotificationCalls, stub.retryNotificationID)
	}

	dueContext := newAdminContext(42)
	dueContext.Request.SetBodyString(`{"limit":25}`)
	handler.DispatchDueNotifications(context.Background(), dueContext)
	if stub.dispatchDueCalls != 1 || stub.dispatchDueLimit != 25 {
		t.Fatalf("dispatch due = calls %d limit %d", stub.dispatchDueCalls, stub.dispatchDueLimit)
	}
	if !strings.Contains(string(dueContext.Response.Body()), `"dispatched":3`) {
		t.Fatalf("dispatch due response = %s", dueContext.Response.Body())
	}

	unauthenticated := app.NewContext(1)
	handler.ListNotifications(context.Background(), unauthenticated)
	if stub.notificationListCalls != 1 {
		t.Fatalf("unauthenticated notification list calls = %d", stub.notificationListCalls)
	}
}

func TestListInstancesParsesStableFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	c.Request.SetRequestURI("/api/v2/admin/workflow-instances?definitionId=7&definitionCategory=performance&status=running&businessType=leave&businessKey=LEAVE-1001&starterId=42&startTimeFrom=1000&startTimeTo=1999&endTimeFrom=2000&endTimeTo=2999&page=2&pageSize=40")

	handler.ListInstances(context.Background(), c)

	want := workflowapp.InstanceQuery{
		DefinitionID: 7, DefinitionCategory: "performance",
		Status: "running", BusinessType: "leave", BusinessKey: "LEAVE-1001", StarterID: "42",
		StartTimeFrom: 1000, StartTimeTo: 1999, EndTimeFrom: 2000, EndTimeTo: 2999,
		Page: 2, PageSize: 40,
	}
	if !reflect.DeepEqual(stub.instanceQuery, want) {
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
		InstanceID: "instance-1", AssigneeID: "42", Status: "pending", HideAdminDeleted: true, Page: 3, PageSize: 15,
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
