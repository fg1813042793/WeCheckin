package httpclient

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/workflowcore"
)

type runtimeServiceStub struct {
	startRequest     workflowapp.StartInstanceRequest
	completeRequest  workflowapp.CompleteTaskRequest
	withdrawRequest  workflowapp.WithdrawInstanceRequest
	commentRequest   workflowapp.CommentInstanceRequest
	remindRequest    workflowapp.RemindInstanceRequest
	instanceQuery    workflowapp.InstanceQuery
	taskQuery        workflowapp.TaskQuery
	actorID          string
	instanceID       string
	startCalls       int
	completeCalls    int
	withdrawCalls    int
	commentCalls     int
	remindCalls      int
	saveDraftRequest workflowapp.SaveStartDraftRequest
	draft            *workflowapp.StartDraft
	deleteDraftID    uint
	deleteDraftActor string
	deleteInstanceID string
	deleteActorID    string
	instanceDetail   *workflowapp.InstanceDetail
}

func (stub *runtimeServiceStub) ListPublishedDefinitionsForStarter(_ context.Context, actorID string) ([]workflowapp.PublishedDefinition, error) {
	stub.actorID = actorID
	return []workflowapp.PublishedDefinition{{
		ID: 7, Key: "leave", Name: "请假审批", Version: 3,
		Initiator: workflowcore.InitiatorConfig{
			Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{9},
		},
		Availability: workflowcore.StartAvailabilityConfig{
			Mode: workflowcore.StartAvailabilityMonthly, Timezone: "Asia/Shanghai", LastDayOfMonth: true, DailyStartTime: "09:00", DailyEndTime: "18:00",
		},
		AvailabilityStatus: workflowcore.StartAvailabilityStateOutsideWindow,
	}}, nil
}

func (stub *runtimeServiceStub) GetPublishedDefinitionForStarter(_ context.Context, definitionID uint, actorID string) (*workflowapp.PublishedDefinition, error) {
	stub.actorID = actorID
	return &workflowapp.PublishedDefinition{
		ID: definitionID,
		Nodes: []workflowapp.PublishedNode{{
			ID: "approval-1", Type: workflowcore.NodeTypeApproval, Name: "直属上级审批", AssigneeDisplay: "发起人的直属上级",
		}},
		Edges: []workflowapp.PublishedEdge{{ID: "flow-1", Source: "start", Target: "approval-1"}},
	}, nil
}

func (stub *runtimeServiceStub) StartInstance(_ context.Context, request workflowapp.StartInstanceRequest) (*workflowdomain.State, error) {
	stub.startCalls++
	stub.startRequest = request
	return &workflowdomain.State{}, nil
}

func (stub *runtimeServiceStub) GetStartDraft(_ context.Context, definitionID uint, actorID string) (*workflowapp.StartDraft, error) {
	stub.actorID = actorID
	stub.saveDraftRequest.DefinitionID = definitionID
	return stub.draft, nil
}

func (stub *runtimeServiceStub) SaveStartDraft(_ context.Context, request workflowapp.SaveStartDraftRequest) (*workflowapp.StartDraft, error) {
	stub.saveDraftRequest = request
	return &workflowapp.StartDraft{
		DefinitionID: request.DefinitionID, DefinitionVersion: request.DefinitionVersion,
		StarterID: request.StarterID, FormData: request.FormData,
	}, nil
}

func (stub *runtimeServiceStub) DeleteStartDraft(_ context.Context, definitionID uint, actorID string) error {
	stub.deleteDraftID = definitionID
	stub.deleteDraftActor = actorID
	return nil
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

func (stub *runtimeServiceStub) CommentInstance(_ context.Context, request workflowapp.CommentInstanceRequest) error {
	stub.commentCalls++
	stub.commentRequest = request
	return nil
}

func (stub *runtimeServiceStub) RemindInstance(_ context.Context, request workflowapp.RemindInstanceRequest) (*workflowapp.RemindInstanceResult, error) {
	stub.remindCalls++
	stub.remindRequest = request
	return &workflowapp.RemindInstanceResult{NodeID: request.NodeID, RemindedCount: 2}, nil
}

func (stub *runtimeServiceStub) DeleteMyInstance(_ context.Context, actorID, instanceID string) error {
	stub.deleteActorID = actorID
	stub.deleteInstanceID = instanceID
	return nil
}

func (stub *runtimeServiceStub) ListMyInstances(_ context.Context, actorID string, query workflowapp.InstanceQuery) (*workflowapp.InstanceList, error) {
	stub.actorID = actorID
	stub.instanceQuery = query
	return &workflowapp.InstanceList{}, nil
}

func (stub *runtimeServiceStub) GetMyInstance(_ context.Context, actorID, instanceID string) (*workflowapp.InstanceDetail, error) {
	stub.actorID = actorID
	stub.instanceID = instanceID
	if stub.instanceDetail != nil {
		return stub.instanceDetail, nil
	}
	return &workflowapp.InstanceDetail{}, nil
}

func (stub *runtimeServiceStub) ListMyTasks(_ context.Context, actorID string, query workflowapp.TaskQuery) (*workflowapp.TaskList, error) {
	stub.actorID = actorID
	stub.taskQuery = query
	return &workflowapp.TaskList{}, nil
}

func TestListDefinitionsReturnsInitiatorDepartmentsForClient(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)

	handler.ListDefinitions(context.Background(), c)

	body := string(c.Response.Body())
	if stub.actorID != "42" {
		t.Fatalf("definition list actor = %q, want 42", stub.actorID)
	}
	if !strings.Contains(body, `"initiator":{"scope":"specified","userIds":[7],"departmentIds":[3,5],"excludedUserIds":[9]}`) {
		t.Fatalf("published definition response missing initiator departments: %s", body)
	}
	if !strings.Contains(body, `"availability":{"mode":"monthly","timezone":"Asia/Shanghai"`) || !strings.Contains(body, `"availabilityStatus":"outside_window"`) {
		t.Fatalf("published definition response missing start availability: %s", body)
	}
}

func TestListDefinitionsUsesDingTalkH5AuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	dingtalkh5session.SetAuth(c, &model.DingTalkH5PerfUser{ID: 66}, "token")

	handler.ListDefinitions(context.Background(), c)

	if stub.actorID != "66" {
		t.Fatalf("DingTalk H5 definition list actor = %q, want 66", stub.actorID)
	}
}

func TestGetDefinitionReturnsPublishedGraphForDingTalkH5(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := app.NewContext(1)
	dingtalkh5session.SetAuth(c, &model.DingTalkH5PerfUser{ID: 66}, "token")
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})

	handler.GetDefinition(context.Background(), c)

	body := string(c.Response.Body())
	for _, snippet := range []string{
		`"nodes":[{"id":"approval-1"`,
		`"assigneeDisplay":"发起人的直属上级"`,
		`"edges":[{"id":"flow-1","source":"start","target":"approval-1"`,
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("published graph response missing %s: %s", snippet, body)
		}
	}
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
	if stub.startRequest.OperatorID != "42" {
		t.Fatalf("operator must come from authenticated user, got %q", stub.startRequest.OperatorID)
	}
	if stub.startRequest.AdminInitiated {
		t.Fatal("client self-start must not be marked as an administrator delegation")
	}
	if !stub.startRequest.ClearStartDraft {
		t.Fatal("client self-start must clear its saved start draft")
	}
	if stub.startRequest.FormData["reason"] != "annual leave" {
		t.Fatalf("unexpected form data: %+v", stub.startRequest.FormData)
	}
}

func TestGetInstanceReturnsStarterName(t *testing.T) {
	stub := &runtimeServiceStub{instanceDetail: &workflowapp.InstanceDetail{
		Instance: workflowapp.InstanceSummary{
			ID: "instance-1", StarterID: "7", StarterName: "张三",
		},
		Tasks: []workflowapp.TaskSummary{{
			ID: "task-1", AssigneeID: "8", AssigneeName: "李四", HandledBy: "9", HandledByName: "王五",
		}},
	}}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})

	handler.GetMyInstance(context.Background(), c)

	body := string(c.Response.Body())
	if !strings.Contains(body, `"starterName":"张三"`) {
		t.Fatalf("workflow instance response missing starter name: %s", body)
	}
	if !strings.Contains(body, `"assigneeName":"李四"`) || !strings.Contains(body, `"handledByName":"王五"`) {
		t.Fatalf("workflow instance response missing task user names: %s", body)
	}
}

func TestSaveStartDraftUsesRouteDefinitionAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "definitionId", Value: "7"})
	c.Request.SetBodyString(`{"definitionId":99,"definitionVersion":3,"starterId":"forged-user","formData":{"reason":"annual leave"}}`)

	handler.SaveStartDraft(context.Background(), c)

	if stub.saveDraftRequest.DefinitionID != 7 || stub.saveDraftRequest.StarterID != "42" {
		t.Fatalf("route definition and authenticated actor must win: %+v", stub.saveDraftRequest)
	}
	if stub.saveDraftRequest.FormData["reason"] != "annual leave" {
		t.Fatalf("unexpected draft form data: %+v", stub.saveDraftRequest.FormData)
	}
}

func TestDeleteStartDraftUsesRouteDefinitionAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "definitionId", Value: "7"})

	handler.DeleteStartDraft(context.Background(), c)

	if stub.deleteDraftID != 7 || stub.deleteDraftActor != "42" {
		t.Fatalf("deleted draft must use route definition and authenticated actor: definition=%d actor=%q", stub.deleteDraftID, stub.deleteDraftActor)
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
			"action": "return",
			"comment": "请补充材料",
			"returnTargetNodeId": "draft",
		"images": [{"id":"uploads/workflow/2026/09/04/reject.png","name":"reject.png","url":"/uploads/workflow/2026/09/04/reject.png","mimeType":"image/png","size":1024}],
		"formData": {"managerComment": "approved"}
	}`)

	handler.CompleteTask(context.Background(), c)

	if stub.completeCalls != 1 {
		t.Fatalf("expected one complete call, got %d", stub.completeCalls)
	}
	if stub.completeRequest.TaskID != "task-1" || stub.completeRequest.ActorID != "42" {
		t.Fatalf("route task and authenticated actor must win: %+v", stub.completeRequest)
	}
	if len(stub.completeRequest.Images) != 1 || stub.completeRequest.Images[0].Name != "reject.png" {
		t.Fatalf("complete task images = %#v", stub.completeRequest.Images)
	}
	if stub.completeRequest.ReturnTargetNodeID != "draft" {
		t.Fatalf("return target node = %q, want draft", stub.completeRequest.ReturnTargetNodeID)
	}
}

func TestMyQueriesIgnoreForgedActorFilters(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Request.SetRequestURI("/api/v2/workflows/instances?starterId=999&scope=copied&definitionCategory=performance&status=running&startTimeFrom=1000&startTimeTo=1999&endTimeFrom=2000&endTimeTo=2999&page=2&pageSize=10")

	handler.ListMyInstances(context.Background(), c)

	if stub.actorID != "42" || stub.instanceQuery.StarterID != "" || stub.instanceQuery.Scope != workflowapp.InstanceScopeCopied {
		t.Fatalf("instance actor must be supplied separately by auth: actor=%q query=%+v", stub.actorID, stub.instanceQuery)
	}
	if stub.instanceQuery.StartTimeFrom != 1000 || stub.instanceQuery.StartTimeTo != 1999 || stub.instanceQuery.EndTimeFrom != 2000 || stub.instanceQuery.EndTimeTo != 2999 {
		t.Fatalf("instance time filters were not parsed: %+v", stub.instanceQuery)
	}
	if stub.instanceQuery.DefinitionCategory != "performance" {
		t.Fatalf("instance definition category filter was not parsed: %+v", stub.instanceQuery)
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

func TestCommentInstanceUsesRouteInstanceAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"instanceId":"forged-instance","actorId":"forged-user","comment":"请补充附件","images":[{"id":"uploads/workflow/2026/09/04/comment.png","name":"comment.png","url":"/uploads/workflow/2026/09/04/comment.png","mimeType":"image/png","size":2048}],"notification":{"userIds":["7","84"],"channels":["in_app","dingtalk_oa"]}}`)

	handler.CommentInstance(context.Background(), c)

	if stub.commentCalls != 1 {
		t.Fatalf("expected one comment call, got %d", stub.commentCalls)
	}
	if stub.commentRequest.InstanceID != "instance-1" || stub.commentRequest.ActorID != "42" || stub.commentRequest.Comment != "请补充附件" {
		t.Fatalf("route instance and authenticated actor must win: %+v", stub.commentRequest)
	}
	if len(stub.commentRequest.Images) != 1 || stub.commentRequest.Images[0].ID != "uploads/workflow/2026/09/04/comment.png" {
		t.Fatalf("comment images = %#v", stub.commentRequest.Images)
	}
	if stub.commentRequest.Notification == nil || len(stub.commentRequest.Notification.UserIDs) != 2 || len(stub.commentRequest.Notification.Channels) != 2 {
		t.Fatalf("comment notification = %#v", stub.commentRequest.Notification)
	}
}

func TestRemindInstanceUsesRouteInstanceAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"instanceId":"forged-instance","actorId":"forged-user","nodeId":"approve"}`)

	handler.RemindInstance(context.Background(), c)

	if stub.remindCalls != 1 {
		t.Fatalf("expected one remind call, got %d", stub.remindCalls)
	}
	if stub.remindRequest.InstanceID != "instance-1" || stub.remindRequest.ActorID != "42" || stub.remindRequest.NodeID != "approve" {
		t.Fatalf("route instance and authenticated actor must win: %+v", stub.remindRequest)
	}
	if body := string(c.Response.Body()); !strings.Contains(body, `"remindedCount":2`) {
		t.Fatalf("reminder response missing result: %s", body)
	}
}

func TestDeleteMyInstanceUsesRouteInstanceAndAuthenticatedUser(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})

	handler.DeleteMyInstance(context.Background(), c)

	if stub.deleteInstanceID != "instance-1" || stub.deleteActorID != "42" {
		t.Fatalf("delete must use route instance and authenticated actor: instance=%q actor=%q", stub.deleteInstanceID, stub.deleteActorID)
	}
}

func newUserContext(userID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("user", &model.User{ID: userID})
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	return c
}
